/*
Copyright 2023 The OpenYurt Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package staticpod

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	nodeutil "k8s.io/kubernetes/pkg/controller/util/node"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/controller/staticpod/info"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/gate"
)

const (
	controllerName = "static-pod-controller"

	StaticPodHashAnnotation = "openyurt.io/static-pod-hash"

	hostPathVolumeName       = "hostpath"
	hostPathVolumeMountPath  = "/etc/kubernetes/manifests/"
	configMapVolumeName      = "configmap"
	configMapVolumeMountPath = "/data"
	hostPathVolumeSourcePath = hostPathVolumeMountPath

	// UpgradeWorkerPodPrefix is the name prefix of worker pod which used for static pod upgrade
	UpgradeWorkerPodPrefix     = "static-pod-upgrade-worker-"
	UpgradeWorkerContainerName = "upgrade-worker"
	UpgradeWorkerImage         = "openyurt/yurt-static-pod-upgrade:v1.0.0-3531a59"
	UpgradeServiceAccount      = "yurt-app-manager"

	concurrentReconciles = 3

	Auto = "auto"
	OTA  = "ota"

	ArgTmpl = "/usr/local/bin/yurt-static-pod-upgrade --name=%s --manifest=%s --hash=%s --namespace=%s --mode=%s"
)

// StaticPodReconciler reconciles a StaticPod object
type StaticPodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	True                      = true
	upgradeSuccessCondition   = NewStaticPodCondition(appsv1alpha1.StaticPodUpgradeSuccess, corev1.ConditionTrue, "", "")
	upgradeExecutingCondition = NewStaticPodCondition(appsv1alpha1.StaticPodUpgradeExecuting, corev1.ConditionTrue, "", "")
)

// upgradeWorker is the pod template used for static pod upgrade
var (
	upgradeWorker = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceSystem},
		Spec: corev1.PodSpec{
			HostPID:            true,
			HostNetwork:        true,
			RestartPolicy:      corev1.RestartPolicyNever,
			ServiceAccountName: UpgradeServiceAccount,
			Containers: []corev1.Container{{
				Name:    UpgradeWorkerContainerName,
				Command: []string{"/bin/sh", "-c"},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      hostPathVolumeName,
						MountPath: hostPathVolumeMountPath,
					},
					{
						Name:      configMapVolumeName,
						MountPath: configMapVolumeMountPath,
					},
				},
				ImagePullPolicy: corev1.PullIfNotPresent,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &True,
				},
				Image: UpgradeWorkerImage,
			}},
			Volumes: []corev1.Volume{{
				Name: hostPathVolumeName,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: hostPathVolumeSourcePath,
					},
				}},
			},
		},
	}
)

// Add creates a new StaticPod Controller and adds it to the Manager with default RBAC.
// The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, _ context.Context) error {
	if !gate.ResourceEnabled(&appsv1alpha1.StaticPod{}) {
		return nil
	}
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &StaticPodReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r, MaxConcurrentReconciles: concurrentReconciles})
	if err != nil {
		return err
	}

	// 1. Watch for changes to StaticPod
	if err := c.Watch(&source.Kind{Type: &appsv1alpha1.StaticPod{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	// 2. Watch for changes to node
	// When node turn ready, reconcile all StaticPod instances
	// nodeReadyPredicate filter events which are node turn ready
	nodeReadyPredicate := predicate.Funcs{
		UpdateFunc: func(evt event.UpdateEvent) bool {
			return nodeTurnReady(evt)
		},
	}

	reconcileAllStaticPods := func(c client.Client) []reconcile.Request {
		staticPodList := &appsv1alpha1.StaticPodList{}
		err := c.List(context.TODO(), staticPodList)
		if err != nil {
			return nil
		}
		var requests []reconcile.Request
		for _, staticPod := range staticPodList.Items {
			requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
				Name: staticPod.Name,
			}})
		}
		return requests
	}

	if err := c.Watch(&source.Kind{Type: &corev1.Node{}},
		handler.EnqueueRequestsFromMapFunc(
			func(client.Object) []reconcile.Request {
				return reconcileAllStaticPods(mgr.GetClient())
			}), nodeReadyPredicate); err != nil {
		return err
	}

	// 3. Watch for changes to upgrade worker pods which are created by static-pod-controller
	if err := c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{IsController: true, OwnerType: &appsv1alpha1.StaticPod{}}); err != nil {
		return err
	}

	return nil
}

// nodeTurnReady filter events: old node is not-ready or unknown, new node is ready
func nodeTurnReady(evt event.UpdateEvent) bool {
	if _, ok := evt.ObjectOld.(*corev1.Node); !ok {
		return false
	}

	oldNode := evt.ObjectOld.(*corev1.Node)
	newNode := evt.ObjectNew.(*corev1.Node)

	_, onc := nodeutil.GetNodeCondition(&oldNode.Status, corev1.NodeReady)
	_, nnc := nodeutil.GetNodeCondition(&newNode.Status, corev1.NodeReady)

	oldReady := (onc != nil) && ((onc.Status == corev1.ConditionFalse) || (onc.Status == corev1.ConditionUnknown))
	newReady := (nnc != nil) && (nnc.Status == corev1.ConditionTrue)

	return oldReady && newReady
}

var _ reconcile.Reconciler = &StaticPodReconciler{}

//+kubebuilder:rbac:groups=apps.openyurt.io,resources=staticpods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.openyurt.io,resources=staticpods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.openyurt.io,resources=staticpods/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *StaticPodReconciler) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(4).Infof("Reconcile StaticPod %s", req.Name)

	instance := &appsv1alpha1.StaticPod{}
	if err := r.Get(context.TODO(), req.NamespacedName, instance); err != nil {
		klog.Errorf("Fail to get StaticPod %v, %v", req.NamespacedName.Name, err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var (
		// totalNumber represents the total number of nodes running the target static pod
		totalNumber int32

		// desiredNumber represents the desired upgraded number of nodes running the target static pod
		// In auto mode: it's the number of ready nodes running the target static pod
		// In ota mode: it's equal to totalNumber
		desiredNumber int32

		// upgradedNumber represents the number of nodes that have been upgraded
		upgradedNumber int32
	)

	// The later upgrade operation is done based on staticPodsInfoList
	staticPodsInfoList, err := info.ConstructStaticPodsUpgradeInfoList(r.Client, instance, UpgradeWorkerPodPrefix)
	if err != nil {
		klog.Errorf("Fail to get static pod and worker pod staticPodInfo for nodes of StaticPod %v, %v", req.NamespacedName.Name, err)
		return ctrl.Result{}, err
	}

	totalNumber = int32(len(staticPodsInfoList))

	// There are no nodes running target static pods in the cluster
	if totalNumber == 0 {
		klog.Infof("No static pods need to be upgraded of StaticPod %v", req.NamespacedName.Name)
		// Todo: set to number
		return r.updateCondition(instance, upgradeSuccessCondition)
	}

	// The latest hash value for static pod spec
	// This hash value is used in three places
	// 1. Automatically added to the annotation of static pods to facilitate checking if the running static pods are up-to-date
	// 2. Automatically added to the annotation of worker pods to facilitate checking if the worker pods are up-to-date
	// 3. Added to static pods' corresponding configmap to facilitate checking if the configmap is up-to-date
	latestHash := ComputeHash(&instance.Spec.Template)

	// The latest static pod manifest generated from user-specified template
	// The above hash value will be added to the annotation
	latestManifest, err := GenStaticPodManifest(&instance.Spec.Template, latestHash)
	if err != nil {
		klog.Errorf("Fail to generate static pod manifest of StaticPod %v, %v", req.NamespacedName.Name, err)
		return ctrl.Result{}, err
	}

	// Sync the corresponding configmap to the latest state
	err = r.syncConfigMap(instance, latestHash, latestManifest)
	if err != nil {
		klog.Errorf("Fail to sync the corresponding configmap of StaticPod %v, %v", req.NamespacedName.Name, err)
		return ctrl.Result{}, err
	}

	// Calculate the number of nodes that have been upgraded and set upgrade related flags
	for _, staticPodInfo := range staticPodsInfoList {
		if staticPodInfo.HasStaticPod {
			if staticPodInfo.StaticPod.Annotations[StaticPodHashAnnotation] != latestHash {
				// Indicate the static pod in this node needs to be upgraded
				staticPodInfo.UpgradeNeeded = true
				// This flag may be modified later when checking worker pods
				staticPodInfo.UpgradeExecuting = false
				continue
			}
			upgradedNumber++
		}
	}

	// allSucceeded flag is used to indicate whether the worker pods in last round have been all succeeded.
	// In auto upgrade mode, if the value is false, it will wait util all the worker pods succeed.
	allSucceeded := true
	for node, staticPodInfo := range staticPodsInfoList {
		// set upgradeWaiting and deal with workerPods
		if staticPodInfo.HasWorkerPod {
			hash := staticPodInfo.WorkerPod.Annotations[StaticPodHashAnnotation]

			// If the worker pod is not up-to-date, then it can be recreated directly
			if hash != latestHash {
				if err := r.Delete(context.TODO(), staticPodInfo.WorkerPod, &client.DeleteOptions{}); err != nil {
					klog.Errorf("Fail to delete out-of-date worker pod %s of StaticPod %v, %v", staticPodInfo.WorkerPod.Name, req.NamespacedName.Name, err)
					return ctrl.Result{}, err
				}

				continue
			}

			// If the worker pod is up-to-date, there are three possible situations
			// 1. The worker pod is failed, then some irreparable failure has occurred. Just stop reconcile and update status
			// 2. The worker pod is succeeded, then this node must be up-to-date. Just delete this worker pod
			// 3. The worker pod is running, pending or unknown, then just wait
			switch staticPodInfo.WorkerPod.Status.Phase {
			case corev1.PodFailed:
				klog.Errorf("Fail to continue upgrade, cause worker pod %s of StaticPod %v in node %s failed", staticPodInfo.WorkerPod.Name, req.NamespacedName.Name, node)
				return r.updateCondition(instance, UpgradeFailedConditionWithNode(node))

			case corev1.PodSucceeded:
				klog.V(4).Infof("Static pod upgrade successfully in node %s of StaticPod %v", node, req.NamespacedName.Name)
				// todo, 这些成功的工作节点可以被删除 , 以及上面得到删除，是不是最好剥离出来
				if err := r.Delete(context.TODO(), staticPodInfo.WorkerPod, &client.DeleteOptions{}); err != nil {
					klog.Errorf("Fail to delete successful worker pod %s of StaticPod %v,%v", staticPodInfo.WorkerPod.Name, req.NamespacedName.Name, err)
					return ctrl.Result{}, err
				}

			default:
				// In this node, the latest worker pod is still running, and we don't need to create new worker for it.
				staticPodInfo.UpgradeExecuting = true
				allSucceeded = false
			}
		}
	}

	// Check the ready status for every node which has the target static pod
	for node, staticPodInfo := range staticPodsInfoList {
		ready, err := NodeReadyByName(r.Client, node)
		if err != nil {
			klog.Errorf("Fail to check ready status for node %s of StaticPod %v,%v", node, req.NamespacedName.Name, err)
			return ctrl.Result{}, err
		}
		staticPodInfo.Ready = ready
	}

	// If all nodes have been upgraded, just return
	if totalNumber == upgradedNumber {
		klog.Infof("All static pods have been upgraded of StaticPod %v", req.NamespacedName.Name)
		return r.updateStatus(instance, totalNumber, instance.Status.DesiredNumber, upgradedNumber, upgradeSuccessCondition)
	}

	switch instance.Spec.UpgradeStrategy.Type {
	// Auto Upgrade is to automate the upgrade process for the target static pods on ready nodes
	// It supports rolling update and the max-unavailable number can be specified by users
	case appsv1alpha1.AutoStaticPodUpgradeStrategyType:
		if !allSucceeded {
			klog.V(4).Infof("Wait last round auto upgrade to finish of StaticPod %v", req.NamespacedName.Name)
			return ctrl.Result{}, nil
		}

		// In auto upgrade mode, desiredNumber is the number of ready nodes
		readyNodes := info.ReadyNodes(staticPodsInfoList)
		desiredNumber = int32(len(readyNodes))
		// This means that all the desired nodes are upgraded. It's considered successful.
		if desiredNumber == upgradedNumber {
			return r.updateStatus(instance, totalNumber, desiredNumber, upgradedNumber, upgradeSuccessCondition)
		}

		err := r.autoUpgrade(instance, staticPodsInfoList, latestHash)
		if err != nil {
			klog.Errorf("Fail to auto upgrade of StaticPod %v, %v", req.NamespacedName.Name, err)
			return ctrl.Result{}, err
		}
		return r.updateStatus(instance, totalNumber, desiredNumber, upgradedNumber, upgradeExecutingCondition)

	// OTA Upgrade can help users control the time of static pods upgrade
	// It will set PodNeedUpgrade condition and work with YurtHub component
	case appsv1alpha1.OTAStaticPodUpgradeStrategyType:
		if err := r.otaUpgrade(instance, staticPodsInfoList, latestHash); err != nil {
			klog.Errorf("Fail to ota upgrade of StaticPod %v, %v", req.NamespacedName.Name, err)
			return ctrl.Result{}, err
		}
		return r.updateStatus(instance, totalNumber, totalNumber, upgradedNumber, upgradeExecutingCondition)
	}

	return ctrl.Result{}, nil
}

// syncConfigMap moves the target static pod's corresponding configmap to the latest state
func (r *StaticPodReconciler) syncConfigMap(instance *appsv1alpha1.StaticPod, hash, data string) error {
	cm := &corev1.ConfigMap{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: WithConfigMapPrefix(instance.Spec.Namespace + "-" + instance.Spec.StaticPodName), Namespace: metav1.NamespaceSystem}, cm)
	if err != nil {
		// if the configmap does not exist, then create a new one
		if kerr.IsNotFound(err) {
			cm = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      WithConfigMapPrefix(instance.Spec.Namespace + "-" + instance.Spec.StaticPodName),
					Namespace: metav1.NamespaceSystem,
					Annotations: map[string]string{
						StaticPodHashAnnotation: hash,
					},
				},

				Data: map[string]string{
					instance.Spec.StaticPodManifest: data,
				},
			}
			err = r.Create(context.TODO(), cm, &client.CreateOptions{})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	// if the hash value in the annotation of the cm does not match the latest hash, then update the data in the cm
	if cm.Annotations[StaticPodHashAnnotation] != hash {
		cm.Annotations[StaticPodHashAnnotation] = hash
		cm.Data[instance.Spec.StaticPodManifest] = data

		err = r.Update(context.TODO(), cm, &client.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

// autoUpgrade automatically rolling upgrade the target static pods in cluster
func (r *StaticPodReconciler) autoUpgrade(instance *appsv1alpha1.StaticPod, staticPodsInfoList map[string]*info.StaticPodInfo, hash string) error {
	// readyUpgradeWaitingNodes represents nodes that need to create worker pods
	readyUpgradeWaitingNodes := info.ReadyUpgradeWaitingNodes(staticPodsInfoList)

	waitingNumber := len(readyUpgradeWaitingNodes)
	if waitingNumber == 0 {
		return nil
	}

	// max is the maximum number of nodes can be upgraded in current round in auto upgrade mode
	max, err := UnavailableCount(&instance.Spec.UpgradeStrategy, len(staticPodsInfoList))
	if err != nil {
		return err
	}

	if waitingNumber < max {
		max = waitingNumber
	}

	readyUpgradeWaitingNodes = readyUpgradeWaitingNodes[:max]
	if err := createUpgradeWorker(r.Client, instance, readyUpgradeWaitingNodes, hash, Auto); err != nil {
		return err
	}
	return nil
}

// otaUpgrade adds condition PodNeedUpgrade to the target static pods and issue the latest manifest to corresponding nodes
func (r *StaticPodReconciler) otaUpgrade(instance *appsv1alpha1.StaticPod, staticPodsInfoList map[string]*info.StaticPodInfo, hash string) error {
	upgradeNeededNodes := info.UpgradeNeededNodes(staticPodsInfoList)
	upgradedNodes := info.UpgradedNodes(staticPodsInfoList)

	// Set condition for upgrade needed static pods
	for _, n := range upgradeNeededNodes {
		if err := SetPodUpgradeCondition(r.Client, corev1.ConditionTrue, staticPodsInfoList[n].StaticPod); err != nil {
			return err
		}
	}

	// Set condition for upgraded static pods
	for _, n := range upgradedNodes {
		if err := SetPodUpgradeCondition(r.Client, corev1.ConditionFalse, staticPodsInfoList[n].StaticPod); err != nil {
			return err
		}
	}

	// Create worker pod to issue the latest manifest to ready node
	readyUpgradeWaitingNodes := info.OTAReadyUpgradeWaitingNodes(staticPodsInfoList, hash)
	if err := createUpgradeWorker(r.Client, instance, readyUpgradeWaitingNodes, hash, OTA); err != nil {
		return err
	}

	return nil
}

// createUpgradeWorker creates static pod upgrade worker to the given nodes
func createUpgradeWorker(c client.Client, instance *appsv1alpha1.StaticPod, nodes []string, hash, mode string) error {
	for _, node := range nodes {
		pod := upgradeWorker.DeepCopy()
		pod.Name = UpgradeWorkerPodPrefix + node
		pod.Spec.NodeName = node
		metav1.SetMetaDataAnnotation(&pod.ObjectMeta, StaticPodHashAnnotation, hash)
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name: configMapVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: WithConfigMapPrefix(instance.Spec.Namespace + "-" + instance.Spec.StaticPodName),
					},
				},
			},
		})
		pod.Spec.Containers[0].Args = []string{fmt.Sprintf(ArgTmpl, instance.Spec.StaticPodName+"-"+node, instance.Spec.StaticPodManifest, hash, instance.Spec.Namespace, mode)}
		if err := controllerutil.SetControllerReference(instance, pod, c.Scheme()); err != nil {
			return err
		}

		if err := c.Create(context.TODO(), pod, &client.CreateOptions{}); err != nil {
			return err
		}
		klog.Infof("Create static pod upgrade worker %s of StaticPod %s", pod.Name, instance.Name)
	}

	return nil
}

// updateCondition only update condition of the given StaticPod CR
func (r *StaticPodReconciler) updateCondition(instance *appsv1alpha1.StaticPod, cond *appsv1alpha1.StaticPodCondition) (reconcile.Result, error) {
	instance.Status.Conditions = []appsv1alpha1.StaticPodCondition{*cond}
	return reconcile.Result{}, r.Client.Status().Update(context.TODO(), instance)
}

func (r *StaticPodReconciler) updateStatus(instance *appsv1alpha1.StaticPod, totalNum, desiredNum, upgradedNum int32, cond *appsv1alpha1.StaticPodCondition) (reconcile.Result, error) {
	instance.Status.Conditions = []appsv1alpha1.StaticPodCondition{*cond}
	instance.Status.TotalNumber = totalNum
	instance.Status.DesiredNumber = desiredNum
	instance.Status.UpgradedNumber = upgradedNum

	return reconcile.Result{}, r.Client.Status().Update(context.TODO(), instance)
}
