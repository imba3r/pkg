package scheduler

import (
	"sync"
	"time"
)

// Task interface describes a runnable task.
type Task interface {
	Run() error
	ID() int64
	Name() string
}

// Logger interface describes the kind of logger we'd like to have.
type Logger interface {
	Infof(msgFormat string, args ...interface{})
	Errorf(msgFormat string, args ...interface{})
}

// Job struct contains all relevant data start,
// identify and communicate with a Enabled task.
type Job struct {
	// The task..
	Task

	// Logger..
	Logger

	// Communication with the go routine.
	interval chan time.Duration
	stop     chan struct{}
	runNow   chan struct{}

	// Job meta data guarded by mutex.
	mutex       sync.Mutex
	nextRun     time.Time
	lastRun     time.Time
	curInterval time.Duration
	inProgress  bool
	state       State

	// Waitgroup to start / stop job, semaphore to
	// make sure no tasks run in parallel.
	wg        sync.WaitGroup
	semaphore chan int
}

// NewJob creates a new job for the given task, name and duration.
// It accepts an optional logger, if that is nil the job will be quiet.
func NewJob(task Task, logger Logger, interval time.Duration) *Job {
	job := &Job{
		Task:        task,
		Logger:      logger,
		interval:    make(chan time.Duration, 1),
		stop:        make(chan struct{}),
		runNow:      make(chan struct{}),
		semaphore:   make(chan int, 1),
		nextRun:     time.Time{},
		lastRun:     time.Time{},
		curInterval: interval,
		inProgress:  false,
		state:       Enabled,
	}
	job.wg.Add(1)
	go job.start()
	return job
}

// Stop the job.
func (j *Job) Stop() {
	close(j.stop)
	j.wg.Wait()
}

// Pause the job.
// This just prevents the Run() method of the
// task interface from being executed. The ticker
// will still continue to trigger.
func (j *Job) Pause() {
	j.setState(Disabled)
}

// Resume the job.
func (j *Job) Resume() {
	j.setState(Enabled)
}

// UpdateInterval updates the interval of the job.
// This will create a new ticker and stop the old one.
func (j *Job) UpdateInterval(d time.Duration) {
	j.interval <- d
}

// RunNow triggers the job manually.
func (j *Job) RunNow() {
	j.runNow <- struct{}{}
}

// LastRun returns when the job ran the last time.
func (j *Job) LastRun() time.Time {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	return j.lastRun
}

// NextRun returns when the job runs the next time.
func (j *Job) NextRun() time.Time {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	return j.nextRun
}

// InProgress returns whether the underlying task currently being exectuted.
func (j *Job) InProgress() bool {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	return j.inProgress
}

// State returns whether current state
func (j *Job) State() State {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	return j.state
}

func (j *Job) setState(state State) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	j.state = state
}

// CurrentInterval returns the current interval.
func (j *Job) CurrentInterval() time.Duration {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	return j.curInterval
}

// Service struct wraps a list of jobs and provides some convenience methods.
type Service struct {
	jobs []*Job
}

// NewService constructs a new scheduler service.
func NewService() *Service {
	return &Service{}
}

// AddJob adds the given job to the job list.
func (s *Service) AddJob(j *Job) {
	s.jobs = append(s.jobs, j)
}

// Jobs returns all jobs associated to this scheduler.
func (s *Service) Jobs() []*Job {
	return s.jobs
}

// State represents a jobs state.
type State int

// All available job states.
const (
	Enabled State = iota
	Disabled
)

var (
	stateNames = map[State]string{
		Enabled:  "Enabled",
		Disabled: "Disabled",
	}
)

// String returns the string representation of the given state.
func (s State) String() string {
	return stateNames[s]
}

func (j *Job) run() {
	if j.State() == Disabled {
		j.info("JOB=%s Job is disabled.", j.Name())
		return
	}
	select {
	case j.semaphore <- 1:
		j.wg.Add(1)
		j.info("JOB=%s Starting task.", j.Name())

		// Update the job meta data.
		j.mutex.Lock()
		j.inProgress = true
		j.nextRun = time.Now().UTC().Add(j.curInterval)
		j.lastRun = time.Now().UTC()
		j.mutex.Unlock()

		go func() {
			defer j.wg.Done()

			// Run the task!
			err := j.Run()
			if err != nil {
				j.error("JOB=%s Error during task execution: %v.", j.Name(), err)
			} else {
				j.info("JOB=%s Finished task.", j.Name())
			}

			j.mutex.Lock()
			j.inProgress = false
			for time.Now().UTC().After(j.nextRun) {
				// We may need to update the next run time
				// if this run took longer than 'interval'.
				j.nextRun = j.nextRun.Add(j.curInterval)
			}
			j.mutex.Unlock()

			// Leave the semaphore.
			<-j.semaphore
		}()
	default:
		j.info("JOB=%s Task is still in progress.", j.Name())
	}
}

func (j *Job) start() {
	j.mutex.Lock()
	ticker := time.NewTicker(j.curInterval)
	j.nextRun = time.Now().UTC().Add(j.curInterval)
	j.mutex.Unlock()

	j.info("JOB=%s Initialized... first run will be at %s.", j.Name(), j.NextRun().Format(time.RFC3339))
	for {
		select {
		case <-ticker.C:
			j.info("JOB=%s Received timer trigger.", j.Name())
			j.run()
		case <-j.runNow:
			j.info("JOB=%s Received manual trigger.", j.Name())
			j.run()
		case interval := <-j.interval:
			j.info("JOB=%s Updating interval to %f minutes.", j.Name(), interval.Minutes())

			// We have to stop the old ticker and create
			// a new one in order to change the job interval.
			ticker.Stop()
			ticker = time.NewTicker(interval)

			j.mutex.Lock()
			j.nextRun = time.Now().UTC().Add(interval)
			j.curInterval = interval
			j.mutex.Unlock()
		case <-j.stop:
			j.info("JOB=%s Stopping job.", j.Name())

			// Stop the ticker and mark the main job go routine as done so
			// that the blocking wait in the Stop() function can continue.
			ticker.Stop()
			j.wg.Done()
			return
		}
	}
}

func (j *Job) info(msgFormat string, args ...interface{}) {
	if j.Logger != nil {
		j.Logger.Infof(msgFormat, args...)
	}
}

func (j *Job) error(msgFormat string, args ...interface{}) {
	if j.Logger != nil {
		j.Logger.Errorf(msgFormat, args...)
	}
}
