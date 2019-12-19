package key

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/example/v1alpha1"
	"github.com/giantswarm/microerror"
)

const (
	AnnotationManagedBy                  = "giantswarm.io/managed-by"
	AnnotationNetworkPolicyRole          = "giantswarm.io/network-policy-role"
	AnnotationNetworkPolicyRoleActive    = "active"
	AnnotationNetworkPolicyRoleSupressed = "supressed"
	PolicyTypeEgress                     = "Egress"
	PodSelectorMatchLabelRandom          = "random-match-label"
	ProjectName                          = "dns-network-policy-operator"
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

func RandomLabel() string {
	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	maxLength := 63

	b := make([]rune, maxLength)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func EffectiveNetworkPolicyName(targetPolicyName string) string {
	return fmt.Sprintf("%s-%s", targetPolicyName, AnnotationNetworkPolicyRoleActive)

}
