# Yurt-app-manager Tutorial 

This document introduces how to install yurt-app-manager and use yurt-app-manager to manage edge nodes and workloads. 

In this tutorial, we will show how the yurt-app-manager helps users manage 
there edge nodes and workload.
Suppose you have a Kubernetes cluster in an Openyurt environment, or a native Kubernetes cluster with at least two nodes.

## Label cloud nodes and edge nodes
``` bash
$ kubectl get nodes -o wide

NAME        STATUS   ROLES    AGE   VERSION   INTERNAL-IP    
k8s-node1   Ready    <none>   20d   v1.16.2   10.48.115.9    
k8s-node2   Ready    <none>   20d   v1.16.2   10.48.115.10   
master      Ready    master   20d   v1.16.2   10.48.115.8    
```
and we will use node `master` as the cloud node.

We label the cloud node with value `false`,
```bash
$ kubectl label node master openyurt.io/is-edge-worker=false
master labeled
```

and the edge node with value `true`.
```bash
$ kubectl label node k8s-node1 openyurt.io/is-edge-worker=true
k8s-node1 labeled
$ kubectl label node k8s-node2 openyurt.io/is-edge-worker=true
k8s-node2 labeled
```

## Install yurt-app-manager

### install yurt-app-manager operator 
```bash
$ cd  yurt-app-manager
$ kubectl apply -f config/setup/all_in_one.yaml
```

Wait for the yurt-app-manager operator  to be created successfully
``` bash
$ kubectl get pod -n kube-system |grep yurt-app-manager
```

## How to Use

The Examples of NodePool and YurtAppSet are in `config/yurt-app-manager/samples/` directory

### NodePool 

- 1 create an nodepool 
```bash
$ cat <<EOF | kubectl apply -f -
apiVersion: apps.openyurt.io/v1alpha1
kind: NodePool
metadata:
  name: beijing
spec:
  type: Cloud
EOF

$ cat <<EOF | kubectl apply -f -
apiVersion: apps.openyurt.io/v1alpha1
kind: NodePool
metadata:
  name: hangzhou
spec:
  type: Edge
  annotations:
    apps.openyurt.io/example: test-hangzhou
  labels:
    apps.openyurt.io/example: test-hangzhou
  taints:
  - key: apps.openyurt.io/example
    value: test-hangzhou
    effect: NoSchedule
EOF

```

- 2 Get NodePool
```bash
$ kubectl get np 

NAME       TYPE   READYNODES   NOTREADYNODES   AGE
beijing    Cloud                               35s
hangzhou   Edge                                28s
```

- 3 Add Node To NodePool

Add Your_Node_Name Cloud node into `beijing` NodePool, Set the `apps.openyurt.io/desired-nodepool` label on the host, and value is the name of the beijing NodePool
```bash
$ kubectl label node {Your_Node_Name} apps.openyurt.io/desired-nodepool=beijing
```
```
For example:
$ kubectl label node master apps.openyurt.io/desired-nodepool=beijing

master labeled
```
Add Your_Node_Name Edge node into `hangzhou` NodePool, Set the `apps.openyurt.io/desired-nodepool` label on the host, and value is the name of the hangzhou NodePool
```bash
$ kubectl label node {Your_Node_Name} apps.openyurt.io/desired-nodepool=hangzhou
```
```
For example:
$ kubectl label node k8s-node1 apps.openyurt.io/desired-nodepool=hangzhou

k8s-node1 labeled

$ kubectl label node k8s-node2 apps.openyurt.io/desired-nodepool=hangzhou

k8s-node2 labeled
```

```bash
$ kubectl get np 

NAME       TYPE    READYNODES   NOTREADYNODES   AGE
beijing    Cloud   1            0               140m
hangzhou   Edge    2            0               4h35m
```

Once a Edge Node adds a NodePool, it inherits the annotations, labels, and taints defined in the nodepool Spec,at the same time, the Node will add a new tag: `apps.openyurt.io/nodepool`. 
```bash
$ kubectl get node {Your_Node_Name} -o yaml 

For Example:
$ kubectl get node k8s-node1 -o yaml

apiVersion: v1
kind: Node
metadata:
  annotations:
    apps.openyurt.io/example: test-hangzhou
    kubeadm.alpha.kubernetes.io/cri-socket: /var/run/dockershim.sock
    node.alpha.kubernetes.io/ttl: "0"
    node.beta.alibabacloud.com/autonomy: "true"
    volumes.kubernetes.io/controller-managed-attach-detach: "true"
  creationTimestamp: "2021-04-14T12:17:39Z"
  labels:
    apps.openyurt.io/desired-nodepool: hangzhou
    apps.openyurt.io/example: test-hangzhou
    apps.openyurt.io/nodepool: hangzhou
    beta.kubernetes.io/arch: amd64
    beta.kubernetes.io/os: linux
    kubernetes.io/arch: amd64
    kubernetes.io/hostname: k8s-node1
    kubernetes.io/os: linux
    openyurt.io/is-edge-worker: "true"
  name: k8s-node1
  resourceVersion: "1244431"
  selfLink: /api/v1/nodes/k8s-node1
  uid: 1323f90b-acf3-4443-a7dc-7a54c212506c
spec:
  podCIDR: 192.168.1.0/24
  podCIDRs:
  - 192.168.1.0/24
  taints:
  - effect: NoSchedule
    key: apps.openyurt.io/example
    value: test-hangzhou
status:
***
```

