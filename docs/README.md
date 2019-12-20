# Use cases

## Allow egress traffic to Kubernetes API via internal service

This case is valid for clusters, where Kubernetes API is running as a static pod. As it uses host networking, you can't apply a regular pod selector to egress rules.
There is `kubernetes.default` service with `ClusterIP` available to access Kubernetes API.
However, this domain can't be used within `dnsnetworkpolicy` CR as network policies applied to pod IPs, not a virtual service IPs.
Therefore, you need to create a headless service, which exposes real pod IP addresses first.

1. Make sure you have labels on your Kubernetes API pods

```
kubectl get pods -l k8s-app=apiserver -n kube-system
NAME                                                         READY   STATUS    RESTARTS   AGE   IP           NODE                                          NOMINATED NODE   READINESS GATES
k8s-api-server-ip-10-0-5-171.eu-central-1.compute.internal   1/1     Running   0          51m   10.0.5.171   ip-10-0-5-171.eu-central-1.compute.internal   <none>           <none>
k8s-api-server-ip-10-0-5-34.eu-central-1.compute.internal    1/1     Running   0          51m   10.0.5.34    ip-10-0-5-34.eu-central-1.compute.internal    <none>           <none>
k8s-api-server-ip-10-0-5-74.eu-central-1.compute.internal    1/1     Running   0          50m   10.0.5.74    ip-10-0-5-74.eu-central-1.compute.internal    <none>           <none>
```

2. Create headless service for API

```
apiVersion: v1
kind: Service
metadata:
  labels:
    component: apiserver
    provider: kubernetes
  name: kubernetes-headless
  namespace: kube-system
spec:
  clusterIP: None
  selector:
    k8s-app: apiserver
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: 443
```

3. Create network policy for your application

This policy allows an egress traffic to the DNS service only.

```
apiVersion: extensions/v1beta1
kind: NetworkPolicy                                                                                                
metadata:      
  name: example
  namespace: default
spec:
  egress:
  - ports:
    - port: 53
      protocol: UDP
    - port: 53
      protocol: TCP
  - to:
    - namespaceSelector: {}
  ingress:
  - {}
  podSelector:
    matchLabels:
      type: example
  policyTypes:
  - Ingress
  - Egress
```

3. Create `dnsnetworkpolicy` CR

```
apiVersion: example.giantswarm.io/v1alpha1
kind: DNSNetworkPolicy
metadata:
  name: example
  namespace: default
spec:
  targetNetworkPolicy: example
  domains:
  - kubernetes.kube-system
```

After CR is created, `dns-network-policy-operator` resolves the listed domain and create the new effective network policy:

```
kubectl get networkpolicies example-active -o yaml

apiVersion: extensions/v1beta1
kind: NetworkPolicy                                 
metadata:     
  annotations:                    
    giantswarm.io/managed-by: dns-network-policy-operator 
    giantswarm.io/network-policy-role: active         
  name: example-active
  namespace: default          
spec:                     
  egress:     
  - ports:                    
    - port: 53
      protocol: UDP            
    - port: 53
      protocol: TCP        
  - to:   
    - namespaceSelector: {}
  - to:       
    - ipBlock:
        cidr: 10.0.5.34/32
    - ipBlock:
        cidr: 10.0.5.74/32
    - ipBlock:
        cidr: 10.0.5.171/32
  ingress:
  - {}
  podSelector:
    matchLabels:
      type: example
  policyTypes:
  - Ingress
  - Egress
```

