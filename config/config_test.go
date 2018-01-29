package config_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/imba3r/pkg/assert"
	"github.com/imba3r/pkg/config"
)

type configStruct struct {
	SomeString string
	SomeFlag   bool
	SomeNumber int
	SomeSlice  []string
}

var testConfig = configStruct{"string", true, 42, []string{"Value 1", "Value 2"}}

func TestNewService_ExistingConfig(t *testing.T) {
	dir, err := ioutil.TempDir("", "grabber-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, ".grabber.cfg")

	s, err := config.NewService(path, testConfig)
	assert.NoError(t, err)
	assert.True(t, s != nil, "service should not be nil")

	var fromDisk configStruct
	err = s.LoadFromDisk(&fromDisk)
	assert.NoError(t, err)
	assert.Equals(t, testConfig, fromDisk)

	var fromMemory configStruct
	err = s.LoadFromDisk(&fromMemory)
	assert.NoError(t, err)
	assert.Equals(t, testConfig, fromMemory)
}

func TestLoad_Gibberish(t *testing.T) {
	dir, err := ioutil.TempDir("", "grabber-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, ".grabber.cfg")
	file, err := os.Create(path)
	assert.NoError(t, err)
	defer file.Close()

	fmt.Fprintf(file, "definitely no json")
	file.Sync()

	_, err = config.NewService(path, testConfig)
	assert.Error(t, err)
}
