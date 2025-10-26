package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/reza-gholizade/k8s-mcp-server/pkg/k8s"
)

// TestGetEvents tests the getEvents functionality with sorting and limiting
func TestGetEvents(t *testing.T) {
	client, err := k8s.NewClient("")
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	handler := GetEvents(client)
	ctx := context.Background()

	t.Run("Get events with default parameters", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      "getEvents",
				Arguments: map[string]interface{}{},
			},
		}

		result, err := handler(ctx, request)
		if err != nil {
			t.Fatalf("Failed to get events: %v", err)
		}

		if result == nil {
			t.Fatal("Expected result but got nil")
		}

		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent in result")
		}

		var events []map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &events); err != nil {
			t.Fatalf("Failed to parse events: %v", err)
		}

		// Should return at most 20 events (default maxEvents)
		if len(events) > 20 {
			t.Errorf("Expected at most 20 events, got %d", len(events))
		}

		t.Logf("Successfully retrieved %d events with default parameters", len(events))
	})

	t.Run("Get events with custom maxEvents", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "getEvents",
				Arguments: map[string]interface{}{
					"maxEvents": float64(5),
				},
			},
		}

		result, err := handler(ctx, request)
		if err != nil {
			t.Fatalf("Failed to get events: %v", err)
		}

		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent in result")
		}

		var events []map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &events); err != nil {
			t.Fatalf("Failed to parse events: %v", err)
		}

		// Should return at most 5 events
		if len(events) > 5 {
			t.Errorf("Expected at most 5 events, got %d", len(events))
		}

		t.Logf("Successfully retrieved %d events with maxEvents=5", len(events))
	})

	t.Run("Get events sorted by lastTime", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "getEvents",
				Arguments: map[string]interface{}{
					"sortBy":    "lastTime",
					"maxEvents": float64(10),
				},
			},
		}

		result, err := handler(ctx, request)
		if err != nil {
			t.Fatalf("Failed to get events: %v", err)
		}

		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent in result")
		}

		var events []map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &events); err != nil {
			t.Fatalf("Failed to parse events: %v", err)
		}

		// Verify events are sorted in descending order (most recent first)
		if len(events) > 1 {
			for i := 0; i < len(events)-1; i++ {
				// Events should be in descending order by lastTime
				t.Logf("Event %d: %v - %s", i, events[i]["lastTime"], events[i]["message"])
			}
		}

		t.Logf("Successfully retrieved %d events sorted by lastTime", len(events))
	})

	t.Run("Get events from specific namespace", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "getEvents",
				Arguments: map[string]interface{}{
					"namespace": "kube-system",
					"maxEvents": float64(10),
				},
			},
		}

		result, err := handler(ctx, request)
		if err != nil {
			t.Fatalf("Failed to get events: %v", err)
		}

		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent in result")
		}

		var events []map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &events); err != nil {
			t.Fatalf("Failed to parse events: %v", err)
		}

		// Verify all events are from kube-system namespace
		for _, event := range events {
			if ns, ok := event["namespace"].(string); ok && ns != "kube-system" {
				t.Errorf("Expected event from kube-system, got %s", ns)
			}
		}

		t.Logf("Successfully retrieved %d events from kube-system namespace", len(events))
	})

	t.Run("Get events with message filter", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "getEvents",
				Arguments: map[string]interface{}{
					"messageFilter": "failed",
					"maxEvents":     float64(10),
				},
			},
		}

		result, err := handler(ctx, request)
		if err != nil {
			t.Fatalf("Failed to get events with message filter: %v", err)
		}

		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent in result")
		}

		var events []map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &events); err != nil {
			t.Fatalf("Failed to parse events: %v", err)
		}

		// Verify all events contain "failed" in their message (case-insensitive)
		for _, event := range events {
			if message, ok := event["message"].(string); ok {
				lowerMessage := strings.ToLower(message)
				if !strings.Contains(lowerMessage, "failed") {
					t.Errorf("Event message does not contain 'failed': %s", message)
				}
			}
		}

		t.Logf("Successfully retrieved %d events containing 'failed' in message", len(events))
	})

	t.Run("Get events with message filter and namespace", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "getEvents",
				Arguments: map[string]interface{}{
					"namespace":     "default",
					"messageFilter": "reconciliation",
					"maxEvents":     float64(5),
				},
			},
		}

		result, err := handler(ctx, request)
		if err != nil {
			t.Fatalf("Failed to get events with message filter and namespace: %v", err)
		}

		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent in result")
		}

		var events []map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &events); err != nil {
			t.Fatalf("Failed to parse events: %v", err)
		}

		// Verify all events are from default namespace and contain "reconciliation"
		for _, event := range events {
			if ns, ok := event["namespace"].(string); ok && ns != "default" {
				t.Errorf("Expected event from default namespace, got %s", ns)
			}
			if message, ok := event["message"].(string); ok {
				lowerMessage := strings.ToLower(message)
				if !strings.Contains(lowerMessage, "reconciliation") {
					t.Errorf("Event message does not contain 'reconciliation': %s", message)
				}
			}
		}

		t.Logf("Successfully retrieved %d events from default namespace containing 'reconciliation'", len(events))
	})
}
