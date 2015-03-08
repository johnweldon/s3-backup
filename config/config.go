package config

import (
	"encoding/json"
	"io/ioutil"

	"gopkg.in/amz.v3/aws"
)

// Settings describes user specified options.
type Settings struct {
	aws.Auth   `json:"auth,omitempty"`
	aws.Region `json:"region,omitempty"`
	Bucket     string `json:"bucket,omitempty"`
}

// Read attempts to read Settings from a file, and
// returns an error if it fails.
func Read(path string) (*Settings, error) {
	var s Settings
	var buf []byte
	var err error
	if buf, err = ioutil.ReadFile(path); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(buf, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Write serializes the Settings to a file in JSON format.
func (s *Settings) Write(path string) error {
	var buf []byte
	var err error
	if buf, err = json.Marshal(s); err != nil {
		return err
	}
	if err = ioutil.WriteFile(path, buf, 0644); err != nil {
		return err
	}
	return nil
}
