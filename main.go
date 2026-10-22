package main

import (
	"os"

	"github.com/MauveSoftware/cni-health-probe/config"
	"github.com/MauveSoftware/cni-health-probe/metrics"
	"github.com/sirupsen/logrus"

	kingpin "github.com/alecthomas/kingpin/v2"
)

type cliArgs struct {
	configFile     string
	metricsAddress string
}

func parseArgs(args []string) (*cliArgs, error) {
	a := &cliArgs{}
	app := kingpin.New("cni-health-probe", "Probe to check network connectivity in kubernetes CNI")
	app.Flag("config", "Path to config file").Envar("CONFIG_FILE").Default("config.yml").StringVar(&a.configFile)
	app.Flag("metrics-address", "Listen address for metrics").Envar("METRICS_ADDRESS").Default(":9999").StringVar(&a.metricsAddress)

	_, err := app.Parse(args)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func main() {
	args, err := parseArgs(os.Args[1:])
	if err != nil {
		logrus.Fatal(err)
	}

	cfg, err := config.Load(args.configFile)
	if err != nil {
		logrus.Panic(err)
	}

	prom, err := metrics.NewServer(args.metricsAddress)
	if err != nil {
		logrus.Panic(err)
	}

	nodes := newAPINodeList(cfg)
	monitor := &monitor{
		metr:  prom,
		nodes: nodes,
		cfg:   cfg,
	}
	go monitor.start()

	logrus.Fatal(prom.Listen())
}
