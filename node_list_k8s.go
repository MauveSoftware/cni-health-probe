package main

import (
	"net"

	"github.com/MauveSoftware/cni-health-probe/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type apiNodeList struct {
	cfg       *config.Config
	clientSet *kubernetes.Clientset
}

func newAPINodeList(cfg *config.Config) *apiNodeList {
	config, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfigPath)
	if err != nil {
		logrus.Panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Panic(err)
	}

	return &apiNodeList{
		cfg:       cfg,
		clientSet: clientset,
	}
}

func (l *apiNodeList) list() ([]*node, error) {
	selector := v1.LabelSelector{MatchLabels: l.cfg.NodeSelector}
	nodes, err := l.clientSet.CoreV1().Nodes().List(v1.ListOptions{LabelSelector: labels.Set(selector.MatchLabels).String()})
	if err != nil {
		return nil, errors.Wrap(err, "could not get node list")
	}

	list := make([]*node, 0)
	for _, n := range nodes.Items {
		v, exists := n.Labels["mauve.cloud/cni-tunnel-ip"]
		if !exists {
			continue
		}

		ip := net.ParseIP(v)
		if ip == nil {
			continue
		}

		list = append(list, &node{
			name: n.Name,
			ip:   ip,
		})
	}

	return list, nil
}