### YurtAppSet

#### use yurtAppset
- 1 create an yurtappset which use deployment template

```bash
$ cat <<EOF | kubectl apply -f -
apiVersion: apps.openyurt.io/v1alpha1
kind: YurtAppSet
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: yas-test
spec:
  selector:
    matchLabels:
      app: yas-test
  workloadTemplate:
    deploymentTemplate:
      metadata:
        labels:
          app: yas-test
      spec:
        template:
          metadata:
            labels:
              app: yas-test
          spec:
            containers:
              - name: nginx
                image: nginx:1.19.3
  topology:
    pools:
    - name: beijing 
      nodeSelectorTerm:
        matchExpressions:
        - key: apps.openyurt.io/nodepool
          operator: In
          values:
          - beijing 
      replicas: 1
      patch:
        spec:
          template:
            spec:
              containers:
                - name: nginx
                  image: nginx:1.19.0
    - name: hangzhou 
      nodeSelectorTerm:
        matchExpressions:
        - key: apps.openyurt.io/nodepool
          operator: In
          values:
          - hangzhou 
      replicas: 2
      tolerations:
      - effect: NoSchedule
        key: apps.openyurt.io/example
        operator: Exists
  revisionHistoryLimit: 5 
EOF

```

- 2 Get YurtAppSet
```bash
$ kubectl get yas

NAME      READY   WORKLOADTEMPLATE   AGE
yas-test   3       Deployment         120m
```

check the sub deployment created by yurt-app-manager controller
```bash
$ kubectl get deploy

NAME                     READY   UP-TO-DATE   AVAILABLE   AGE
yas-test-beijing-fp58z    1/1     1            1           122m
yas-test-hangzhou-xv454   2/2     2            2           122m
```

```bash
$ kubectl get pod -l app=yas-test

  NAME                                      READY   STATUS    RESTARTS   AGE
yas-test-beijing-fp58z-787d5b6b54-g4jk6    1/1     Running   0          100m
yas-test-hangzhou-xv454-5cd9c4f6b5-b5tsr   1/1     Running   0          124m
yas-test-hangzhou-xv454-5cd9c4f6b5-gmbgp   1/1     Running   0          124m
```
#### yurtAppSet add patch
- 1 in 'config/yurt-app-manager/samples/yurtappset/yurtappset_deployment_test.yaml' file’s 35 to 41 lines
```bash
$ kubectl get yas yas-test -o yaml
   
  topology:
    pools:
    - name: beijing 
      nodeSelectorTerm:
        matchExpressions:
        - key: apps.openyurt.io/nodepool
          operator: In
          values:
          - beijing 
      replicas: 1
      patch:
        spec:
          template:
            spec:
              containers:
                - name: nginx
                  image: nginx:1.19.0
    - name: hangzhou 
      nodeSelectorTerm:
        matchExpressions:
        - key: apps.openyurt.io/nodepool
          operator: In
          values:
          - hangzhou 
      replicas: 2
      tolerations:
  *** 
```
- 2 Patch makes the image of deployment and pod named beijing created by yurtAppSet `nginx:1.19.0`,Other images used are `nginx:1.19.3`.
```bash
$ kubectl get deploy yas-test-beijing-fp58z -o yaml

containers:
  - image: nginx:1.19.0
$ kubectl get deploy yas-test-hangzhou-xv454 -o yaml

containers:
  - image: nginx:1.19.3
```
The result of pod is consistent with that of deployment.

- 3 After deleting this file, all the pods created by YurtAppSet use the same image: `nginx:1.19.3`. 
```bash
$ kubectl get pod yas-test-beijing-fp58z-787d5b6b54-g4jk6 -o yaml

containers:
  - image: nginx:1.19.3
$ kubectl get pod yas-test-hangzhou-xv454-5cd9c4f6b5-b5tsr -o yaml
containers:
  - image: nginx:1.19.3
```
- 4 conclusion
Patch solves the problem of single attribute upgrade and full release of nodepool.

 ### YurtAppDaemon
 For details please see the [tutorial](./YurtAppDaemon.md).

 ### YurtIngress
 For details please see the [tutorial](https://github.com/openyurtio/openyurt.io/blob/master/docs/user-manuals/network/edge-ingress.md).
