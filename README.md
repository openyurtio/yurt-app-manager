# Yurt-app-manager

This repository contains two CRD/controllers, NodePool and UnitedDeployment.
The NodePool provides a convenient management experience for a pool of nodes 
within the same region or site. The UnitedDeployment defines a new edge 
application management methodology of using per node pool workload. For details 
of the design, please see the [document](https://github.com/openyurtio/openyurt/blob/master/docs/enhancements/20201211-nodepool_uniteddeployment.md).

## Getting Start

Since the OpenYurt is extended from the upstream Kubernetes using only plugins,
the NodePool and UnitedDeployment can be used with upstream Kubernetes as well. 
But to make the best use of them, we recommend using them with the OpenYurt. 
For a complete example, please check out the [tutorial](docs/yurt-app-manager-tutorial.md)

## Contributing 

Contributions are welcome, whether by creating new issues or pull requests. See 
our [contributing document](https://github.com/openyurtio/openyurt/blob/master/CONTRIBUTING.md) to get started.

## Contact

- Mailing List: openyurt@googlegroups.com
- Slack: [channel](https://join.slack.com/t/openyurt/shared_invite/zt-iw2lvjzm-MxLcBHWm01y1t2fiTD15Gw)
- Dingtalk Group (钉钉讨论群)

<div align="left">
    <img src="https://github.com/openyurtio/openyurt/blob/master/docs/img/ding.jpg" width=25% title="dingtalk">
</div>

## License
Yurt-app-manager is under the Apache 2.0 license. See the [LICENSE](LICENSE) file 
for details. Certain implementations in Yurt-app-manager rely on the existing code 
from [Kubernetes](https://github.com/kubernetes/kubernetes) and 
[OpenKruise](https://github.com/openkruise/kruise) the credits go to the 
original authors.
