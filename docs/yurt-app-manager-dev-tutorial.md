# Yurt-app-manager Tutorial for Developor

This document introduces how to build and install yurt-app-manager controller. 

##  clone yurt-app-manger code
```
# cd $GOPATH/src/github.com/openyurtio
# git clone git@github.com:openyurtio/yurt-app-manager.git
# cd  yurt-app-manager
```

## push yurt-app-manager image to your own registry

for example
```
make push REPO=registry.cn-your-registry.com/edge-kubernetes
```

if REPO value is assigned `registry.cn-your-registry.com/edge-kubernetes`， the `make push ` command will eventually build an image named registry.cn-your-registry.com/edge-kubernetes/yurt-app-manager:{git commit id} and push it into your  own repository. And `make push` command will also create a file named `yurt-app-manager.yaml` in _output/yamls dir. You need to set the REPO variable correctly。

## install nodepool and uniteddeployment controller
```
kubectl apply -f _output/yamls/yurt-app-manager.yaml
```

## check 

>  use `kubectl get crd` command to check that the CRD is successfully installed
```
# kubectl get crd
nodepools.apps.openyurt.io                       2021-04-23T08:54:31Z
uniteddeployments.apps.openyurt.io               2021-04-23T08:54:31Z
```

> use `kubectl get pod -n kube-system` command to check whether the yurt-app-manager pod is running 
```
# kubectl get pod -n kube-system
yurt-app-manager-78f657cbf4-c94gm                   1/1     Running     0          5d2h
yurt-app-manager-78f657cbf4-zwt22                   1/1     Running     0          5d2h
```

