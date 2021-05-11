package uniteddaemonset

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/refmanager"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

type TemplateControllor interface {
	GetAllWorkloads(set *v1alpha1.UnitedDaemonSet) ([]*Workload, error)
	CreateWorkload(set *v1alpha1.UnitedDaemonSet, nodepool v1alpha1.NodePool) error
}

type DeploymentControllor struct {
	client.Client
	scheme *runtime.Scheme
}

func (d *DeploymentControllor) CreateWorkload(set *v1alpha1.UnitedDaemonSet, nodepool v1alpha1.NodePool) error {
	deploy := appsv1.Deployment{
		ObjectMeta: set.Spec.WorkloadTemplate.DeploymentTemplate.ObjectMeta,
		Spec:       set.Spec.WorkloadTemplate.DeploymentTemplate.Spec,
	}

	deploy.ObjectMeta.Namespace = set.GetNamespace()
	deploy.ObjectMeta.GenerateName = getWorkloadPrefix(set.GetName(), nodepool.GetName())
	deploy.ObjectMeta.Name = ""

	annotations := deploy.ObjectMeta.GetAnnotations()
	annotations[v1alpha1.AnnotationRefNodePool] = nodepool.GetName()
	deploy.ObjectMeta.SetAnnotations(annotations)

	//selector
	deploy.Spec.Template.Spec.NodeSelector = nodepool.Spec.Selector.MatchLabels
	if deploy.Spec.Template.Spec.Affinity != nil && deploy.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		deploy.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = nil
	}

	// toleration
	tolerations := []corev1.Toleration{}
	for _, taint := range nodepool.Spec.Taints {
		toleation := corev1.Toleration{
			Key:      taint.Key,
			Operator: corev1.TolerationOpExists,
			Value:    taint.Value,
			Effect:   taint.Effect,
		}
		tolerations = append(tolerations, toleation)
	}
	deploy.Spec.Template.Spec.Tolerations = tolerations

	return d.Client.Create(context.TODO(), &deploy)
}

func (d *DeploymentControllor) GetAllWorkloads(set *v1alpha1.UnitedDaemonSet) ([]*Workload, error) {
	allDeployments := appsv1.DeploymentList{}
	// 获得UnitedDaemonset 对应的 所有Deployment, 根据OwnerRef
	selector, err := metav1.LabelSelectorAsSelector(set.Spec.Selector)
	if err != nil {
		return nil, err
	}
	// List all Deployment to include those that don't match the selector anymore but
	// have a ControllerRef pointing to this controller.
	if err := d.Client.List(context.TODO(), &allDeployments, &client.ListOptions{LabelSelector: selector}); err != nil {
		return nil, err
	}

	manager, err := refmanager.New(d.Client, set.Spec.Selector, set, d.scheme)
	if err != nil {
		return nil, err
	}

	selected := make([]metav1.Object, len(allDeployments.Items))
	for i := 0; i < len(allDeployments.Items); i++ {
		var t = &(allDeployments.Items[i])
		selected = append(selected, t)
	}

	objs, err := manager.ClaimOwnedObjects(selected)
	if err != nil {
		return nil, err
	}

	workloads := make([]*Workload, 0, len(objs))
	for i, o := range objs {
		spec := o.(*appsv1.Deployment).Spec
		w := &Workload{
			Name:      o.GetName(),
			Namespace: o.GetNamespace(),
			Spec: WorkloadSpec{
				Ref: objs[i],
			},
			Status: WorkloadStatus{
				NodeSelector: spec.Template.Spec.NodeSelector,
				Toleration:   spec.Template.Spec.Tolerations,
			},
		}
		workloads = append(workloads, w)
	}
	return workloads, nil
}

type StatefulSetControllor struct {
	client.Client

	scheme *runtime.Scheme
}

var _ TemplateControllor = &DeploymentControllor{}

// var _ TemplateControllor = &StatefulSetControllor{}
