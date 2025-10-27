package discovery

import (
	"testing"

	x402types "github.com/seanmcgary/x402-go/pkg/types"
)

func TestNewService(t *testing.T) {
	service := NewService()

	if service == nil {
		t.Fatal("Expected non-nil service")
	}

	if len(service.resources) != 0 {
		t.Errorf("Expected empty resources, got %d", len(service.resources))
	}
}

func TestAddResource(t *testing.T) {
	service := NewService()

	resource := x402types.DiscoveredResource{
		Resource:    "https://api.example.com/premium-data",
		Type:        "http",
		X402Version: 1,
		Accepts: []x402types.PaymentRequirements{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/premium-data",
				Description:       "Premium market data",
				MaxTimeoutSeconds: 60,
			},
		},
		LastUpdated: 1703123456,
	}

	service.AddResource(resource)

	all := service.GetAll()
	if len(all) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(all))
	}

	if all[0].Resource != resource.Resource {
		t.Errorf("Expected resource %s, got %s", resource.Resource, all[0].Resource)
	}
}

func TestRemoveResource(t *testing.T) {
	service := NewService()

	resource := x402types.DiscoveredResource{
		Resource:    "https://api.example.com/premium-data",
		Type:        "http",
		X402Version: 1,
		LastUpdated: 1703123456,
	}

	service.AddResource(resource)

	// Remove existing resource
	removed := service.RemoveResource(resource.Resource)
	if !removed {
		t.Error("Expected resource to be removed")
	}

	if len(service.GetAll()) != 0 {
		t.Errorf("Expected empty resources after removal, got %d", len(service.GetAll()))
	}

	// Try to remove non-existent resource
	removed = service.RemoveResource("https://nonexistent.com")
	if removed {
		t.Error("Expected false when removing non-existent resource")
	}
}

func TestQuery(t *testing.T) {
	service := NewService()

	// Add multiple resources
	resources := []x402types.DiscoveredResource{
		{
			Resource:    "https://api.example.com/data1",
			Type:        "http",
			X402Version: 1,
			LastUpdated: 1703123456,
		},
		{
			Resource:    "https://api.example.com/data2",
			Type:        "http",
			X402Version: 1,
			LastUpdated: 1703123457,
		},
		{
			Resource:    "grpc://api.example.com/data3",
			Type:        "grpc",
			X402Version: 1,
			LastUpdated: 1703123458,
		},
	}

	for _, r := range resources {
		service.AddResource(r)
	}

	// Test query without filter
	t.Run("no filter", func(t *testing.T) {
		resp := service.Query("", 10, 0)

		if resp.X402Version != 1 {
			t.Errorf("Expected x402Version 1, got %d", resp.X402Version)
		}

		if len(resp.Items) != 3 {
			t.Errorf("Expected 3 items, got %d", len(resp.Items))
		}

		if resp.Pagination.Total != 3 {
			t.Errorf("Expected total 3, got %d", resp.Pagination.Total)
		}

		if resp.Pagination.Limit != 10 {
			t.Errorf("Expected limit 10, got %d", resp.Pagination.Limit)
		}

		if resp.Pagination.Offset != 0 {
			t.Errorf("Expected offset 0, got %d", resp.Pagination.Offset)
		}
	})

	// Test query with type filter
	t.Run("filter by type", func(t *testing.T) {
		resp := service.Query("http", 10, 0)

		if len(resp.Items) != 2 {
			t.Errorf("Expected 2 http items, got %d", len(resp.Items))
		}

		if resp.Pagination.Total != 2 {
			t.Errorf("Expected total 2, got %d", resp.Pagination.Total)
		}
	})

	// Test pagination
	t.Run("pagination", func(t *testing.T) {
		resp := service.Query("", 2, 0)

		if len(resp.Items) != 2 {
			t.Errorf("Expected 2 items with limit=2, got %d", len(resp.Items))
		}

		// Get second page
		resp2 := service.Query("", 2, 2)

		if len(resp2.Items) != 1 {
			t.Errorf("Expected 1 item on second page, got %d", len(resp2.Items))
		}
	})

	// Test limit validation
	t.Run("default limit", func(t *testing.T) {
		resp := service.Query("", 0, 0)

		if resp.Pagination.Limit != 20 {
			t.Errorf("Expected default limit 20, got %d", resp.Pagination.Limit)
		}
	})

	// Test max limit
	t.Run("max limit", func(t *testing.T) {
		resp := service.Query("", 200, 0)

		if resp.Pagination.Limit != 20 {
			t.Errorf("Expected limit capped at 20, got %d", resp.Pagination.Limit)
		}
	})

	// Test offset beyond total
	t.Run("offset beyond total", func(t *testing.T) {
		resp := service.Query("", 10, 100)

		if len(resp.Items) != 0 {
			t.Errorf("Expected 0 items with offset beyond total, got %d", len(resp.Items))
		}
	})
}

func TestClear(t *testing.T) {
	service := NewService()

	// Add resources
	service.AddResource(x402types.DiscoveredResource{
		Resource:    "https://api.example.com/data1",
		Type:        "http",
		X402Version: 1,
		LastUpdated: 1703123456,
	})

	service.AddResource(x402types.DiscoveredResource{
		Resource:    "https://api.example.com/data2",
		Type:        "http",
		X402Version: 1,
		LastUpdated: 1703123457,
	})

	if len(service.GetAll()) != 2 {
		t.Errorf("Expected 2 resources before clear, got %d", len(service.GetAll()))
	}

	// Clear
	service.Clear()

	if len(service.GetAll()) != 0 {
		t.Errorf("Expected 0 resources after clear, got %d", len(service.GetAll()))
	}
}

func TestGetAll(t *testing.T) {
	service := NewService()

	resources := []x402types.DiscoveredResource{
		{
			Resource:    "https://api.example.com/data1",
			Type:        "http",
			X402Version: 1,
			LastUpdated: 1703123456,
		},
		{
			Resource:    "https://api.example.com/data2",
			Type:        "http",
			X402Version: 1,
			LastUpdated: 1703123457,
		},
	}

	for _, r := range resources {
		service.AddResource(r)
	}

	all := service.GetAll()

	if len(all) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(all))
	}

	// Verify it's a copy (modifying returned slice shouldn't affect internal state)
	all[0].Resource = "modified"
	if service.resources[0].Resource == "modified" {
		t.Error("GetAll should return a copy, not a reference")
	}
}
