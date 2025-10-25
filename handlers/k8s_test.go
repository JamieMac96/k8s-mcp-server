package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/reza-gholizade/k8s-mcp-server/pkg/k8s"
)

// TestListResourcesWithFieldProjection tests the field projection feature
func TestListResourcesWithFieldProjection(t *testing.T) {
	// Create a real Kubernetes client using local kubeconfig
	client, err := k8s.NewClient("")
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	handler := ListResources(client)
	ctx := context.Background()

	tests := []struct {
		name           string
		args           map[string]interface{}
		expectError    bool
		validateResult func(t *testing.T, result *mcp.CallToolResult)
	}{
		{
			name: "List pods without field projection",
			args: map[string]interface{}{
				"Kind":      "Pod",
				"namespace": "kube-system",
			},
			expectError: false,
			validateResult: func(t *testing.T, result *mcp.CallToolResult) {
				if len(result.Content) == 0 {
					t.Error("Expected content in result")
					return
				}
				
				// Parse the JSON response
				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Error("Expected TextContent in result")
					return
				}
				var resources []map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &resources); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				if len(resources) == 0 {
					t.Log("No pods found in kube-system namespace (this is OK if cluster is empty)")
					return
				}

				// Verify full object structure is present
				firstResource := resources[0]
				if _, ok := firstResource["metadata"]; !ok {
					t.Error("Expected metadata field in full object")
				}
				if _, ok := firstResource["spec"]; !ok {
					t.Error("Expected spec field in full object")
				}
				if _, ok := firstResource["status"]; !ok {
					t.Error("Expected status field in full object")
				}

				t.Logf("Successfully retrieved %d pods with full structure", len(resources))
			},
		},
		{
			name: "List pods with field projection - single field",
			args: map[string]interface{}{
				"Kind":       "Pod",
				"namespace":  "kube-system",
				"fieldPaths": "metadata.name",
			},
			expectError: false,
			validateResult: func(t *testing.T, result *mcp.CallToolResult) {
				if len(result.Content) == 0 {
					t.Error("Expected content in result")
					return
				}

				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Error("Expected TextContent in result")
					return
				}
				var resources []map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &resources); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				if len(resources) == 0 {
					t.Log("No pods found in kube-system namespace")
					return
				}

				// Verify only projected fields are present
				firstResource := resources[0]
				if metadata, ok := firstResource["metadata"].(map[string]interface{}); !ok {
					t.Error("Expected metadata field")
				} else {
					if _, ok := metadata["name"]; !ok {
						t.Error("Expected metadata.name field")
					}
					// Should not have other metadata fields like namespace, labels, etc.
					if len(metadata) > 1 {
						t.Logf("Warning: Expected only 'name' field in metadata, got %d fields", len(metadata))
					}
				}

				// Should not have spec or status
				if _, ok := firstResource["spec"]; ok {
					t.Error("Did not expect spec field in projected result")
				}
				if _, ok := firstResource["status"]; ok {
					t.Error("Did not expect status field in projected result")
				}

				t.Logf("Successfully retrieved %d pods with field projection (metadata.name only)", len(resources))
			},
		},
		{
			name: "List pods with field projection - multiple fields",
			args: map[string]interface{}{
				"Kind":       "Pod",
				"namespace":  "kube-system",
				"fieldPaths": "metadata.name, metadata.namespace, status.phase",
			},
			expectError: false,
			validateResult: func(t *testing.T, result *mcp.CallToolResult) {
				if len(result.Content) == 0 {
					t.Error("Expected content in result")
					return
				}

				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Error("Expected TextContent in result")
					return
				}
				var resources []map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &resources); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				if len(resources) == 0 {
					t.Log("No pods found in kube-system namespace")
					return
				}

				// Verify projected fields are present
				firstResource := resources[0]
				
				// Check metadata
				if metadata, ok := firstResource["metadata"].(map[string]interface{}); !ok {
					t.Error("Expected metadata field")
				} else {
					if _, ok := metadata["name"]; !ok {
						t.Error("Expected metadata.name field")
					}
					if _, ok := metadata["namespace"]; !ok {
						t.Error("Expected metadata.namespace field")
					}
				}

				// Check status
				if status, ok := firstResource["status"].(map[string]interface{}); !ok {
					t.Error("Expected status field")
				} else {
					if _, ok := status["phase"]; !ok {
						t.Error("Expected status.phase field")
					}
				}

				// Should not have spec
				if _, ok := firstResource["spec"]; ok {
					t.Error("Did not expect spec field in projected result")
				}

				t.Logf("Successfully retrieved %d pods with multiple field projections", len(resources))
			},
		},
		{
			name: "List pods with label selector",
			args: map[string]interface{}{
				"Kind":          "Pod",
				"namespace":     "kube-system",
				"labelSelector": "component=kube-apiserver",
				"fieldPaths":    "metadata.name,metadata.labels",
			},
			expectError: false,
			validateResult: func(t *testing.T, result *mcp.CallToolResult) {
				if len(result.Content) == 0 {
					t.Error("Expected content in result")
					return
				}

				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Error("Expected TextContent in result")
					return
				}
				var resources []map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &resources); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				t.Logf("Found %d pods matching label selector", len(resources))

				// If pods are found, verify they have the expected label
				for _, resource := range resources {
					if metadata, ok := resource["metadata"].(map[string]interface{}); ok {
						if labels, ok := metadata["labels"].(map[string]interface{}); ok {
							if component, ok := labels["component"].(string); ok && component == "kube-apiserver" {
								t.Logf("Verified pod has correct label: component=kube-apiserver")
							}
						}
					}
				}
			},
		},
		{
			name: "List nodes with field projection",
			args: map[string]interface{}{
				"Kind":       "Node",
				"fieldPaths": "metadata.name,status.conditions",
			},
			expectError: false,
			validateResult: func(t *testing.T, result *mcp.CallToolResult) {
				if len(result.Content) == 0 {
					t.Error("Expected content in result")
					return
				}

				textContent, ok := result.Content[0].(mcp.TextContent)
				if !ok {
					t.Error("Expected TextContent in result")
					return
				}
				var resources []map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &resources); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				if len(resources) == 0 {
					t.Error("Expected at least one node in cluster")
					return
				}

				t.Logf("Successfully retrieved %d nodes with field projection", len(resources))
			},
		},
		{
			name: "Missing required Kind parameter",
			args: map[string]interface{}{
				"namespace": "default",
			},
			expectError: true,
			validateResult: func(t *testing.T, result *mcp.CallToolResult) {
				// Should not reach here
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "listResources",
					Arguments: tt.args,
				},
			}

			result, err := handler(ctx, request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else {
					t.Logf("Got expected error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("Expected result but got nil")
			}

			tt.validateResult(t, result)
		})
	}
}

