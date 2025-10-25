# Field Projection Feature

## Overview
The `listResources` tool now supports field projection via the `fieldPaths` parameter. This allows you to specify exactly which fields to return from each resource, dramatically reducing response size and preventing LLM timeouts.

## How It Works

### Without Field Projection (returns full objects)
```json
{
  "Kind": "pods",
  "namespace": "default"
}
```
Returns complete pod objects with all fields (metadata, spec, status, etc.) - can be very large!

### With Field Projection (returns only specified fields)
```json
{
  "Kind": "pods",
  "namespace": "default",
  "fieldPaths": "metadata.name,metadata.namespace,status.phase"
}
```
Returns only:
```json
[
  {
    "metadata": {
      "name": "my-pod-1",
      "namespace": "default"
    },
    "status": {
      "phase": "Running"
    }
  },
  {
    "metadata": {
      "name": "my-pod-2",
      "namespace": "default"
    },
    "status": {
      "phase": "Pending"
    }
  }
]
```

## Example Use Cases

### 1. List pod names and statuses only
```json
{
  "Kind": "pods",
  "fieldPaths": "metadata.name,status.phase,status.conditions"
}
```

### 2. List deployment names and replica counts
```json
{
  "Kind": "deployments",
  "namespace": "production",
  "fieldPaths": "metadata.name,spec.replicas,status.availableReplicas"
}
```

### 3. List service names and types
```json
{
  "Kind": "services",
  "fieldPaths": "metadata.name,spec.type,spec.ports"
}
```

### 4. Combine with filtering
```json
{
  "Kind": "pods",
  "namespace": "default",
  "labelSelector": "app=nginx",
  "fieldSelector": "status.phase=Running",
  "fieldPaths": "metadata.name,status.podIP,status.hostIP"
}
```

## Benefits

1. **Reduced Response Size**: Only return the data you need
2. **Prevent Timeouts**: Smaller responses are faster to process
3. **Better Performance**: Less data to serialize/deserialize
4. **Clearer Intent**: Explicitly specify what information you need

## Implementation Details

- **Path Format**: Use dot-notation for nested fields (e.g., `metadata.name`, `status.conditions`)
- **Multiple Paths**: Separate multiple paths with commas
- **Whitespace**: Spaces around commas are automatically trimmed
- **Missing Fields**: If a field doesn't exist in a resource, it's simply omitted from the result
- **Nested Structures**: The projection preserves the nested structure of the original object
