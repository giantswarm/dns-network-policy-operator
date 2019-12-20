package controller

import (
	"github.com/giantswarm/apiextensions/pkg/apis/example/v1alpha1"
	"github.com/giantswarm/k8sclient"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/dns-network-policy-operator/pkg/project"
	"github.com/giantswarm/dns-network-policy-operator/service/controller/resource/dnsnetworkpolicy"
)

type DNSNetworkPolicyConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	// resolver settings
	Resolver dnsnetworkpolicy.Resolver
}

type DNSNetworkPolicy struct {
	*controller.Controller
	Resolver dnsnetworkpolicy.Resolver
}

func NewDNSNetworkPolicy(config DNSNetworkPolicyConfig) (*DNSNetworkPolicy, error) {
	var err error

	resourceSets, err := newDNSNetworkPolicyResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		c := controller.Config{
			CRD:          v1alpha1.NewDNSNetworkPolicyCRD(),
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			ResourceSets: resourceSets,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.DNSNetworkPolicy)
			},

			Name: project.Name() + "-dns-network-policy-controller",
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &DNSNetworkPolicy{
		Controller: operatorkitController,
		Resolver:   config.Resolver,
	}

	return c, nil
}

func newDNSNetworkPolicyResourceSets(config DNSNetworkPolicyConfig) ([]*controller.ResourceSet, error) {
	var err error

	var resourceSet *controller.ResourceSet
	{
		c := DNSNetworkPolicyResourceSetConfig{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		resourceSet, err = newDNSNetworkPolicyResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSet,
	}

	return resourceSets, nil
}
