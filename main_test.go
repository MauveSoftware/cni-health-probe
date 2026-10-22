package main

import (
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name                   string
		env                    map[string]string
		args                   []string
		expectedConfigFile     string
		expectedMetricsAddress string
	}{
		{
			name:                   "defaults",
			expectedConfigFile:     "config.yml",
			expectedMetricsAddress: ":9999",
		},
		{
			name: "from env",
			env: map[string]string{
				"CONFIG_FILE":     "/config/config.yml",
				"METRICS_ADDRESS": ":8080",
			},
			expectedConfigFile:     "/config/config.yml",
			expectedMetricsAddress: ":8080",
		},
		{
			name: "flags override env",
			env: map[string]string{
				"CONFIG_FILE":     "/config/config.yml",
				"METRICS_ADDRESS": ":8080",
			},
			args:                   []string{"--config", "other.yml", "--metrics-address", ":7070"},
			expectedConfigFile:     "other.yml",
			expectedMetricsAddress: ":7070",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("CONFIG_FILE", "")
			t.Setenv("METRICS_ADDRESS", "")
			for k, v := range test.env {
				t.Setenv(k, v)
			}

			a, err := parseArgs(test.args)
			if err != nil {
				t.Fatal(err)
			}

			if a.configFile != test.expectedConfigFile {
				t.Errorf("config file: expected %q, got %q", test.expectedConfigFile, a.configFile)
			}

			if a.metricsAddress != test.expectedMetricsAddress {
				t.Errorf("metrics address: expected %q, got %q", test.expectedMetricsAddress, a.metricsAddress)
			}
		})
	}
}
