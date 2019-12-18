package test

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

	// Restore egress policy
	if targetNetworkPolicy.ObjectMeta.Annotations == nil {
		return nil
	}
	if _, ok := targetNetworkPolicy.ObjectMeta.Annotations[key.AnnotationRestoreEgress]; ok {
		if targetNetworkPolicy.ObjectMeta.Annotations[key.AnnotationRestoreEgress] == "true" {
			delete(targetNetworkPolicy.ObjectMeta.Annotations, key.AnnotationRestoreEgress)
			targetNetworkPolicy.Spec.PolicyTypes = remove(targetNetworkPolicy.Spec.PolicyTypes, key.PolicyTypeEgress)
		}
	}

	// Delete annotation from the target network policy
	delete(targetNetworkPolicy.ObjectMeta.Annotations, key.AnnotationManagedBy)

	// Update target network policy
	_, err = r.k8sClient.NetworkingV1().NetworkPolicies(ns).Update(targetNetworkPolicy)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func remove(s []networkingV1.PolicyType, r networkingV1.PolicyType) []networkingV1.PolicyType {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
