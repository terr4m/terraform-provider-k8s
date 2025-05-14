package provider

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"k8s.io/client-go/rest"
	aggregator "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
)

func TestNewK8sProviderClient(t *testing.T) {
	t.Parallel()

	t.Run("set_rest_config", func(t *testing.T) {
		restConfig := rest.Config{
			Host: "https://example.com",
		}

		got := NewK8sProviderClient(&restConfig)

		if diff := cmp.Diff(&restConfig, got.restConfig); diff != "" {
			t.Errorf(
				"NewK8sProviderClient rest config:\n%v\nwant:\n%v\ndiff:\n%v",
				got,
				restConfig,
				diff,
			)
		}
	})
}

func TestK8sProviderClient(t *testing.T) {
	t.Parallel()

	t.Run("AggregatorClient", func(t *testing.T) {
		t.Parallel()

		for _, d := range []struct {
			testName   string
			mockSetup  func() K8sProviderClient
			restConfig *rest.Config
			errMsg     string
		}{
			{
				testName:  "rest_config_nil",
				mockSetup: func() K8sProviderClient { return K8sProviderClient{} },
				errMsg:    "rest config is required",
			},
			{
				testName: "aggregator_client_cached",
				mockSetup: func() K8sProviderClient {
					return K8sProviderClient{
						restConfig:       &rest.Config{},
						aggregatorClient: &aggregator.Clientset{},
					}
				},
			},
			{
				testName: "new_aggregator_client",
				mockSetup: func() K8sProviderClient {
					return K8sProviderClient{
						restConfig: &rest.Config{
							Host: "https://example.com",
						},
					}
				},
			},
		} {
			t.Run(d.testName, func(t *testing.T) {
				t.Parallel()

				client := d.mockSetup()

				got, err := client.AggregatorClient()

				if len(d.errMsg) == 0 && got == nil {
					t.Errorf("K8sProviderClient.AggregatorClient returned nil, want non-nil")
				}

				var errMsg string
				if err != nil {
					errMsg = err.Error()
				}

				if errMsg != d.errMsg {
					t.Errorf("K8sProviderClient.AggregatorClient returned error message %q, want %q", errMsg, d.errMsg)
				}
			})
		}
	})

	t.Run("DiscoveryClient", func(t *testing.T) {
		t.Parallel()

		for _, d := range []struct {
			testName   string
			mockSetup  func() K8sProviderClient
			restConfig *rest.Config
			errMsg     string
		}{
			{
				testName:  "rest_config_nil",
				mockSetup: func() K8sProviderClient { return K8sProviderClient{} },
				errMsg:    "rest config is required",
			},
			{
				testName: "discovery_client_cached",
				mockSetup: func() K8sProviderClient {
					return K8sProviderClient{
						restConfig:      &rest.Config{},
						discoveryClient: &discoveryClientStub{},
					}
				},
			},
			{
				testName: "new_discovery_client",
				mockSetup: func() K8sProviderClient {
					return K8sProviderClient{
						restConfig: &rest.Config{
							Host: "https://example.com",
						},
					}
				},
			},
		} {
			t.Run(d.testName, func(t *testing.T) {
				t.Parallel()

				client := d.mockSetup()

				got, err := client.DiscoveryClient()

				if len(d.errMsg) == 0 && got == nil {
					t.Errorf("K8sProviderClient.DiscoveryClient returned nil, want non-nil")
				}

				var errMsg string
				if err != nil {
					errMsg = err.Error()
				}

				if errMsg != d.errMsg {
					t.Errorf("K8sProviderClient.DiscoveryClient returned error message %q, want %q", errMsg, d.errMsg)
				}
			})
		}
	})

	t.Run("DynamicClient", func(t *testing.T) {
		t.Parallel()

		for _, d := range []struct {
			testName   string
			mockSetup  func() K8sProviderClient
			restConfig *rest.Config
			errMsg     string
		}{
			{
				testName:  "rest_config_nil",
				mockSetup: func() K8sProviderClient { return K8sProviderClient{} },
				errMsg:    "rest config is required",
			},
			{
				testName: "dynamic_client_cached",
				mockSetup: func() K8sProviderClient {
					return K8sProviderClient{
						restConfig:    &rest.Config{},
						dynamicClient: &dynamicClientStub{},
					}
				},
			},
			{
				testName: "new_dynamic_client",
				mockSetup: func() K8sProviderClient {
					return K8sProviderClient{
						restConfig: &rest.Config{
							Host: "https://example.com",
						},
					}
				},
			},
		} {
			t.Run(d.testName, func(t *testing.T) {
				t.Parallel()

				client := d.mockSetup()

				got, err := client.DynamicClient()

				if len(d.errMsg) == 0 && got == nil {
					t.Errorf("K8sProviderClient.DynamicClient returned nil, want non-nil")
				}

				var errMsg string
				if err != nil {
					errMsg = err.Error()
				}

				if errMsg != d.errMsg {
					t.Errorf("K8sProviderClient.DynamicClient returned error message %q, want %q", errMsg, d.errMsg)
				}
			})
		}
	})

	t.Run("RESTMapper", func(t *testing.T) {
		t.Parallel()

		for _, d := range []struct {
			testName   string
			mockSetup  func() K8sProviderClient
			restConfig *rest.Config
			errMsg     string
		}{
			{
				testName:  "rest_config_nil",
				mockSetup: func() K8sProviderClient { return K8sProviderClient{} },
				errMsg:    "rest config is required",
			},
			{
				testName: "rest_mapper_cached",
				mockSetup: func() K8sProviderClient {
					return K8sProviderClient{
						restConfig: &rest.Config{},
						restMapper: &restMapperStub{},
					}
				},
			},
			{
				testName: "new_rest_mapper",
				mockSetup: func() K8sProviderClient {
					return K8sProviderClient{
						restConfig: &rest.Config{
							Host: "https://example.com",
						},
						discoveryClient: &discoveryClientStub{},
					}
				},
			},
		} {
			t.Run(d.testName, func(t *testing.T) {
				t.Parallel()

				client := d.mockSetup()

				got, err := client.RESTMapper()

				if len(d.errMsg) == 0 && got == nil {
					t.Errorf("K8sProviderClient.RESTMapper returned nil, want non-nil")
				}

				var errMsg string
				if err != nil {
					errMsg = err.Error()
				}

				if errMsg != d.errMsg {
					t.Errorf("K8sProviderClient.RESTMapper returned error message %q, want %q", errMsg, d.errMsg)
				}
			})
		}
	})
}
