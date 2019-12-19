package test

import (
	"context"
	"fmt"
	"net"

	"github.com/giantswarm/microerror"
	networkingV1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	targetNetworkPolicy.ObjectMeta.Annotations[key.AnnotationNetworkPolicyRole] = key.AnnotationNetworkPolicyRoleSupressed

	//create new effective policy if it doesn't exist
	effectiveNetworkPolicyName := key.EffectiveNetworkPolicyName(targetNetworkPolicy.Name)
	effectiveNetworkPolicy, err := r.k8sClient.NetworkingV1().NetworkPolicies(ns).Get(effectiveNetworkPolicyName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		effectiveNetworkPolicy := targetNetworkPolicy.DeepCopy()
		effectiveNetworkPolicy.ObjectMeta.ResourceVersion = ""
		effectiveNetworkPolicy.Name = effectiveNetworkPolicyName

		_, err = r.k8sClient.NetworkingV1().NetworkPolicies(ns).Create(effectiveNetworkPolicy)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if err != nil {
		return microerror.Mask(err)
	}

	// Add annotation to the target network policy
	if effectiveNetworkPolicy.ObjectMeta.Annotations == nil {
		effectiveNetworkPolicy.ObjectMeta.Annotations = make(map[string]string)
	}
	effectiveNetworkPolicy.ObjectMeta.Annotations[key.AnnotationManagedBy] = key.ProjectName
	effectiveNetworkPolicy.ObjectMeta.Annotations[key.AnnotationNetworkPolicyRole] = key.AnnotationNetworkPolicyRoleActive

	// Verify there is Egress in PolicyTypes
	var policyTypeEgressExists bool
	{
		for _, policyType := range effectiveNetworkPolicy.Spec.PolicyTypes {
			if policyType == key.PolicyTypeEgress {
				policyTypeEgressExists = true
			}
		}
	}
	if !policyTypeEgressExists {
		effectiveNetworkPolicy.Spec.PolicyTypes = append(effectiveNetworkPolicy.Spec.PolicyTypes, key.PolicyTypeEgress)
	}

	// Resolve domains into IP addresses
	var desiredIPs []net.IP
	{
		for _, domain := range cr.Spec.Domains {
			currentIPs, err := net.LookupIP(domain)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", err.Error())
				continue
			}
			desiredIPs = append(desiredIPs, currentIPs...)
		}
	}

	var desiredNetworkPolicyPeers []networkingV1.NetworkPolicyPeer
	{
		for _, ip := range desiredIPs {
			var ipType string
			if ip.To4() != nil {
				ipType = "32"
			} else {
				ipType = "128"
			}
			desiredNetworkPolicyPeer := networkingV1.NetworkPolicyPeer{
				IPBlock: &networkingV1.IPBlock{
					CIDR: fmt.Sprintf("%s/%s", ip.String(), ipType),
				},
			}
			desiredNetworkPolicyPeers = append(desiredNetworkPolicyPeers, desiredNetworkPolicyPeer)
		}
	}
	dnsPolicyEgressRule := networkingV1.NetworkPolicyEgressRule{
		To: desiredNetworkPolicyPeers,
	}
	effectiveNetworkPolicy.Spec.Egress = append(targetNetworkPolicy.Spec.Egress, dnsPolicyEgressRule)

	if targetNetworkPolicy.Spec.PodSelector.MatchLabels == nil {
		targetNetworkPolicy.Spec.PodSelector.MatchLabels = make(map[string]string)
	}
	targetNetworkPolicy.Spec.PodSelector.MatchLabels[key.PodSelectorMatchLabelRandom] = key.RandomLabel()

	_, err = r.k8sClient.NetworkingV1().NetworkPolicies(ns).Update(effectiveNetworkPolicy)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.k8sClient.NetworkingV1().NetworkPolicies(ns).Update(targetNetworkPolicy)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}
