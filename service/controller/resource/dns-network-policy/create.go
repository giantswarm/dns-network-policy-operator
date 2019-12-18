package test

import (
	"context"

	"github.com/giantswarm/microerror"
     metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/dns-network-policy-operator/service/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
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

	// Add annotation to the target network policy
    if targetNetworkPolicy.ObjectMeta.Annotations == nil {
		targetNetworkPolicy.ObjectMeta.Annotations = make(map[string]string)
    }
	targetNetworkPolicy.ObjectMeta.Annotations[key.AnnotationManagedBy] = key.ProjectName

	// Verify there is Egress in PolicyTypes
	var policyTypeEgressExists bool
	{
	  for _, policyType := range targetNetworkPolicy.Spec.PolicyTypes {
		if policyType == key.PolicyTypeEgress {
			policyTypeEgressExists = true
		}
      }
    }
	if !policyTypeEgressExists {
		r.logger.LogCtx(ctx, "level", "debug", "message", "updating target network policy with PolicyType Egress")
		targetNetworkPolicy.ObjectMeta.Annotations[key.AnnotationRestoreEgress] = "true"
		targetNetworkPolicy.Spec.PolicyTypes = append(targetNetworkPolicy.Spec.PolicyTypes, key.PolicyTypeEgress)
	}

	// Resolve domains into IP addresses

	_, err = r.k8sClient.NetworkingV1().NetworkPolicies(ns).Update(targetNetworkPolicy)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
