package scheduler_test

import (
	"sync"
	"testing"
	"time"

	"github.com/imba3r/grabber/core/scheduler"
)

type testTask struct {
	countMutex sync.Mutex
	count      int
}

func (task *testTask) getCount() int {
	task.countMutex.Lock()
	defer task.countMutex.Unlock()
	return task.count
}

func (task *testTask) Run() error {
	task.countMutex.Lock()
	defer task.countMutex.Unlock()
	task.count = task.count + 1
	return nil
}

func (task *testTask) Name() string {
	return "testTask"
}

func (task *testTask) ID() int64 {
	return 0
}

func TestScheduler_Scheduling(t *testing.T) {
	task := &testTask{}
	job := scheduler.NewJob(task, time.Second*60)

	time.Sleep(time.Millisecond * 50)

	nextRun := job.NextRun()
	lastRun := job.LastRun()

	if nextRun == (time.Time{}) {
		t.Error("Job should have a scheduled time by now..")
	}
	if lastRun != (time.Time{}) {
		t.Error("Job should not have run yet..")
	}

	job.UpdateInterval(time.Second * 30)
	time.Sleep(time.Millisecond * 50)

	newNextRun := job.NextRun()

	if nextRun == newNextRun {
		t.Error("A new run time should have been scheduled..")
	}
	job.Stop()
}

func TestScheduler_PauseResume(t *testing.T) {
	task := &testTask{}

	// Start the job, interval of 10 ms, run 10 times.
	job := scheduler.NewJob(task, time.Millisecond*10)
	time.Sleep(time.Millisecond * 105)

	// Pause for 100 ms.
	job.Pause()
	time.Sleep(time.Millisecond * 100)

	// Change interval to 20 ms and resume, adds another 5 runs.
	job.UpdateInterval(time.Millisecond * 20)
	job.Resume()
	time.Sleep(time.Millisecond * 115)

	// Job should have run 15 times now.
	job.Stop()
	if task.getCount() != 15 {
		t.Error("Expected 15 runs, got", task.getCount())
	}
}

type longRunningTask struct {
	countMutex sync.Mutex
	count      int
}

func (task *longRunningTask) getCount() int {
	task.countMutex.Lock()
	defer task.countMutex.Unlock()
	return task.count
}

func (task *longRunningTask) Run() error {
	task.countMutex.Lock()
	defer task.countMutex.Unlock()
	task.count = task.count + 1
	time.Sleep(1 * time.Second)
	return nil
}

func (task *longRunningTask) Name() string {
	return "longRunningTask"
}

func (task *longRunningTask) ID() int64 {
	return 0
}

func TestScheduler_ControlSpam(t *testing.T) {
	task := &longRunningTask{}

	job := scheduler.NewJob(task, time.Millisecond*10)
	time.Sleep(time.Millisecond * 15)

	if job.InProgress() == false {
		t.Error("Job should be in progress!")
	}

	// The job is running for one second, meanwhile we spam pause & resume.
	for i := 1; i <= 100; i++ {
		if i%2 == 1 {
			if job.State() != scheduler.Enabled {
				t.Error("Job should be enabled!")
			}
			job.Pause()
		} else {
			if job.State() != scheduler.Disabled {
				t.Error("Job should be disabled!")
			}
			job.Resume()
		}
	}
	job.Stop()
}

func TestScheduler_RunNow(t *testing.T) {
	task := &longRunningTask{}

	// Start the job which will not run for one hour, hence we trigger it manually.
	job := scheduler.NewJob(task, time.Hour*1)
	time.Sleep(time.Millisecond * 15)

	if job.InProgress() != false {
		t.Error("Job should not be in progress!")
	}

	job.RunNow()
	time.Sleep(time.Millisecond * 15)

	if job.InProgress() == false {
		t.Error("Job should run now!")
	}

	job.Stop()
}
