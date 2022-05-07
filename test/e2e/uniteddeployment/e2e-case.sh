
MasterNode=kind-control-plane
BeijingNode=kind-worker
HangzhouNode=kind-worker2

kubectl label node ${MasterNode} openyurt.io/is-edge-worker=false

kubectl label node ${BeijingNode} openyurt.io/is-edge-worker=true

kubectl label node ${HangzhouNode} openyurt.io/is-edge-worker=true

cat <<EOF | kubectl apply -f -
apiVersion: apps.openyurt.io/v1alpha1
kind: NodePool
metadata:
  name: beijing
spec:
  type: Cloud
EOF

cat <<EOF | kubectl apply -f -
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

kubectl label node ${BeijingNode} apps.openyurt.io/desired-nodepool=beijing

kubectl label node ${HangzhouNode} apps.openyurt.io/desired-nodepool=hangzhou

cat <<EOF | kubectl apply -f -
apiVersion: apps.openyurt.io/v1alpha1
kind: UnitedDeployment
metadata:
  name: ud-test
spec:
  selector:
    matchLabels:
      app: ud-test
  workloadTemplate:
    deploymentTemplate:
      metadata:
        labels:
          app: ud-test
      spec:
        template:
          metadata:
            labels:
              app: ud-test
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
