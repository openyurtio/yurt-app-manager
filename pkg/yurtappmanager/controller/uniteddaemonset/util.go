package uniteddaemonset

import (
	unitv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/core/v1/helper"
)

const updateRetries = 5

func IsTolerationsAllTaints(tolerations []corev1.Toleration, taints []corev1.Taint) bool {
	for i, _ := range taints {
		if !helper.TolerationsTolerateTaint(tolerations, &taints[i]) {
			return false
		}
	}
	return true
}

// NewUnitedDaemonSetCondition creates a new UnitedDaemonSet condition.
func NewUnitedDaemonSetCondition(condType unitv1alpha1.UnitedDaemonSetConditionType, status corev1.ConditionStatus, reason, message string) *unitv1alpha1.UnitedDaemonSetCondition {
	return &unitv1alpha1.UnitedDaemonSetCondition{
		Type:               condType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// GetUnitedDaemonSetCondition returns the condition with the provided type.
func GetUnitedDaemonSetCondition(status unitv1alpha1.UnitedDaemonSetStatus, condType unitv1alpha1.UnitedDaemonSetConditionType) *unitv1alpha1.UnitedDaemonSetCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// SetUnitedDaemonSetCondition updates the UnitedDaemonSet to include the provided condition. If the condition that
// we are about to add already exists and has the same status, reason and message then we are not going to update.
func SetUnitedDaemonSetCondition(status *unitv1alpha1.UnitedDaemonSetStatus, condition *unitv1alpha1.UnitedDaemonSetCondition) {
	currentCond := GetUnitedDaemonSetCondition(*status, condition.Type)
	if currentCond != nil && currentCond.Status == condition.Status && currentCond.Reason == condition.Reason {
		return
	}

	if currentCond != nil && currentCond.Status == condition.Status {
		condition.LastTransitionTime = currentCond.LastTransitionTime
	}
	newConditions := filterOutCondition(status.Conditions, condition.Type)
	status.Conditions = append(newConditions, *condition)
}

// RemoveUnitedDaemonSetCondition removes the UnitedDaemonSet condition with the provided type.
func RemoveUnitedDaemonSetCondition(status *unitv1alpha1.UnitedDaemonSetStatus, condType unitv1alpha1.UnitedDaemonSetConditionType) {
	status.Conditions = filterOutCondition(status.Conditions, condType)
}

func filterOutCondition(conditions []unitv1alpha1.UnitedDaemonSetCondition, condType unitv1alpha1.UnitedDaemonSetConditionType) []unitv1alpha1.UnitedDaemonSetCondition {
	var newConditions []unitv1alpha1.UnitedDaemonSetCondition
	for _, c := range conditions {
		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}
