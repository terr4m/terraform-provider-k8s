package k8sutils

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceEmptyOrExists checks if the input namespace is empty or exists.
func NamespaceEmptyOrExists(ctx context.Context, client *Client, namespace string) (bool, error) {
	if len(namespace) == 0 {
		return true, nil
	}

	nsri, err := client.ResourceInterface("v1", "Namespace", "", false)
	if err != nil {
		return false, fmt.Errorf("failed to get namespace resource interface: %w", err)
	}

	_, err = nsri.Get(ctx, namespace, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to get namespace %q: %w", namespace, err)
	}

	return true, nil
}
