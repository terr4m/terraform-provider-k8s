package k8sutils

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type applyClientStub struct {
	resource dynamic.ResourceInterface
	err      error
	gvk      *schema.GroupVersionKind
	ns       string
	require  bool
}

func (c *applyClientStub) ResourceInterface(gvk *schema.GroupVersionKind, namespace string, requireNamespace bool) (dynamic.ResourceInterface, error) {
	c.gvk = gvk
	c.ns = namespace
	c.require = requireNamespace
	return c.resource, c.err
}

type applyResourceInterfaceStub struct {
	dynamic.ResourceInterface
	applyResult  *unstructured.Unstructured
	applyErr     error
	getResult    *unstructured.Unstructured
	getErr       error
	deleteErr    error
	applyOptions metav1.ApplyOptions
	name         string
}

func (r *applyResourceInterfaceStub) Apply(_ context.Context, name string, _ *unstructured.Unstructured, opts metav1.ApplyOptions, _ ...string) (*unstructured.Unstructured, error) {
	r.name = name
	r.applyOptions = opts
	return r.applyResult, r.applyErr
}

func (r *applyResourceInterfaceStub) Get(_ context.Context, name string, _ metav1.GetOptions, _ ...string) (*unstructured.Unstructured, error) {
	r.name = name
	return r.getResult, r.getErr
}

func (r *applyResourceInterfaceStub) Delete(_ context.Context, name string, _ metav1.DeleteOptions, _ ...string) error {
	r.name = name
	return r.deleteErr
}

func TestApplyManagerLifecycle(t *testing.T) {
	t.Parallel()

	manifest := map[string]any{"apiVersion": "v1", "kind": "ConfigMap", "metadata": map[string]any{"name": "example", "namespace": "default"}, "data": map[string]any{"foo": "bar"}}
	live := &unstructured.Unstructured{Object: map[string]any{"apiVersion": "v1", "kind": "ConfigMap", "metadata": map[string]any{"name": "example", "namespace": "default", "uid": "abc"}, "data": map[string]any{"foo": "bar"}}}

	t.Run("create update read delete", func(t *testing.T) {
		resource := &applyResourceInterfaceStub{applyResult: live, getResult: live}
		client := &applyClientStub{resource: resource}
		manager := NewApplyManager(client, FieldManager{Name: "provider"})

		fingerprint, err := manager.Create(context.Background(), manifest, &FieldManager{Name: "resource", ForceConflicts: true})
		if err != nil || fingerprint == "" {
			t.Fatalf("Create() = (%q, %v), want non-empty fingerprint and nil error", fingerprint, err)
		}
		if resource.applyOptions.FieldManager != "resource" || !resource.applyOptions.Force {
			t.Fatalf("Create() apply options = %#v, want override applied", resource.applyOptions)
		}

		readResult, err := manager.Read(context.Background(), manifest)
		if err != nil || !readResult.Found || readResult.Fingerprint == "" {
			t.Fatalf("Read() = (%#v, %v), want found fingerprint", readResult, err)
		}

		fingerprint, err = manager.Update(context.Background(), manifest, nil)
		if err != nil || fingerprint == "" {
			t.Fatalf("Update() = (%q, %v), want non-empty fingerprint and nil error", fingerprint, err)
		}

		if err := manager.Delete(context.Background(), manifest); err != nil {
			t.Fatalf("Delete() error = %v, want nil", err)
		}
	})

	t.Run("read not found", func(t *testing.T) {
		resource := &applyResourceInterfaceStub{getErr: errors.NewNotFound(schema.GroupResource{}, "example")}
		manager := NewApplyManager(&applyClientStub{resource: resource}, FieldManager{Name: "provider"})

		readResult, err := manager.Read(context.Background(), manifest)
		if err != nil || readResult.Found {
			t.Fatalf("Read() = (%#v, %v), want not found and nil error", readResult, err)
		}
	})
}
