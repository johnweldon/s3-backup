package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/amz.v3/aws"

	"gopkg.in/johnweldon/s3backup.v0/config"
	"gopkg.in/johnweldon/s3backup.v0/worker"
)

var configFile string
var generateConfig bool

func init() {
	flag.StringVar(&configFile, "config", "backup.config", "config file name")
	flag.BoolVar(&generateConfig, "generate", false, "generate default config file if missing")
}

func main() {
	flag.Parse()

	settings, err := config.Read(configFile)
	if err != nil && generateConfig {
		logf("error opening %q: %s\n", configFile, err)
		settings = generateConf(configFile)
	}

	plan := worker.NewPlan(settings, flag.Args())
	err = plan.Execute()
	if err != nil {
		logf("problem: %s\n", err)
		os.Exit(-1)
	}
}

func generateConf(path string) *config.Settings {
	var err error
	settings := &config.Settings{
		Region: aws.USEast,
		Bucket: "default-s3-backup-bucket",
	}
	settings.Auth, err = aws.EnvAuth()
	if err != nil {
		logf("problem: %s\n", err)
	}
	if err = settings.Write(path); err != nil {
		logf("problem: %s\n", err)
	}
	return settings
}

func logf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}