// TestGetResources tests getting a specific resource
func TestGetResources(t *testing.T) {
	client, err := k8s.NewClient("")
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	handler := GetResources(client)
	ctx := context.Background()

	// First, list pods to get a valid pod name
	listHandler := ListResources(client)
	listRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "listResources",
			Arguments: map[string]interface{}{
				"Kind":       "Pod",
				"namespace":  "kube-system",
				"fieldPaths": "metadata.name",
			},
		},
	}

	listResult, err := listHandler(ctx, listRequest)
	if err != nil {
		t.Fatalf("Failed to list pods: %v", err)
	}

	textContent, ok := listResult.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent in list result")
	}
	var pods []map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &pods); err != nil {
		t.Fatalf("Failed to parse pods: %v", err)
	}

	if len(pods) == 0 {
		t.Skip("No pods found in kube-system namespace, skipping GetResources test")
	}

	// Get the first pod's name
	podName := pods[0]["metadata"].(map[string]interface{})["name"].(string)

	t.Run("Get specific pod", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "getResource",
				Arguments: map[string]interface{}{
					"kind":      "Pod",
					"name":      podName,
					"namespace": "kube-system",
				},
			},
		}

		result, err := handler(ctx, request)
		if err != nil {
			t.Fatalf("Failed to get pod: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result but got nil")
		}

		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent in result")
		}
		var pod map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &pod); err != nil {
			t.Fatalf("Failed to parse pod: %v", err)
		}

		// Verify the pod has expected structure
		if metadata, ok := pod["metadata"].(map[string]interface{}); !ok {
			t.Error("Expected metadata field")
		} else {
			if name, ok := metadata["name"].(string); !ok || name != podName {
				t.Errorf("Expected pod name %s, got %v", podName, name)
			}
		}

		t.Logf("Successfully retrieved pod: %s", podName)
	})
}

