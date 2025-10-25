# End-to-End Testing Guide

## Overview

This guide explains how to run end-to-end tests for the Kubernetes MCP server. The tests use your local Kubernetes connection to verify that the MCP server behaves correctly.

## Prerequisites

1. **Local Kubernetes Cluster**: You need a running Kubernetes cluster accessible via your local kubeconfig
   - Minikube, Kind, Docker Desktop, or any other local cluster
   - Or access to a remote cluster via kubeconfig

2. **Go 1.24+**: Required to run the tests

3. **kubectl configured**: Ensure `kubectl` can access your cluster
   ```bash
   kubectl cluster-info
   kubectl get nodes
   ```

## Running the Tests

### Run All Tests

```bash
cd /Users/jamiemacmanus/Desktop/mcp_servers/k8s-mcp-server
go test ./handlers -v
```

### Run Specific Test

```bash
# Test field projection feature
go test ./handlers -v -run TestListResourcesWithFieldProjection

# Test getting specific resources
go test ./handlers -v -run TestGetResources

# Test API resources listing
go test ./handlers -v -run TestGetAPIResources

# Test field projection helper functions
go test ./handlers -v -run TestFieldProjectionHelpers
```

### Run with Coverage

```bash
go test ./handlers -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Structure

### `handlers/k8s_test.go`

The main test file contains the following test suites:

#### 1. **TestListResourcesWithFieldProjection**
Tests the field projection feature with various scenarios:

- **List pods without field projection**: Verifies full object structure is returned
- **List pods with single field projection**: Tests `fieldPaths: "metadata.name"`
- **List pods with multiple field projections**: Tests `fieldPaths: "metadata.name,metadata.namespace,status.phase"`
- **List pods with label selector**: Combines filtering with field projection
- **List nodes with field projection**: Tests on different resource types
- **Missing required parameters**: Validates error handling

#### 2. **TestGetResources**
Tests retrieving specific resources:

- Lists pods to find a valid pod name
- Retrieves a specific pod by name
- Verifies the returned structure

#### 3. **TestGetAPIResources**
Tests listing all available API resources in the cluster:

- Retrieves all namespace-scoped and cluster-scoped resources
- Validates the response structure

#### 4. **TestFieldProjectionHelpers**
Unit tests for field projection helper functions:

- `extractFieldValue()`: Tests extracting values from nested maps
- `setFieldValue()`: Tests setting values in nested maps
- `projectFields()`: Tests the complete projection logic

## What the Tests Verify

### Field Projection Tests

1. **Without Field Projection**:
   - Full objects are returned with `metadata`, `spec`, and `status`
   - All nested fields are present

2. **With Field Projection**:
   - Only specified fields are returned
   - Nested structure is preserved
   - Unspecified fields are excluded
   - Multiple paths can be specified

3. **Edge Cases**:
   - Empty clusters (no resources)
   - Missing fields in resources
   - Invalid field paths
   - Whitespace handling in comma-separated paths

### Integration Tests

1. **Real Kubernetes Connection**:
   - Tests use your actual kubeconfig
   - Queries real cluster resources
   - Validates against actual API responses

2. **Error Handling**:
   - Missing required parameters
   - Invalid resource types
   - Non-existent resources

## Expected Test Output

### Successful Run

```
=== RUN   TestListResourcesWithFieldProjection
=== RUN   TestListResourcesWithFieldProjection/List_pods_without_field_projection
    k8s_test.go:XX: Successfully retrieved 5 pods with full structure
=== RUN   TestListResourcesWithFieldProjection/List_pods_with_field_projection_-_single_field
    k8s_test.go:XX: Successfully retrieved 5 pods with field projection (metadata.name only)
=== RUN   TestListResourcesWithFieldProjection/List_pods_with_field_projection_-_multiple_fields
    k8s_test.go:XX: Successfully retrieved 5 pods with multiple field projections
=== RUN   TestListResourcesWithFieldProjection/List_pods_with_label_selector
    k8s_test.go:XX: Found 1 pods matching label selector
