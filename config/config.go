package config

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Config represents the configuration for the health check daemon
type Config struct {
	Ping           PingConfig        `yaml:"ping"`
	KubeConfigPath string            `yaml:"kubeconfig"`
	NodeSelector   map[string]string `yaml:"nodeSelector"`
	PodSelector    map[string]string `yaml:"podSelector"`
	Namespace      string            `yaml:"namespace"`
}

// PingConfig defines the parameter for the ping health check
type PingConfig struct {
	Count    uint32        `yaml:"count"`
	Timeout  time.Duration `yaml:"timeout"`
	Interval time.Duration `yaml:"interval"`
}

// New creates a new config with default values
func New() *Config {
	return &Config{
		Ping: PingConfig{
			Count:    100,
			Timeout:  10 * time.Second,
			Interval: 1 * time.Second,
		},
		NodeSelector: make(map[string]string),
		PodSelector: map[string]string{
			"app": "cni-health-probe",
		},
		Namespace: "kube-system",
	}
}

// Load reads and loads an config file
func Load(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "could not open config file")
	}

	return Parse(b)
}

// Parse parses the yaml of a config file
func Parse(b []byte) (*Config, error) {
	cfg := New()
	err := yaml.NewDecoder(bytes.NewReader(b)).Decode(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse config")
	}

	return cfg, nil
}
