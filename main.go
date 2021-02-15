package main

import (
	"github.com/MauveSoftware/cni-health-probe/config"
	"github.com/MauveSoftware/cni-health-probe/metrics"
	"github.com/sirupsen/logrus"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	configFile := kingpin.Flag("config", "Path to config file").Default("config.yml").String()
	metricsAddress := kingpin.Flag("metrics-address", "Listen address for metrics").Default(":9999").String()
	kingpin.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		logrus.Panic(err)
	}

	prom, err := metrics.NewServer(*metricsAddress)
	if err != nil {
		logrus.Panic(err)
	}

	nodes := newAPINodeList(cfg.KubeConfigPath)
	monitor := &monitor{
		metr:  prom,
		nodes: nodes,
		cfg:   cfg,
	}
	go monitor.start()

	logrus.Fatal(prom.Listen())
}
