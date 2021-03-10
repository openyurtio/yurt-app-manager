# Yurt-app-manager Tutorial 

This document introduces how to use NodePool CRD to manage nodes by pools. 

### Setting up the Yurt-app-manager
The following example is using minikube as the execution environment.

1. Provisioning a two-nodes [minikube](https://minikube.sigs.k8s.io/docs/) cluster.
```
# on OSX 
$ minikube start --nodes 2 --driver hyperkit
# on Linux
$ minikube start --nodes 2 --driver virtualbox
```

2. Generate the yaml file for deploying the yurt-app-manager.  
```bash
$ make generate-deploy-yaml
```
The output file will be found under `_output/yamls/`

3. Deploying the yurt-app-manager
```bash
$ kubectl apply -f _output/yamls/yurt-app-manager.yaml
```
If everything goes smoothly, the yurt-app-manager will be up and running in a minute.

### Grouping Nodes into Pools

1. Applying the NodePool CRD
```bash
$ kubectl apply -f config/yurt-app-manager/samples/nodepool/apps_v1alpha1_nodepool_edge.yaml
```

2. We can join a node into the Pool by adding the label `apps.openyurt.io/desired-nodepool=<nodepool-name>` 
to the node
```bash
$ kubectl label node/minikube-m02 apps.openyurt.io/desired-nodepool=hangzhou
```

3. Then, the NodePool attribute (e.g., labels, annotations, and tolerations) will 
be tagged on the node
```bash
$ kubectl get nodepool hangzhou -o yaml
apiVersion: apps.openyurt.io/v1alpha1
kind: NodePool
metadata:
  ...
  name: hangzhou
  ...
spec:
  annotations:
    test.openyurt.io: test-hangzhou
  labels:
    test.openyurt.io: test-hangzhou
  selector:
    matchLabels:
      apps.openyurt.io/nodepool: hangzhou
  taints:
  - effect: NoSchedule
    key: test.openyurt.io
    value: test-hangzhou
  type: Edge
status:
  nodes:
  - minikube-m02
  readyNodeNum: 1
  unreadyNodeNum: 0
...
$ kubectl get node minikube-m02 -o yaml
metadata:
  annotations: 
    ...
    test.openyurt.io: test-hangzhou
    ...
  labels:
    ...
    apps.openyurt.io/desired-nodepool: hangzhou
    apps.openyurt.io/nodepool: hangzhou
    test.openyurt.io: test-hangzhou
    ...
spec:
  taints:
  - effect: NoSchedule
    key: test.openyurt.io
    value: test-hangzhou
...
```

4. If we add/remove attributes to/from the NodePool, they will be tagged/removed 
to/from nodes belonging to the NodePool

5. We can remove a node from a pool by deleting the label `apps.openyurt.io/desired-nodepool` 
on the node, and all pool related attributes will be removed from the node.
```bash
$ kubectl label node/minikube-m02 apps.openyurt.io/desired-nodepool-
```

NOTE: The node migration is not supported. Migrating nodes from one pool to 
another may result in undefined behaviors, as the node migration normally 
requires network resetting. However, Kubernetes does not support assigning a new 
value to the filed `Node.Spec.podCIDR` if it is non-empty.

### Deploying Workloads by Pools

TODO
