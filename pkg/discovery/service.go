package discovery

import (
	"sync"

	x402types "github.com/seanmcgary/x402-go/pkg/types"
)

// Service manages discoverable x402 resources (the "Bazaar")
type Service struct {
	mu        sync.RWMutex
	resources []x402types.DiscoveredResource
}

// NewService creates a new discovery service
func NewService() *Service {
	return &Service{
		resources: make([]x402types.DiscoveredResource, 0),
	}
}

// AddResource adds a resource to the discovery registry
func (s *Service) AddResource(resource x402types.DiscoveredResource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources = append(s.resources, resource)
}

// RemoveResource removes a resource by its resource URL
func (s *Service) RemoveResource(resourceURL string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, r := range s.resources {
		if r.Resource == resourceURL {
			s.resources = append(s.resources[:i], s.resources[i+1:]...)
			return true
		}
	}
	return false
}

// Query queries resources with filtering and pagination
func (s *Service) Query(resourceType string, limit, offset int) *x402types.DiscoveryResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Apply default limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Filter by type if specified
	filtered := make([]x402types.DiscoveredResource, 0)
	for _, r := range s.resources {
		if resourceType == "" || r.Type == resourceType {
			filtered = append(filtered, r)
		}
	}

	total := len(filtered)

	// Apply pagination
	start := offset
	if start > total {
		start = total
	}

	end := start + limit
	if end > total {
		end = total
	}

	items := filtered[start:end]

	return &x402types.DiscoveryResponse{
		X402Version: 1,
		Items:       items,
		Pagination: x402types.Pagination{
			Limit:  limit,
			Offset: offset,
			Total:  total,
		},
	}
}

// GetAll returns all resources without filtering
func (s *Service) GetAll() []x402types.DiscoveredResource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modifications
	resources := make([]x402types.DiscoveredResource, len(s.resources))
	copy(resources, s.resources)
	return resources
}

// Clear removes all resources
func (s *Service) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources = make([]x402types.DiscoveredResource, 0)
}
