package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

// Service struct gives access to the config file.
type Service struct {
	path string

	mutex  sync.Mutex
	config []byte
}

// NewService constructs a new config service.
func NewService(cfgPath string, defaultConfig interface{}) (*Service, error) {
	s := &Service{path: cfgPath}

	// Create a default config if it does not exist yet.
	_, err := os.Stat(cfgPath)
	if os.IsNotExist(err) {
		s.Save(defaultConfig)
	} else if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	err = s.LoadFromDisk(defaultConfig)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// LoadFromMemory marshals the current config into the given struct.
func (s *Service) LoadFromMemory(dest interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := json.Unmarshal(s.config, dest)
	if err != nil {
		return fmt.Errorf("could not unmarshal config file: %v", err)
	}

	return nil
}

// LoadFromDisk marshals the config from disk into the given struct.
func (s *Service) LoadFromDisk(dest interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	bytes, err := ioutil.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("could not read file: %v", err)
	}

	err = json.Unmarshal(bytes, dest)
	if err != nil {
		return fmt.Errorf("could not unmarshal config file: %v", err)
	}

	// Marshal the the config back to make sure the version in-memory
	// contains everything - the file on disk doesn't need to.
	s.config, err = json.Marshal(dest)
	if err != nil {
		return fmt.Errorf("could marshal config file: %v", err)
	}
	return nil
}

// Save saves the given configuration to disk.
func (s *Service) Save(config interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	bytes, err := json.Marshal(&config)
	if err != nil {
		return fmt.Errorf("could marshal config file: %v", err)
	}

	err = ioutil.WriteFile(s.path, bytes, 0644)
	if err != nil {
		return fmt.Errorf("could write config file to disk: %v", err)
	}

	s.config = bytes
	return nil
}

func (s *Service) Handler() {
	h := http.NewServeMux()

	h.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, you hit foo!")
	})

	h.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, you hit bar!")
	})

	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprintln(w, "You're lost, go home")
	})

}
