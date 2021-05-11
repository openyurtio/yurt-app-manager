/*
Copyright 2021 The OpenYurt Authors.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UnitedDaemonSetConditionType indicates valid conditions type of a UnitedDaemonSet.
type UnitedDaemonSetConditionType string

// UnitedDaemonSetSpec defines the desired state of UnitedDaemonSet.
type UnitedDaemonSetSpec struct {
	// Selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	Selector *metav1.LabelSelector `json:"selector"`

	// WorkloadTemplate describes the pool that will be created.
	// +optional
	WorkloadTemplate WorkloadTemplate `json:"workloadTemplate,omitempty"`

	// NodePoolSelector is a label query over nodepool that should match the replica count.
	// It must match the nodepool's labels.
	NodePoolSelector *metav1.LabelSelector `json:"nodepoolSelector"`

	// Indicates the number of histories to be conserved.
	// If unspecified, defaults to 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`
}

// UnitedDaemonSetStatus defines the observed state of UnitedDaemonSet.
type UnitedDaemonSetStatus struct {
	// ObservedGeneration is the most recent generation observed for this UnitedDaemonSet. It corresponds to the
	// UnitedDaemonSet's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Count of hash collisions for the UnitedDaemonSet. The UnitedDaemonSet controller
	// uses this field as a collision avoidance mechanism when it needs to
	// create the name for the newest ControllerRevision.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty"`

	// CurrentRevision, if not empty, indicates the current version of the UnitedDaemonSet.
	CurrentRevision string `json:"currentRevision"`

	// Represents the latest available observations of a UnitedDaemonSet's current state.
	// +optional
	Conditions []UnitedDaemonSetCondition `json:"conditions,omitempty"`

	// TemplateType indicates the type of PoolTemplate
	TemplateType TemplateType `json:"templateType"`
}

// UnitedDaemonSetCondition describes current state of a UnitedDaemonSet.
type UnitedDaemonSetCondition struct {
	// Type of in place set condition.
	Type UnitedDaemonSetConditionType `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status,omitempty"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=udd
// +kubebuilder:printcolumn:name="WorkloadTemplate",type="string",JSONPath=".status.templateType",description="The WorkloadTemplate Type."
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."

// UnitedDaemonSet is the Schema for the uniteddeployments API
type UnitedDaemonSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UnitedDaemonSetSpec   `json:"spec,omitempty"`
	Status UnitedDaemonSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UnitedDaemonSetList contains a list of UnitedDaemonSet
type UnitedDaemonSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UnitedDaemonSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UnitedDaemonSet{}, &UnitedDaemonSetList{})
}
