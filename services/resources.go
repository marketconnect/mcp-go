// Resources provide context to AI models by exposing data such as:
// File contents, Database records, API responses, System information, Application state.
// https://modelcontextprotocol.io/sdk/java/mcp-server#resource-specification
package services

type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
	Content     string `json:"-"` // content data (excluded from JSON listing)
}

type ResourceService struct {
	items []Resource
}

// NewResourceService creates a new resource service
func NewResourceService(initial []Resource) *ResourceService {
	return &ResourceService{items: initial}
}

// List returns a list of resources
func (rs *ResourceService) List() []Resource {
	return rs.items
}

// Add adds a new resource
func (rs *ResourceService) Add(res Resource) {
	rs.items = append(rs.items, res)
}

// Get returns a resource by URI
func (rs *ResourceService) Get(uri string) (Resource, bool) {
	for _, res := range rs.items {
		if res.URI == uri {
			return res, true
		}
	}
	return Resource{}, false
}
