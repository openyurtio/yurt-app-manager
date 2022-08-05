/*
Copyright 2021 The OpenYurt Authors.
Copyright 2020 The Kruise Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

@CHANGELOG
OpenYurt Authors:
change UnitedDeployment API Definition
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UnitedDeploymentConditionType indicates valid conditions type of a UnitedDeployment.
type UnitedDeploymentConditionType string

const (
	// PoolProvisioned means all the expected pools are provisioned and unexpected pools are deleted.
	UnitedPoolProvisioned UnitedDeploymentConditionType = "PoolProvisioned"
	// PoolUpdated means all the pools are updated.
	UnitedPoolUpdated UnitedDeploymentConditionType = "PoolUpdated"
	// PoolFailure is added to a YurtAppSet when one of its pools has failure during its own reconciling.
	UnitedPoolFailure UnitedDeploymentConditionType = "PoolFailure"
)

// UnitedDeploymentSpec defines the desired state of UnitedDeployment.
type UnitedDeploymentSpec struct {
	// Selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	Selector *metav1.LabelSelector `json:"selector"`

	// WorkloadTemplate describes the pool that will be created.
	// +optional
	WorkloadTemplate WorkloadTemplate `json:"workloadTemplate,omitempty"`

	// Topology describes the pods distribution detail between each of pools.
	// +optional
	Topology Topology `json:"topology,omitempty"`

	// Indicates the number of histories to be conserved.
	// If unspecified, defaults to 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`
}

// UnitedDeploymentStatus defines the observed state of UnitedDeployment.
type UnitedDeploymentStatus struct {
	// ObservedGeneration is the most recent generation observed for this UnitedDeployment. It corresponds to the
	// UnitedDeployment's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Count of hash collisions for the UnitedDeployment. The UnitedDeployment controller
	// uses this field as a collision avoidance mechanism when it needs to
	// create the name for the newest ControllerRevision.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty"`

	// CurrentRevision, if not empty, indicates the current version of the UnitedDeployment.
	CurrentRevision string `json:"currentRevision"`

	// Represents the latest available observations of a UnitedDeployment's current state.
	// +optional
	Conditions []UnitedDeploymentCondition `json:"conditions,omitempty"`

	// Records the topology detail information of the replicas of each pool.
	// +optional
	PoolReplicas map[string]int32 `json:"poolReplicas,omitempty"`

	// The number of ready replicas.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas"`

	// Replicas is the most recently observed number of replicas.
	Replicas int32 `json:"replicas"`

	// TemplateType indicates the type of PoolTemplate
	TemplateType TemplateType `json:"templateType"`
}

// UnitedDeploymentCondition describes current state of a UnitedDeployment.
type UnitedDeploymentCondition struct {
	// Type of in place set condition.
	Type UnitedDeploymentConditionType `json:"type,omitempty"`

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
// +kubebuilder:resource:shortName=ud
// +kubebuilder:printcolumn:name="READY",type="integer",JSONPath=".status.readyReplicas",description="The number of pods ready."
// +kubebuilder:printcolumn:name="WorkloadTemplate",type="string",JSONPath=".status.templateType",description="The WorkloadTemplate Type."
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."

// UnitedDeployment is the Schema for the uniteddeployments API
type UnitedDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UnitedDeploymentSpec   `json:"spec,omitempty"`
	Status UnitedDeploymentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UnitedDeploymentList contains a list of UnitedDeployment
type UnitedDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UnitedDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UnitedDeployment{}, &UnitedDeploymentList{})
}
