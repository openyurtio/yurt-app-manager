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

package v1alpha2

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// YurtIngressFinalizer is used to cleanup ingress resources when YurtIngress CR is deleted
const YurtIngressFinalizer string = "ingress.operator.openyurt.io"

// YurtIngressSpec defines the desired state of YurtIngress
type YurtIngressSpec struct {
	Enabled bool `json:"enabled,omitempty"`

	Values *apiextensionsv1.JSON `json:"values,omitempty"`
}

// YurtIngressStatus defines the observed state of YurtIngress
type YurtIngressStatus struct {

	// +optional
	Status string `json:"status,omitempty"`

	// Conditions holds the conditions for the HelmRelease.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastAppliedRevision is the revision of the last successfully applied source.
	// +optional
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`
}

func (in *YurtIngress) GetStatusConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,path=yurtingresses,shortName=ying,categories=all
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +genclient:nonNamespaced
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion

// YurtIngress is the Schema for the yurtingresses API
type YurtIngress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   YurtIngressSpec   `json:"spec,omitempty"`
	Status YurtIngressStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// YurtIngressList contains a list of YurtIngress
type YurtIngressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []YurtIngress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&YurtIngress{}, &YurtIngressList{})
}
