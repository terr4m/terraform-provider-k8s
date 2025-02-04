package k8sutils

import (
	"fmt"

	apps_v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// TimedOutReason is the reason for a Deployment being considered failed due to a timeout.
	//
	// This is borrowed from: https://github.com/kubernetes/kubernetes/blob/2b3da7dfc846fec7c4044a320f8f38b4a45367a3/pkg/controller/deployment/util/deployment_util.go#L86
	TimedOutReason = "ProgressDeadlineExceeded"
)

// RolloutKinds returns a list of kinds that support rollouts.
func RolloutKinds() []string {
	return []string{"DaemonSet", "Deployment", "StatefulSet"}
}

// checkRolloutComplete checks if the rollout is complete for the given object.
func checkRolloutComplete(u *unstructured.Unstructured, kind string) (bool, error) {
	switch {
	case kind == "DaemonSet":
		return checkDaemonSetRolloutComplete(u)
	case kind == "Deployment":
		return checkDeploymentRolloutComplete(u)
	case kind == "StatefulSet":
		return checkStatefulSetRolloutComplete(u)
	default:
		return false, fmt.Errorf("unsupported rollout kind %q", kind)
	}
}

// checkDaemonSetRolloutComplete checks if the rollout is complete for the given DaemonSet.
//
// This is borrowed from: https://github.com/kubernetes/kubectl/blob/c4be63c54b7188502c1a63bb884a0b05fac51ebd/pkg/util/deployment/deployment.go#L60
func checkDaemonSetRolloutComplete(u *unstructured.Unstructured) (bool, error) {
	var ds *apps_v1.DaemonSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &ds)
	if err != nil {
		return false, err
	}

	if ds.Spec.UpdateStrategy.Type != apps_v1.RollingUpdateDaemonSetStrategyType {
		return true, nil
	}

	if ds.Generation <= ds.Status.ObservedGeneration {
		if ds.Status.UpdatedNumberScheduled < ds.Status.DesiredNumberScheduled {
			return false, nil
		}

		if ds.Status.NumberAvailable < ds.Status.DesiredNumberScheduled {
			return false, nil
		}

		return true, nil
	}

	return false, nil
}

// checkDeploymentRolloutComplete checks if the rollout is complete for the given Deployment.
//
// This is borrowed from: https://github.com/kubernetes/kubectl/blob/c4be63c54b7188502c1a63bb884a0b05fac51ebd/pkg/polymorphichelpers/rollout_status.go#L59
func checkDeploymentRolloutComplete(u *unstructured.Unstructured) (bool, error) {
	var d *apps_v1.Deployment
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &d)
	if err != nil {
		return false, err
	}

	if d.Generation <= d.Status.ObservedGeneration {
		condition := getDeploymentCondition(d.Status, apps_v1.DeploymentProgressing)
		if condition != nil && condition.Reason == TimedOutReason {
			return false, nil
		}

		if d.Spec.Replicas != nil && d.Status.UpdatedReplicas < *d.Spec.Replicas {
			return false, nil
		}

		if d.Status.Replicas > d.Status.UpdatedReplicas {
			return false, nil
		}

		if d.Status.AvailableReplicas < d.Status.UpdatedReplicas {
			return false, nil
		}

		return true, nil
	}

	return false, nil
}

// checkStatefulSetRolloutComplete checks if the rollout is complete for the given StatefulSet.
//
// This is borrowed from: https://github.com/kubernetes/kubectl/blob/c4be63c54b7188502c1a63bb884a0b05fac51ebd/pkg/polymorphichelpers/rollout_status.go#L120
func checkStatefulSetRolloutComplete(u *unstructured.Unstructured) (bool, error) {
	var sts *apps_v1.StatefulSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &sts)
	if err != nil {
		return false, err
	}

	if sts.Spec.UpdateStrategy.Type != apps_v1.RollingUpdateStatefulSetStrategyType {
		return true, nil
	}

	if sts.Status.ObservedGeneration == 0 || sts.Generation > sts.Status.ObservedGeneration {
		return false, nil
	}

	if sts.Spec.Replicas != nil && sts.Status.ReadyReplicas < *sts.Spec.Replicas {
		return false, nil
	}

	if sts.Spec.UpdateStrategy.Type == apps_v1.RollingUpdateStatefulSetStrategyType && sts.Spec.UpdateStrategy.RollingUpdate != nil {
		if sts.Spec.Replicas != nil && sts.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
			if sts.Status.UpdatedReplicas < (*sts.Spec.Replicas - *sts.Spec.UpdateStrategy.RollingUpdate.Partition) {
				return false, nil
			}
		}

		return true, nil
	}

	if sts.Status.UpdateRevision != sts.Status.CurrentRevision {
		return false, nil
	}

	return true, nil
}

// getDeploymentCondition returns the condition with the given type from the given Deployment status.
//
// This is borrowed from: https://github.com/kubernetes/kubectl/blob/c4be63c54b7188502c1a63bb884a0b05fac51ebd/pkg/util/deployment/deployment.go#L60
func getDeploymentCondition(status apps_v1.DeploymentStatus, condType apps_v1.DeploymentConditionType) *apps_v1.DeploymentCondition {
	for _, c := range status.Conditions {
		if c.Type == condType {
			return &c
		}
	}
	return nil
}
