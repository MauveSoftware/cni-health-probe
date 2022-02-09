package config

import (
	"testing"
	"time"

	"gopkg.in/go-playground/assert.v1"
)

func TestLoad(t *testing.T) {
	cfg, err := Load("test-files/config.yml")
	if err != nil {
		t.Fatal(err)
	}

	expected := &Config{
		Ping: PingConfig{
			Count:    10,
			Timeout:  1 * time.Second,
			Interval: 100 * time.Millisecond,
		},
		KubeConfigPath: "~/kubeconfig",
		NodeSelector: map[string]string{
			"mauve.cloud/test": "foo",
		},
		PodSelector: map[string]string{
			"app": "my-app",
		},
		Namespace: "default",
	}
	assert.Equal(t, expected, cfg)
}
