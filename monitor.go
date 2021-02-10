package main

import (
	"context"
	"sync"
	"time"

	"github.com/MauveSoftware/cni-health-probe/config"
	"github.com/go-ping/ping"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/label"
)

type nodeList interface {
	list() ([]*node, error)
}

type monitor struct {
	nodes nodeList
	cfg   *config.Config
	metr  *metrics
	mu    sync.Mutex
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
	p, err := ping.NewPinger(n.ip.String())
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
	logrus.Infof("%s: Loss = %v (%v/%v), Dups = %v, Min = %v, Max = %v, Avg = %v, StdDev = %v",
		n.name, s.PacketLoss, s.PacketsRecv, s.PacketsSent, s.PacketsRecvDuplicates, s.MinRtt, s.MaxRtt, s.AvgRtt, s.StdDevRtt)

	ls := []label.KeyValue{
		label.String("destination", n.name),
	}

	ctx := context.Background()
	m.metr.sentPackets.Add(ctx, int64(s.PacketsSent), ls...)
	m.metr.receivedPackets.Add(ctx, int64(s.PacketsRecv), ls...)
	m.metr.lostPackets.Add(ctx, int64(s.PacketsSent-p.PacketsRecv), ls...)
	m.metr.receivedDups.Add(ctx, int64(s.PacketsRecvDuplicates), ls...)
	if s.PacketLoss < 100 {
		for _, r := range s.Rtts {
			m.metr.rtt.Record(ctx, float64(r.Milliseconds()), ls...)
		}
	}
}
