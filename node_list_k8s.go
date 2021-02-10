package main

import (
	"net"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type apiNodeList struct {
	clientSet *kubernetes.Clientset
}

func newAPINodeList(kubeconfig string) *apiNodeList {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logrus.Panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Panic(err)
	}

	return &apiNodeList{
		clientSet: clientset,
	}
}

func (l *apiNodeList) list() ([]*node, error) {
	nodes, err := l.clientSet.CoreV1().Nodes().List(v1.ListOptions{})
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