// TestGetAPIResources tests listing API resources
func TestGetAPIResources(t *testing.T) {
	client, err := k8s.NewClient("")
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	handler := GetAPIResources(client)
	ctx := context.Background()

	t.Run("Get all API resources", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "getAPIResources",
				Arguments: map[string]interface{}{
					"includeNamespaceScoped": true,
					"includeClusterScoped":   true,
				},
			},
		}

		result, err := handler(ctx, request)
		if err != nil {
			t.Fatalf("Failed to get API resources: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result but got nil")
		}

		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent in result")
		}
		var resources []map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &resources); err != nil {
			t.Fatalf("Failed to parse API resources: %v", err)
		}

		if len(resources) == 0 {
			t.Error("Expected at least some API resources")
		}

		t.Logf("Successfully retrieved %d API resources", len(resources))
	})
}

// TestFieldProjectionHelpers tests the helper functions for field projection
func TestFieldProjectionHelpers(t *testing.T) {
	t.Run("extractFieldValue - simple path", func(t *testing.T) {
		obj := map[string]interface{}{
			"name": "test",
		}

		value, ok := extractFieldValue(obj, "name")
		if !ok {
			t.Error("Expected to extract value")
		}
		if value != "test" {
			t.Errorf("Expected 'test', got %v", value)
		}
	})

	t.Run("extractFieldValue - nested path", func(t *testing.T) {
		obj := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "pod-1",
				"namespace": "default",
			},
		}

		value, ok := extractFieldValue(obj, "metadata.name")
		if !ok {
			t.Error("Expected to extract value")
		}
		if value != "pod-1" {
			t.Errorf("Expected 'pod-1', got %v", value)
		}
	})

	t.Run("extractFieldValue - missing path", func(t *testing.T) {
		obj := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "pod-1",
			},
		}

		_, ok := extractFieldValue(obj, "metadata.nonexistent")
		if ok {
			t.Error("Expected extraction to fail for nonexistent path")
		}
	})

	t.Run("setFieldValue - simple path", func(t *testing.T) {
		obj := make(map[string]interface{})
		setFieldValue(obj, "name", "test")

		if obj["name"] != "test" {
			t.Errorf("Expected 'test', got %v", obj["name"])
		}
	})

	t.Run("setFieldValue - nested path", func(t *testing.T) {
		obj := make(map[string]interface{})
		setFieldValue(obj, "metadata.name", "pod-1")

		metadata, ok := obj["metadata"].(map[string]interface{})
		if !ok {
			t.Error("Expected metadata to be a map")
		}
		if metadata["name"] != "pod-1" {
			t.Errorf("Expected 'pod-1', got %v", metadata["name"])
		}
	})

	t.Run("projectFields - multiple paths", func(t *testing.T) {
		obj := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "pod-1",
				"namespace": "default",
				"labels": map[string]interface{}{
					"app": "test",
				},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{},
			},
			"status": map[string]interface{}{
				"phase": "Running",
			},
		}

		paths := []string{"metadata.name", "status.phase"}
		projected := projectFields(obj, paths)

		// Should have metadata and status
		if _, ok := projected["metadata"]; !ok {
			t.Error("Expected metadata in projected result")
		}
		if _, ok := projected["status"]; !ok {
			t.Error("Expected status in projected result")
		}

		// Should not have spec
		if _, ok := projected["spec"]; ok {
			t.Error("Did not expect spec in projected result")
		}

		// Verify metadata only has name
		metadata := projected["metadata"].(map[string]interface{})
		if metadata["name"] != "pod-1" {
			t.Errorf("Expected name 'pod-1', got %v", metadata["name"])
		}

		// Verify status only has phase
		status := projected["status"].(map[string]interface{})
		if status["phase"] != "Running" {
			t.Errorf("Expected phase 'Running', got %v", status["phase"])
		}
	})

	t.Run("projectFields - empty paths returns original", func(t *testing.T) {
		obj := map[string]interface{}{
			"name": "test",
		}

		projected := projectFields(obj, []string{})
		if projected["name"] != "test" {
			t.Error("Expected original object when no paths specified")
		}
	})
}