=== RUN   TestListResourcesWithFieldProjection/List_nodes_with_field_projection
    k8s_test.go:XX: Successfully retrieved 1 nodes with field projection
=== RUN   TestListResourcesWithFieldProjection/Missing_required_Kind_parameter
    k8s_test.go:XX: Got expected error: missing required parameter: Kind
--- PASS: TestListResourcesWithFieldProjection (0.XX s)
=== RUN   TestGetResources
=== RUN   TestGetResources/Get_specific_pod
    k8s_test.go:XX: Successfully retrieved pod: coredns-XXXXX
--- PASS: TestGetResources (0.XX s)
=== RUN   TestGetAPIResources
=== RUN   TestGetAPIResources/Get_all_API_resources
    k8s_test.go:XX: Successfully retrieved 150 API resources
--- PASS: TestGetAPIResources (0.XX s)
=== RUN   TestFieldProjectionHelpers
--- PASS: TestFieldProjectionHelpers (0.XX s)
PASS
ok      github.com/reza-gholizade/k8s-mcp-server/handlers    X.XXXs
```

## Troubleshooting

### Test Failures

1. **"Failed to create Kubernetes client"**
   - Check your kubeconfig: `kubectl config view`
   - Ensure cluster is accessible: `kubectl cluster-info`
   - Verify KUBECONFIG environment variable

2. **"No pods found in kube-system namespace"**
   - This is usually OK - some tests will skip if no resources exist
   - For more comprehensive testing, ensure your cluster has some running pods

3. **Timeout errors**
   - Your cluster might be slow to respond
   - Check cluster health: `kubectl get nodes`
   - Increase timeout if needed (modify test code)

### Running Tests Against Different Clusters

To test against a specific cluster context:

```bash
# List available contexts
kubectl config get-contexts

# Switch context
kubectl config use-context <context-name>

# Run tests
go test ./handlers -v
```

Or use KUBECONFIG environment variable:

```bash
KUBECONFIG=/path/to/kubeconfig go test ./handlers -v
```

## Adding New Tests

To add new test cases:

1. Add a new test function in `handlers/k8s_test.go`:
   ```go
   func TestNewFeature(t *testing.T) {
       client, err := k8s.NewClient("")
       if err != nil {
           t.Fatalf("Failed to create client: %v", err)
       }
       
       // Your test logic here
   }
   ```

2. Follow the existing patterns:
   - Use table-driven tests for multiple scenarios
   - Include validation functions
   - Add descriptive log messages
   - Handle empty clusters gracefully

3. Run your new test:
   ```bash
   go test ./handlers -v -run TestNewFeature
   ```

## Continuous Integration

To run these tests in CI:

1. Ensure a Kubernetes cluster is available (e.g., Kind, k3s)
2. Set up kubeconfig access
3. Run tests as part of your CI pipeline:

```yaml
# Example GitHub Actions
- name: Set up Kind cluster
  uses: helm/kind-action@v1
  
- name: Run e2e tests
  run: go test ./handlers -v
```

## Performance Considerations

- Tests query real Kubernetes API - expect some latency
- Field projection tests should show reduced response sizes
- Use `-short` flag to skip slow tests if needed:
  ```bash
  go test ./handlers -v -short
  ```

## Test Coverage

Current test coverage includes:

- ✅ Field projection with single and multiple paths
- ✅ Full object retrieval (no projection)
- ✅ Label selector filtering
- ✅ Different resource types (pods, nodes)
- ✅ Error handling (missing parameters)
- ✅ Helper function unit tests
- ✅ Real Kubernetes API integration

## Best Practices

1. **Clean Up**: Tests are read-only and don't modify cluster state
2. **Idempotent**: Tests can be run multiple times safely
3. **Isolated**: Each test is independent
4. **Descriptive**: Test names clearly describe what they verify
5. **Validated**: Each test includes specific assertions

## Next Steps

After running tests successfully:

1. Review test output to understand behavior
2. Add tests for new features you implement
3. Use tests to validate bug fixes
4. Consider adding tests for edge cases specific to your use case
