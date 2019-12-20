[![CircleCI](https://circleci.com/gh/giantswarm/dns-network-policy-operator.svg?&style=shield)](https://circleci.com/gh/giantswarm/dns-network-policy-operator) [![Docker Repository on Quay](https://quay.io/repository/giantswarm/dns-network-policy-operator/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/dns-network-policy-operator)

# dns-network-policy-operator

This `dns-network-policy-operator` manages Kubernetes network policies with DNS-based egress rules support.

## Architecture

The operator uses our operatorkit framework. It manages a `dnsnetworkpolicy` CRD using a generated client stored in our apiextensions repo.

Basic DNS network policy may look like:

```
apiVersion: example.giantswarm.io/v1alpha1
kind: DNSNetworkPolicy
metadata:
  name: example
  namespace: default
spec:
  targetNetworkPolicy: example
  domains:
  - example.com
  - google.com
  - kubernetes-headless.kube-system
```

There are only two configurable fields in CR:

    - `domains` - list of domains, which are allowed for egress traffic
    - `targetNetworkPolicy` - this is base policy, used to generate new effective policy with IPs of resolved domains

### How does it work?

1. `dns-network-policy-opererator` reconciles `dnsnetworkpolicy` CR. 
2. If there is `targetNetworkPolicy` network policy found in the CR namespace,
it is duplicated into new network policy with `<target network policy name>-active`.
3. All the domains from CR are resolved into IP addresses. Failing resolves ignored.
4. Newly created effective network policy gets updated with list of resolved IP addresses.
5. `targetNetworkPolicy` supressed by adding random label into pod selector of the policy.

## Samples

You can find more samples in [samples](docs/README.md).
