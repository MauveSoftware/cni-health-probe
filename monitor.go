package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/MauveSoftware/cni-health-probe/config"
	"github.com/MauveSoftware/cni-health-probe/metrics"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/sirupsen/logrus"
)

type nodeList interface {
	list() ([]*node, error)
}

type monitor struct {
	nodes nodeList
	cfg   *config.Config
	metr  metrics.Reporter
}

func (m *monitor) start() {
	for {
		m.run()
		time.Sleep(1 * time.Second)
	}
}

func (m *monitor) run() {
	l, err := m.nodes.list()
	if err != nil {
		if errors.Is(err, errHostDrained) {
			logrus.Info("Host is drained. Wait for 60 seconds before next try.")
			time.Sleep(60 * time.Second)
			return
		}

		logrus.Error(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(l))
	for _, n := range l {
		node := n
		go func() {
			m.checkConnectivity(node)
			wg.Done()
		}()
	}
	wg.Wait()
}

func (m *monitor) checkConnectivity(n *node) {
	p, err := probing.NewPinger(n.ip.String())
	if err != nil {
		logrus.Error(err)
	}
	p.Count = int(m.cfg.Ping.Count)
	p.Timeout = m.cfg.Ping.Timeout
	p.Interval = m.cfg.Ping.Interval
	p.SetPrivileged(true)

	err = p.Run()
	if err != nil {
		logrus.Error(err)
	}
	s := p.Statistics()
	logrus.Infof("%s (%s): Loss = %v (%v/%v), Dups = %v, Min = %v, Max = %v, Avg = %v, StdDev = %v",
		n.name, n.ip, s.PacketLoss, s.PacketsRecv, s.PacketsSent, s.PacketsRecvDuplicates, s.MinRtt, s.MaxRtt, s.AvgRtt, s.StdDevRtt)

	ctx := context.Background()
	m.metr.ReportSentPackets(ctx, n.name, int64(s.PacketsSent))
	m.metr.ReportReceivedPackets(ctx, n.name, int64(s.PacketsRecv))
	m.metr.ReportDupPackets(ctx, n.name, int64(s.PacketsRecvDuplicates))
	m.metr.ReportLostPackets(ctx, n.name, int64(s.PacketsSent-s.PacketsRecv))
	if s.PacketLoss < 100 {
		for _, r := range s.Rtts {
			m.metr.ReportRTT(ctx, n.name, float64(r.Milliseconds()))
		}
	}
}
