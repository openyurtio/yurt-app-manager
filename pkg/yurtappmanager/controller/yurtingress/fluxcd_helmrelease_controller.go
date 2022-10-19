/*
Copyright 2020 The Flux authors

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

package yurtingress

import (
	"context"
	"errors"
	"fmt"

	v2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha2"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/fluxcd"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/fluxcd/events"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/fluxcd/runner"
	util2 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/fluxcd/util"
)

func (r *YurtIngressReconciler) reconcileRelease(ctx context.Context,
	hr v2.HelmRelease, chart *chart.Chart, values chartutil.Values) (v2.HelmRelease, error) {
	log := ctrl.LoggerFrom(ctx)

	// Initialize Helm action runner
	//getter, err := r.buildRESTClientGetter(ctx, hr)
	//if err != nil {
	//	return v2.HelmReleaseNotReady(hr, v2.InitFailedReason, err.Error()), err
	//}
	//run, err := runner.NewRunner(getter, hr.GetStorageNamespace(), log)
	//if err != nil {
	//	return v2.HelmReleaseNotReady(hr, v2.InitFailedReason, "failed to initialize Helm action runner"), err
	//}
	run, err := fluxcd.NewDefaultRunner(ctx, r.namespace)
	if err != nil {
		return v2.HelmReleaseNotReady(hr, v2.InitFailedReason, "failed to initialize Helm action runner"), err
	}

	// Determine last release revision.
	rel, observeLastReleaseErr := run.ObserveLastRelease(hr)
	if observeLastReleaseErr != nil {
		err = fmt.Errorf("failed to get last release revision: %w", observeLastReleaseErr)
		return v2.HelmReleaseNotReady(hr, v2.GetLastReleaseFailedReason, "failed to get last release revision"), err
	}

	// Register the current release attempt.
	revision := chart.Metadata.Version
	releaseRevision := util2.ReleaseRevision(rel)
	valuesChecksum := util2.ValuesChecksum(values)
	hr, hasNewState := v2.HelmReleaseAttempted(hr, revision, releaseRevision, valuesChecksum)
	if hasNewState {
		hr = v2.HelmReleaseProgressing(hr)
		if updateStatusErr := r.patchStatus(ctx, &hr); updateStatusErr != nil {
			log.Error(updateStatusErr, "unable to update status after state update")
			return hr, updateStatusErr
		}
		// Record progressing status
		r.recordReadiness(ctx, hr)
	}

	// Check status of any previous release attempt.
	released := apimeta.FindStatusCondition(hr.Status.Conditions, v2.ReleasedCondition)
	if released != nil {
		switch released.Status {
		// Succeed if the previous release attempt succeeded.
		case metav1.ConditionTrue:
			return v2.HelmReleaseReady(hr), nil
		case metav1.ConditionFalse:
			// Fail if the previous release attempt remediation failed.
			remediated := apimeta.FindStatusCondition(hr.Status.Conditions, v2.RemediatedCondition)
			if remediated != nil && remediated.Status == metav1.ConditionFalse {
				err = fmt.Errorf("previous release attempt remediation failed")
				return v2.HelmReleaseNotReady(hr, remediated.Reason, remediated.Message), err
			}
		}

		// Fail if install retries are exhausted.
		if hr.Spec.GetInstall().GetRemediation().RetriesExhausted(hr) {
			err = fmt.Errorf("install retries exhausted")
			return v2.HelmReleaseNotReady(hr, released.Reason, err.Error()), err
		}

		// Fail if there is a release and upgrade retries are exhausted.
		// This avoids failing after an upgrade uninstall remediation strategy.
		if rel != nil && hr.Spec.GetUpgrade().GetRemediation().RetriesExhausted(hr) {
			err = fmt.Errorf("upgrade retries exhausted")
			return v2.HelmReleaseNotReady(hr, released.Reason, err.Error()), err
		}
	}

	// Deploy the release.
	var deployAction v2.DeploymentAction
	if rel == nil {
		r.event(ctx, hr, revision, events.EventSeverityInfo, "Helm install has started")
		deployAction = hr.Spec.GetInstall()
		rel, err = run.Install(hr, chart, values)
		err = r.handleHelmActionResult(ctx, &hr, revision, err, deployAction.GetDescription(),
			v2.ReleasedCondition, v2.InstallSucceededReason, v2.InstallFailedReason)
	} else {
		r.event(ctx, hr, revision, events.EventSeverityInfo, "Helm upgrade has started")
		deployAction = hr.Spec.GetUpgrade()
		rel, err = run.Upgrade(hr, chart, values)
		err = r.handleHelmActionResult(ctx, &hr, revision, err, deployAction.GetDescription(),
			v2.ReleasedCondition, v2.UpgradeSucceededReason, v2.UpgradeFailedReason)
	}
	remediation := deployAction.GetRemediation()

	// If there is a new release revision...
	if util2.ReleaseRevision(rel) > releaseRevision {
		// Ensure release is not marked remediated.
		apimeta.RemoveStatusCondition(&hr.Status.Conditions, v2.RemediatedCondition)

		// If new release revision is successful and tests are enabled, run them.
		if err == nil && hr.Spec.GetTest().Enable {
			_, testErr := run.Test(hr)
			testErr = r.handleHelmActionResult(ctx, &hr, revision, testErr, "test",
				v2.TestSuccessCondition, v2.TestSucceededReason, v2.TestFailedReason)

			// Propagate any test error if not marked ignored.
			if testErr != nil && !remediation.MustIgnoreTestFailures(hr.Spec.GetTest().IgnoreFailures) {
				testsPassing := apimeta.FindStatusCondition(hr.Status.Conditions, v2.TestSuccessCondition)
				newCondition := metav1.Condition{
					Type:    v2.ReleasedCondition,
					Status:  metav1.ConditionFalse,
					Reason:  testsPassing.Reason,
					Message: testsPassing.Message,
				}
				apimeta.SetStatusCondition(hr.GetStatusConditions(), newCondition)
				err = testErr
			}
		}
	}

	if err != nil {
		// Increment failure count for deployment action.
		remediation.IncrementFailureCount(&hr)
		// Remediate deployment failure if necessary.
		if !remediation.RetriesExhausted(hr) || remediation.MustRemediateLastFailure() {
			if util2.ReleaseRevision(rel) <= releaseRevision {
				log.Info(fmt.Sprintf("skipping remediation, no new release revision created"))
			} else {
				var remediationErr error
				switch remediation.GetStrategy() {
				case v2.RollbackRemediationStrategy:
					rollbackErr := run.Rollback(hr)
					remediationErr = r.handleHelmActionResult(ctx, &hr, revision, rollbackErr, "rollback",
						v2.RemediatedCondition, v2.RollbackSucceededReason, v2.RollbackFailedReason)
				case v2.UninstallRemediationStrategy:
					uninstallErr := run.Uninstall(hr)
					remediationErr = r.handleHelmActionResult(ctx, &hr, revision, uninstallErr, "uninstall",
						v2.RemediatedCondition, v2.UninstallSucceededReason, v2.UninstallFailedReason)
				}
				if remediationErr != nil {
					err = remediationErr
				}
			}

			// Determine release after remediation.
			rel, observeLastReleaseErr = run.ObserveLastRelease(hr)
			if observeLastReleaseErr != nil {
				err = &ConditionError{
					Reason: v2.GetLastReleaseFailedReason,
					Err:    errors.New("failed to get last release revision after remediation"),
				}
			}
		}
	}

	hr.Status.LastReleaseRevision = util2.ReleaseRevision(rel)

	if err != nil {
		reason := v2.ReconciliationFailedReason
		if condErr := (*ConditionError)(nil); errors.As(err, &condErr) {
			reason = condErr.Reason
		}
		return v2.HelmReleaseNotReady(hr, reason, err.Error()), err
	}
	return v2.HelmReleaseReady(hr), nil
}

func (r *YurtIngressReconciler) handleHelmActionResult(ctx context.Context,
	hr *v2.HelmRelease, revision string, err error, action string, condition string, succeededReason string, failedReason string) error {
	if err != nil {
		err = fmt.Errorf("Helm %s failed: %w", action, err)
		msg := err.Error()
		if actionErr := (*runner.ActionError)(nil); errors.As(err, &actionErr) {
			msg = msg + "\n\nLast Helm logs:\n\n" + actionErr.CapturedLogs
		}
		newCondition := metav1.Condition{
			Type:    condition,
			Status:  metav1.ConditionFalse,
			Reason:  failedReason,
			Message: msg,
		}
		apimeta.SetStatusCondition(hr.GetStatusConditions(), newCondition)
		r.event(ctx, *hr, revision, events.EventSeverityError, msg)
		return &ConditionError{Reason: failedReason, Err: err}
	} else {
		msg := fmt.Sprintf("Helm %s succeeded", action)
		newCondition := metav1.Condition{
			Type:    condition,
			Status:  metav1.ConditionTrue,
			Reason:  succeededReason,
			Message: msg,
		}
		apimeta.SetStatusCondition(hr.GetStatusConditions(), newCondition)
		r.event(ctx, *hr, revision, events.EventSeverityInfo, msg)
		return nil
	}
}

func (r *YurtIngressReconciler) event(_ context.Context, hr v2.HelmRelease, revision, severity, msg string) {

	eventtype := "Normal"
	if severity == events.EventSeverityError {
		eventtype = "Warning"
	}

	yi := v1alpha2.YurtIngress{
		ObjectMeta: hr.ObjectMeta,
	}

	r.EventRecorder.Eventf(&yi, eventtype, severity, msg)
}

func (r *YurtIngressReconciler) recordReadiness(ctx context.Context, hr v2.HelmRelease) {

}

// ConditionError represents an error with a status condition reason attached.
type ConditionError struct {
	Reason string
	Err    error
}

func (c ConditionError) Error() string {
	return c.Err.Error()
}
