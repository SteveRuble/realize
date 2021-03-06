package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// settings const
const (
	permission = 0775
	directory  = ".realize"
	file       = "realize.yaml"
	fileOut    = "outputs.log"
	fileErr    = "errors.log"
	fileLog    = "logs.log"
)

// random string preference
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// Settings defines a group of general settings and options
type Settings struct {
	file      string
	Files     `yaml:"files,omitempty" json:"files,omitempty"`
	Legacy    Legacy `yaml:"legacy" json:"legacy"`
	FileLimit int32  `yaml:"flimit,omitempty" json:"flimit,omitempty"`
	Recovery  bool   `yaml:"recovery,omitempty" json:"recovery,omitempty"`
}

// Legacy is used to force polling and set a custom interval
type Legacy struct {
	Force    bool          `yaml:"force" json:"force"`
	Interval time.Duration `yaml:"interval" json:"interval"`
}

// Files defines the files generated by realize
type Files struct {
	Clean   bool     `yaml:"clean,omitempty" json:"clean,omitempty"`
	Outputs Resource `yaml:"outputs,omitempty" json:"outputs,omitempty"`
	Logs    Resource `yaml:"logs,omitempty" json:"log,omitempty"`
	Errors  Resource `yaml:"errors,omitempty" json:"error,omitempty"`
}

// Resource status and file name
type Resource struct {
	Status bool
	Name   string
}

// Rand is used for generate a random string
func random(n int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

// Delete realize folder
func (s *Settings) del(d string) error {
	_, err := os.Stat(d)
	if !os.IsNotExist(err) {
		return os.RemoveAll(d)
	}
	return err
}

// Validate checks a fatal error
func (s Settings) validate(err error) error {
	if err != nil {
		s.fatal(err, "")
	}
	return nil
}

// Read from config file
func (s *Settings) read(out interface{}) error {
	localConfigPath := s.file
	// backward compatibility
	path := filepath.Join(directory, s.file)
	if _, err := os.Stat(path); err == nil {
		localConfigPath = path
	}
	content, err := s.stream(localConfigPath)
	if err == nil {
		err = yaml.Unmarshal(content, out)
		return err
	}
	return err
}

// Record create and unmarshal the yaml config file
func (s *Settings) record(out interface{}) error {
	y, err := yaml.Marshal(out)
	if err != nil {
		return err
	}
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err = os.Mkdir(directory, permission); err != nil {
			return s.write(s.file, y)
		}
	}
	return s.write(filepath.Join(directory, s.file), y)
}

// Stream return a byte stream of a given file
func (s Settings) stream(file string) ([]byte, error) {
	_, err := os.Stat(file)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadFile(file)
	s.validate(err)
	return content, err
}

// Fatal prints a fatal error with its additional messages
func (s Settings) fatal(err error, msg ...interface{}) {
	if len(msg) > 0 && err != nil {
		log.Fatalln(red.regular(msg...), err.Error())
	} else if err != nil {
		log.Fatalln(err.Error())
	}
}

// Write a file
func (s Settings) write(name string, data []byte) error {
	err := ioutil.WriteFile(name, data, permission)
	return s.validate(err)
}

// Create a new file and return its pointer
func (s Settings) create(path string, name string) *os.File {
	var file string
	if _, err := os.Stat(directory); err == nil {
		file = filepath.Join(path, directory, name)
	} else {
		file = filepath.Join(path, name)
	}
	out, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, permission)
	s.validate(err)
	return out
}
