package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/example/v1alpha1"
	"github.com/giantswarm/microerror"
)

func ToDNSNetworkPolicy(v interface{}) (v1alpha1.DNSNetworkPolicy, error) {
	if v == nil {
		return v1alpha1.DNSNetworkPolicy{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.DNSNetworkPolicy{}, v)
	}

	p, ok := v.(*v1alpha1.DNSNetworkPolicy)
	if !ok {
		return v1alpha1.DNSNetworkPolicy{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.DNSNetworkPolicy{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}
