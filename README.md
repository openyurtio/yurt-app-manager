# Yurt-app-manager

**IMPORTANT: This project is no longer being actively maintained and has been archived.**

## Archived Project

This project has been archived and is no longer being actively maintained. This means you can view and copy the code, but cannot make changes or propose pull requests.

While you're here, feel free to review the code and learn from it. If you wish to use the code or revive the project, you can fork it to your own GitHub account.

## Project Description

This repository contains 4 CRD/controllers: NodePool, YurtAppSet, YurtAppDaemon and YurtIngress.

The NodePool provides a convenient management experience for a pool of nodes within the same region or site.

The YurtAppSet defines a new edge application management methodology of using per node pool workload.

The YurtAppDaemon provides a similar K8S DaemonSet support for user app workload from the NodePool level.

The YurtIngress is responsible to deploy configurable ingress controller to the user specified NodePools.

For details of the design, please see the documents below:

NodePool and YurtAppSet: [document](https://github.com/openyurtio/openyurt/blob/master/docs/enhancements/20201211-nodepool_uniteddeployment.md).

YurtAppDaemon: [document](https://github.com/openyurtio/openyurt/blob/master/docs/enhancements/20210729-yurtappdaemon.md).

YurtIngress: [document](https://github.com/openyurtio/openyurt/blob/master/docs/proposals/20210628-nodepool-ingress-support.md).

## Previous Contributions

We want to take a moment to thank all of the previous contributors to this project. Your work has been greatly appreciated and has made a significant impact.

- [huiwq1990](https://github.com/huiwq1990)
- [kadisi](https://github.com/kadisi)
- [rambohe-ch](https://github.com/rambohe-ch)
- [luc99hen](https://github.com/luc99hen)
- [charleszheng44](https://github.com/charleszheng44)
- [rudolf-chy](https://github.com/rudolf-chy)
- [River-sh](https://github.com/River-sh)
- [YTGhost](https://github.com/YTGhost)
- [LindaYu17](https://github.com/LindaYu17)
- [JameKeal](https://github.com/JameKeal)
- [xavier-hou](https://github.com/xavier-hou)
- [ahmedwaleedmalik](https://github.com/ahmedwaleedmalik)
- [kyakdan](https://github.com/kyakdan)
- [donychen1134](https://github.com/donychen1134)
- [Congrool](https://github.com/Congrool)
- [cndoit18](https://github.com/cndoit18)
- [maoyangLiu](https://github.com/maoyangLiu)
- [wawlian](https://github.com/wawlian)
- [gnunu](https://github.com/gnunu)
- [cuisongliu](https://github.com/cuisongliu)
- [ZBoIsHere](https://github.com/ZBoIsHere)
- [yanyhui](https://github.com/yanyhui)

## Alternative Projects

All the functions of this project have been migrated into `yurt-manager` component in [openyurt](https://github.com/openyurtio/openyurt) repo.

- [controllers](https://github.com/openyurtio/openyurt/tree/master/pkg/yurtmanager/controller)
- [webhooks](https://github.com/openyurtio/openyurt/tree/master/pkg/yurtmanager/webhook)