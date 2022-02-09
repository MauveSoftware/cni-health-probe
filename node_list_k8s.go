package main

import (
	"net"
	"os"

	"github.com/MauveSoftware/cni-health-probe/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type apiNodeList struct {
	cfg        *config.Config
	clientSet  *kubernetes.Clientset
	myNodeName string
}

func newAPINodeList(cfg *config.Config) *apiNodeList {
	config, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfigPath)
	if err != nil {
		logrus.Panic(err)
	}

	nodeName, _ := os.LookupEnv("NODE_NAME")

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Panic(err)
	}

	return &apiNodeList{
		cfg:        cfg,
		clientSet:  clientset,
		myNodeName: nodeName,
	}
}

func (l *apiNodeList) list() ([]*node, error) {
	selector := v1.LabelSelector{MatchLabels: l.cfg.NodeSelector}
	nodes, err := l.clientSet.CoreV1().Nodes().List(v1.ListOptions{LabelSelector: labels.Set(selector.MatchLabels).String()})
	if err != nil {
		return nil, errors.Wrap(err, "could not get node list")
	}

	containsMyHost := false
	list := make([]*node, 0)
	for _, n := range nodes.Items {
		if n.Name == l.myNodeName {
			containsMyHost = true
			continue
		}

		if len(n.Status.Addresses) == 0 {
			continue
		}

		ip, err := l.podIPForNode(n)
		if err != nil {
			return nil, errors.Wrap(err, "could not get pod IP for node "+n.Name)
		}

		if ip == nil {
			continue
		}

		list = append(list, &node{
			name: n.Name,
			ip:   ip,
		})
	}

	if !containsMyHost {
		return nil, hostDrainedErr
	}

	return list, nil
}

func (l *apiNodeList) podIPForNode(n corev1.Node) (net.IP, error) {
	selector := v1.LabelSelector{MatchLabels: l.cfg.PodSelector}
	pods, err := l.clientSet.CoreV1().Pods(l.cfg.Namespace).List(v1.ListOptions{
		LabelSelector: labels.Set(selector.MatchLabels).String(),
	})
	if err != nil {
		return nil, err
	}

	for _, p := range pods.Items {
		if p.Status.HostIP == n.Status.Addresses[0].Address && p.Status.Phase == corev1.PodRunning {
			return net.ParseIP(p.Status.PodIP), nil
		}
	}

	return nil, err
}
