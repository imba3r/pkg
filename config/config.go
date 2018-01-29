package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "could not read config file")
	}

	var fromDisk interface{}
	err = s.LoadFromDisk(&fromDisk)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// LoadFromMemory marshals the current config into the given struct.
func (s *Service) LoadFromMemory(dest interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := json.Unmarshal(s.config, &dest)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal config file")
	}

	return nil
}

// LoadFromDisk marshals the config from disk into the given struct.
func (s *Service) LoadFromDisk(dest interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	bytes, err := ioutil.ReadFile(s.path)
	if err != nil {
		return errors.Wrap(err, "could not read file")
	}

	err = json.Unmarshal(bytes, &dest)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal config file")
	}

	s.config = bytes
	return nil
}

// Save saves the given configuration to disk.
func (s *Service) Save(config interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	bytes, err := json.Marshal(&config)
	if err != nil {
		return errors.Wrap(err, "could not marshal config file")
	}

	err = ioutil.WriteFile(s.path, bytes, 0644)
	if err != nil {
		return errors.Wrap(err, "could write config file to disk")
	}

	s.config = bytes
	return nil
}
