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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// YurtIngressFinalizer is used to cleanup ingress resources when YurtIngress CR is deleted
const YurtIngressFinalizer string = "ingress.operator.openyurt.io"

type IngressNotReadyType string

const (
	IngressPending IngressNotReadyType = "Pending"
	IngressFailure IngressNotReadyType = "Failure"
)

// IngressPool defines the details of a Pool for ingress
type IngressPool struct {
	// Indicates the pool name.
	Name string `json:"name"`

	// IngressIPs is a list of IP addresses for which nodes will also accept traffic for this service.
	IngressIPs []string `json:"ingress_ips,omitempty"`
}

// IngressNotReadyConditionInfo defines the details info of an ingress not ready Pool
type IngressNotReadyConditionInfo struct {
	// Type of ingress not ready condition.
	Type IngressNotReadyType `json:"type,omitempty"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// IngressNotReadyPool defines the condition details of an ingress not ready Pool
type IngressNotReadyPool struct {
	// Indicates the base pool info.
	Pool IngressPool `json:"pool"`

	// Info of ingress not ready condition.
	Info *IngressNotReadyConditionInfo `json:"unreadyinfo,omitempty"`
}

// YurtIngressSpec defines the desired state of YurtIngress
type YurtIngressSpec struct {
	// Indicates the number of the ingress controllers to be deployed under all the specified nodepools.
	// +optional
	Replicas int32 `json:"ingress_controller_replicas_per_pool,omitempty"`

	// Indicates the ingress controller image url.
	// +optional
	IngressControllerImage string `json:"ingress_controller_image,omitempty"`

	// Indicates the ingress webhook image url.
	// +optional
	IngressWebhookCertGenImage string `json:"ingress_webhook_certgen_image,omitempty"`

	// Indicates all the nodepools on which to enable ingress.
	// +optional
	Pools []IngressPool `json:"pools,omitempty"`
}

// YurtIngressCondition describes current state of a YurtIngress
type YurtIngressCondition struct {
	// Indicates the pools that ingress controller is deployed successfully.
	IngressReadyPools []IngressPool `json:"ingressreadypools,omitempty"`

	// Indicates the pools that ingress controller is being deployed or deployed failed.
	IngressNotReadyPools []IngressNotReadyPool `json:"ingressunreadypools,omitempty"`
}

// YurtIngressStatus defines the observed state of YurtIngress
type YurtIngressStatus struct {
	// Indicates the number of the ingress controllers deployed under all the specified nodepools.
	// +optional
	Replicas int32 `json:"ingress_controller_replicas_per_pool,omitempty"`

	// Indicates all the nodepools on which to enable ingress.
	// +optional
	Conditions YurtIngressCondition `json:"conditions,omitempty"`

	// Indicates the ingress controller image url.
	// +optional
	IngressControllerImage string `json:"ingress_controller_image"`

	// Indicates the ingress webhook image url.
	// +optional
	IngressWebhookCertGenImage string `json:"ingress_webhook_certgen_image"`

	// Total number of ready pools on which ingress is enabled.
	// +optional
	ReadyNum int32 `json:"readyNum"`

	// Total number of unready pools on which ingress is enabling or enable failed.
	// +optional
	UnreadyNum int32 `json:"unreadyNum"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,path=yurtingresses,shortName=ying,categories=all
// +kubebuilder:printcolumn:name="Replicas-Per-Pool",type="integer",JSONPath=".status.ingress_controller_replicas_per_pool",description="The nginx ingress controller replicas per pool"
// +kubebuilder:printcolumn:name="ReadyNum",type="integer",JSONPath=".status.readyNum",description="The number of pools on which ingress is enabled"
// +kubebuilder:printcolumn:name="NotReadyNum",type="integer",JSONPath=".status.unreadyNum",description="The number of pools on which ingress is enabling or enable failed"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +genclient:nonNamespaced
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

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
