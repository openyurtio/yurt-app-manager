package yurtappdaemon

import (
	unitv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1helper "k8s.io/component-helpers/scheduling/corev1"

)

const updateRetries = 5

func IsTolerationsAllTaints(tolerations []corev1.Toleration, taints []corev1.Taint) bool {
	for i, _ := range taints {
		if !v1helper.TolerationsTolerateTaint(tolerations, &taints[i]) {
			return false
		}
	}
	return true
}

// NewYurtAppDaemonCondition creates a new YurtAppDaemon condition.
func NewYurtAppDaemonCondition(condType unitv1alpha1.YurtAppDaemonConditionType, status corev1.ConditionStatus, reason, message string) *unitv1alpha1.YurtAppDaemonCondition {
	return &unitv1alpha1.YurtAppDaemonCondition{
		Type:               condType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// GetYurtAppDaemonCondition returns the condition with the provided type.
func GetYurtAppDaemonCondition(status unitv1alpha1.YurtAppDaemonStatus, condType unitv1alpha1.YurtAppDaemonConditionType) *unitv1alpha1.YurtAppDaemonCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// SetYurtAppDaemonCondition updates the YurtAppDaemon to include the provided condition. If the condition that
// we are about to add already exists and has the same status, reason and message then we are not going to update.
func SetYurtAppDaemonCondition(status *unitv1alpha1.YurtAppDaemonStatus, condition *unitv1alpha1.YurtAppDaemonCondition) {
	currentCond := GetYurtAppDaemonCondition(*status, condition.Type)
	if currentCond != nil && currentCond.Status == condition.Status && currentCond.Reason == condition.Reason {
		return
	}

	if currentCond != nil && currentCond.Status == condition.Status {
		condition.LastTransitionTime = currentCond.LastTransitionTime
	}
	newConditions := filterOutCondition(status.Conditions, condition.Type)
	status.Conditions = append(newConditions, *condition)
}

func filterOutCondition(conditions []unitv1alpha1.YurtAppDaemonCondition, condType unitv1alpha1.YurtAppDaemonConditionType) []unitv1alpha1.YurtAppDaemonCondition {
	var newConditions []unitv1alpha1.YurtAppDaemonCondition
	for _, c := range conditions {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}
