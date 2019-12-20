package dnsnetworkpolicy

import (
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"k8s.io/client-go/kubernetes"
)

const (
	Name = "dns-network-policy"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	Resolver Resolver
}

type Resource struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	resolver Resolver
}

type Resolver struct {
	RoundRobinAttempts int
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient.K8sClient(),
		logger:    config.Logger,
		resolver:  config.Resolver,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
