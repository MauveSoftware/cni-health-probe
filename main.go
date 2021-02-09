package main

import (
	"github.com/MauveSoftware/cni-health-probe/config"
	"github.com/sirupsen/logrus"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	configFile := kingpin.Flag("config", "Path to config file").Default("config.yml").String()
	metricsAddress := kingpin.Flag("metrics-address", "Listen address for metrics").Default(":9999").String()
	kingpin.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		logrus.Panic(err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfigPath)
	if err != nil {
		logrus.Panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Panic(err)
	}

	prom, err := newMetrics(*metricsAddress)
	if err != nil {
		logrus.Panic(err)
	}

	nodes := &nodeList{clientSet: clientset}
	monitor := &monitor{
		metr:  prom,
		nodes: nodes,
		cfg:   cfg,
	}
	go monitor.start()

	logrus.Fatal(prom.listen())
}
