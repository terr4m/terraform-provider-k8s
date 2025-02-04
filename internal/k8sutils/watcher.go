package k8sutils

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

// GetWatcher returns a watcher for the given resource.
func GetWatcher(ctx context.Context, ri dynamic.ResourceInterface, name string, fieldSelectors ...string) (watch.Interface, error) {
	selector := fields.OneTermEqualSelector("metadata.name", name)

	if len(fieldSelectors) > 0 {
		fsls := make([]fields.Selector, len(fieldSelectors)+1)
		fsls[0] = selector

		for i, s := range fieldSelectors {
			fsl, err := fields.ParseSelector(s)
			if err != nil {
				return nil, err
			}

			fsls[i+1] = fsl
		}

		selector = fields.AndSelectors(fsls...)
	}

	return ri.Watch(ctx, metav1.ListOptions{Watch: true, FieldSelector: selector.String()})
}

// WatchForAddedModified watches the given resource being created or modified matches the conditions.
func WatchForAddedModified(ctx context.Context, w watch.Interface, conditions map[string]string, rollout bool, rolloutKind string) (*unstructured.Unstructured, error) {
	conditionsMet := len(conditions) == 0
	rolloutComplete := !rollout

	for {
		select {
		case event := <-w.ResultChan():
			if event.Type == watch.Added || event.Type == watch.Modified {
				u, ok := event.Object.(*unstructured.Unstructured)
				if !ok {
					return nil, fmt.Errorf("expected *unstructured.Unstructured, got %T", event.Object)
				}

				if len(conditions) > 0 {
					success, err := checkConditions(u, conditions)
					if err != nil {
						return nil, err
					}

					conditionsMet = success
				}

				if rollout {
					success, err := checkRolloutComplete(u, rolloutKind)
					if err != nil {
						return nil, err
					}

					rolloutComplete = success
				}

				if conditionsMet && rolloutComplete {
					return u, nil
				}
			}

		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for resource: %w", ctx.Err())
		}
	}
}

// WatchForDelete watches the given resource until it is deleted.
func WatchForDelete(ctx context.Context, w watch.Interface) error {
	for {
		select {
		case event := <-w.ResultChan():
			if event.Type == watch.Deleted {
				return nil
			}

		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for resource to be deleted: %w", ctx.Err())
		}
	}
}

// checkConditions checks if all of the given conditions are met.
func checkConditions(u *unstructured.Unstructured, conditions map[string]string) (bool, error) {
	required := len(conditions)
	actual := 0

	condSlice, ok, err := unstructured.NestedSlice(u.Object, "status", "conditions")
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	for _, c := range condSlice {
		cond, ok := c.(map[string]any)
		if !ok {
			return false, fmt.Errorf("condition wrong type")
		}

		if cond["type"] == nil || cond["status"] == nil {
			continue
		}

		condType, ok := cond["type"].(string)
		if !ok {
			return false, fmt.Errorf("condition type field not a string")
		}

		condStatus, ok := cond["status"].(string)
		if !ok {
			return false, fmt.Errorf("condition status field not a string")
		}

		condWanted, ok := conditions[condType]
		if !ok {
			continue
		}

		if condWanted == condStatus {
			actual++
		}
	}

	return actual == required, nil
}
