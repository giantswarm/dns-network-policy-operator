package service

import (
	"github.com/giantswarm/dns-network-policy-operator/flag/service/resolver"
	"github.com/giantswarm/operatorkit/flag/service/kubernetes"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	Kubernetes kubernetes.Kubernetes
	Resolver   resolver.Resolver
}
