package dnsnetworkpolicy

import (
	"context"
	"net"
	"sync"

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

	if len(cr.Spec.Domains) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no domains found")
		return nil
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
	effectiveNetworkPolicy := targetNetworkPolicy.DeepCopy()
	effectiveNetworkPolicy.ObjectMeta.ResourceVersion = ""
	effectiveNetworkPolicy.ObjectMeta.UID = ""
	effectiveNetworkPolicy.ObjectMeta.Name = effectiveNetworkPolicyName
	delete(effectiveNetworkPolicy.Spec.PodSelector.MatchLabels, key.PodSelectorMatchLabelRandom)

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
	ipChan := make(chan net.IP, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go resolveDomains(cr.Spec.Domains, &wg, ipChan, r.resolver.RoundRobinAttempts)
	wg.Add(1)
	uniqueIPs := collectResult(&wg, ipChan)
	wg.Wait()

	var desiredNetworkPolicyPeers []networkingV1.NetworkPolicyPeer
	{
		for _, cidr := range uniqueIPs {
			desiredNetworkPolicyPeer := networkingV1.NetworkPolicyPeer{
				IPBlock: &networkingV1.IPBlock{
					CIDR: cidr,
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

	_, err = r.k8sClient.NetworkingV1().NetworkPolicies(ns).Get(effectiveNetworkPolicyName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = r.k8sClient.NetworkingV1().NetworkPolicies(ns).Create(effectiveNetworkPolicy)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if err != nil {
		return microerror.Mask(err)
	}

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
