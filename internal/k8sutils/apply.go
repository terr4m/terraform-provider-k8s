package k8sutils

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type ResourceClient interface {
	ResourceInterface(gvk *schema.GroupVersionKind, namespace string, requireNamespace bool) (dynamic.ResourceInterface, error)
}

type FieldManager struct {
	Name           string
	ForceConflicts bool
}

type ReadResult struct {
	Found       bool
	Fingerprint string
}

type ApplyManager struct {
	client       ResourceClient
	fieldManager FieldManager
}

func NewApplyManager(client ResourceClient, fieldManager FieldManager) *ApplyManager {
	return &ApplyManager{client: client, fieldManager: fieldManager}
}

func (m *ApplyManager) Create(ctx context.Context, manifest map[string]any, override *FieldManager) (string, error) {
	prepared, err := m.prepare(manifest, override)
	if err != nil {
		return "", err
	}

	u := &unstructured.Unstructured{Object: prepared.Manifest}
	uu, err := prepared.Resource.Apply(ctx, prepared.Name, u, prepared.ApplyOptions)
	if err != nil {
		return "", err
	}

	return ObservedManifestFingerprint(prepared.Manifest, uu.Object)
}

func (m *ApplyManager) Read(ctx context.Context, manifest map[string]any) (ReadResult, error) {
	prepared, err := m.prepare(manifest, nil)
	if err != nil {
		return ReadResult{}, err
	}

	uu, err := prepared.Resource.Get(ctx, prepared.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return ReadResult{Found: false}, nil
	}
	if err != nil {
		return ReadResult{}, err
	}

	fingerprint, err := ObservedManifestFingerprint(prepared.Manifest, uu.Object)
	if err != nil {
		return ReadResult{}, err
	}

	return ReadResult{Found: true, Fingerprint: fingerprint}, nil
}

func (m *ApplyManager) Update(ctx context.Context, manifest map[string]any, override *FieldManager) (string, error) {
	return m.Create(ctx, manifest, override)
}

func (m *ApplyManager) Delete(ctx context.Context, manifest map[string]any) error {
	prepared, err := m.prepare(manifest, nil)
	if err != nil {
		return err
	}

	err = prepared.Resource.Delete(ctx, prepared.Name, metav1.DeleteOptions{})
	if errors.IsNotFound(err) || errors.IsGone(err) {
		return nil
	}

	return err
}

type preparedApply struct {
	Manifest     map[string]any
	Name         string
	Resource     dynamic.ResourceInterface
	ApplyOptions metav1.ApplyOptions
}

func (m *ApplyManager) prepare(manifest map[string]any, override *FieldManager) (*preparedApply, error) {
	gvk, namespace, name, err := ManifestIdentity(manifest)
	if err != nil {
		return nil, err
	}

	resourceInterface, err := m.client.ResourceInterface(gvk, namespace, true)
	if err != nil {
		return nil, err
	}

	return &preparedApply{
		Manifest:     manifest,
		Name:         name,
		Resource:     resourceInterface,
		ApplyOptions: BuildApplyOptions(m.fieldManager, override),
	}, nil
}
