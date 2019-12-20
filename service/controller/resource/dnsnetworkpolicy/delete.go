package dnsnetworkpolicy

import (
	"context"

	"github.com/giantswarm/dns-network-policy-operator/service/key"
	"github.com/giantswarm/microerror"
	networkingV1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToDNSNetworkPolicy(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	ns := cr.ObjectMeta.Namespace
	targetNetworkPolicyName := cr.Spec.TargetNetworkPolicy
	targetNetworkPolicy, err := r.k8sClient.NetworkingV1().NetworkPolicies(ns).Get(targetNetworkPolicyName, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	// delete random matcher

	// Delete annotations from the target network policy
	delete(targetNetworkPolicy.ObjectMeta.Annotations, key.AnnotationManagedBy)
	delete(targetNetworkPolicy.ObjectMeta.Annotations, key.AnnotationNetworkPolicyRole)

	// Delete random match label
	delete(targetNetworkPolicy.Spec.PodSelector.MatchLabels, key.PodSelectorMatchLabelRandom)

	// Update target network policy
	_, err = r.k8sClient.NetworkingV1().NetworkPolicies(ns).Update(targetNetworkPolicy)
	if err != nil {
		return microerror.Mask(err)
	}

	// Delete effectve network policy
	effectiveNetworkPolicyName := key.EffectiveNetworkPolicyName(targetNetworkPolicy.Name)
	err = r.k8sClient.NetworkingV1().NetworkPolicies(ns).Delete(effectiveNetworkPolicyName, &metav1.DeleteOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func removePolicyType(s []networkingV1.PolicyType, r networkingV1.PolicyType) []networkingV1.PolicyType {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
