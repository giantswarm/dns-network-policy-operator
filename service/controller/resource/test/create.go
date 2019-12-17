package test

import (
	"context"
	"fmt"

	"github.com/giantswarm/dns-network-policy-operator/service/key"
	"github.com/giantswarm/microerror"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToDNSNetworkPolicy(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	fmt.Printf("cr %#q with reference networkpolicy %#q", cr.ObjectMeta.Name, cr.Spec.TargetNetworkPolicy)

	return nil
}
