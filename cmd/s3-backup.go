package main

import (
	"fmt"
	"os"

	"gopkg.in/amz.v3/aws"
	"gopkg.in/amz.v3/s3"

	"gopkg.in/johnweldon/s3backup.v0/config"
)

func main() {

	settings, err := config.Read("backup.config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "problem: %s\n", err)
		settings = generateConf("backup.config")
	}

	server := s3.New(settings.Auth, settings.Region)

	bucket, err := server.Bucket(settings.Bucket)
	if err != nil {
		fmt.Fprintf(os.Stderr, "problem: %s\n", err)
		return
	}

	resp, err := bucket.List("", "/", "", 100)
	if err != nil {
		if se, ok := err.(*s3.Error); ok {
			switch se.Code {
			case "NoSuchBucket":
				fmt.Fprintf(os.Stderr, "no such bucket %q, creating...\n", settings.Bucket)
				if err = bucket.PutBucket(s3.BucketOwnerFull); err != nil {
					fmt.Fprintf(os.Stderr, "failed to create bucket: %#v\n", err)
					return
				}
				if resp, err = bucket.List("", "/", "", 100); err != nil {
					fmt.Fprintf(os.Stderr, "problem: %#v\n", err)
					return
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "problem: %q\n", err)
			return
		}
	}

	fmt.Fprintf(os.Stdout, "list: %+v\n", resp)
}

func generateConf(path string) *config.Settings {
	var err error
	settings := &config.Settings{
		Region: aws.USEast,
		Bucket: "default-s3-backup-bucket",
	}
	settings.Auth, err = aws.EnvAuth()
	if err != nil {
		fmt.Fprintf(os.Stderr, "problem: %s\n", err)
	}
	if err = settings.Write(path); err != nil {
		fmt.Fprintf(os.Stderr, "problem: %s\n", err)
	}
	return settings
}
