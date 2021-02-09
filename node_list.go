package main

import (
	"net"

	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type nodeList struct {
	clientSet *kubernetes.Clientset
}

func (l *nodeList) list() ([]*node, error) {
	nodes, err := l.clientSet.CoreV1().Nodes().List(v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "could not get node list")
	}

	list := make([]*node, 0)
	for _, n := range nodes.Items {
		ip, _, _ := net.ParseCIDR(n.Spec.PodCIDR)
		list = append(list, &node{
			name: n.Name,
			ip:   ip,
		})
	}

	return list, nil
}
