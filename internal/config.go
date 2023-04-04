package internal

import (
	"context"
	"flag"
	"fmt"
	"github.com/sebasmannem/bdtools/pkg/quotagroups"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"time"
)

/*
 * This module reads the config file and returns a config object with all entries from the config yaml file.
 */

const (
	envConfName     = "BDTOOLS_CONFIG"
	defaultConfFile = "/etc/bdtools/config.yaml"
)

var (
	debug      bool
	version    bool
	configFile string
)

func ProcessFlags() (err error) {
	if configFile != "" {
		return
	}

	flag.BoolVar(&debug, "d", false, "Add debugging output")
	flag.BoolVar(&version, "v", false, "Show version information")

	flag.StringVar(&configFile, "c", os.Getenv(envConfName), "Path to configfile")

	flag.Parse()

	if version {
		//nolint
		fmt.Println(appVersion)
		os.Exit(0)
	}

	if configFile == "" {
		configFile = defaultConfFile
	}

	configFile, err = filepath.EvalSymlinks(configFile)
	return err
}

type Config struct {
	QuotaGroups quotagroups.Config `yaml:"quota_groups"`
	LogFile     string             `yaml:"log_file"`
	Debug       bool               `yaml:"debug"`
	Timeout     string             `yaml:"timeout"`
}

func (c *Config) Initialize() {
	c.QuotaGroups.Initialize()
}

func NewConfig() (config Config, err error) {
	if err = ProcessFlags(); err != nil {
		return
	}

	// This only parsed as yaml, nothing else
	// #nosec
	yamlConfig, err := os.ReadFile(configFile)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(yamlConfig, &config)
	config.Initialize()

	if debug {
		config.Debug = true
	}

	return config, err
}

func (c Config) GetTimeoutContext(parentContext context.Context) (context.Context, context.CancelFunc) {
	if c.Timeout == "" {
		return parentContext, nil
	}
	lockDuration, err := time.ParseDuration(c.Timeout)
	if err != nil {
		log.Fatal(err)
	}
	return context.WithTimeout(parentContext, lockDuration)
}
