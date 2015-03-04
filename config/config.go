package config

import (
	"encoding/json"
	"io/ioutil"

	"gopkg.in/amz.v3/aws"
)

type Settings struct {
	aws.Auth   `json:"auth,omitempty"`
	aws.Region `json:"region,omitempty"`
	Bucket     string `json:"bucket,omitempty"`
}

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
