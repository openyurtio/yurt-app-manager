package uniteddaemonset

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Workload struct {
	Name      string
	Namespace string
	Spec      WorkloadSpec
	Status    WorkloadStatus
}

// WorkloadSpec stores the spec details of the workload
type WorkloadSpec struct {
	Ref metav1.Object
}

// WorkloadStatus stores the observed state of the Workload.
type WorkloadStatus struct {
	Toleration   []corev1.Toleration
	NodeSelector map[string]string
}
