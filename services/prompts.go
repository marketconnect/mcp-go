// MCP provides a standardized way for servers to expose prompt templates to clients.
// The Prompt Specification is a structured template for AI model interactions that enables
// consistent message formatting, parameter substitution, context injection, response formatting, and instruction templating.

package services

// https://modelcontextprotocol.io/sdk/java/mcp-server#prompt-specification
type Prompt struct {
	Name    string
	Content string
}

type PromptService struct {
	prompts map[string]Prompt
}

// NewPromptService creates a new prompt service
func NewPromptService(initial []Prompt) *PromptService {
	ps := &PromptService{prompts: make(map[string]Prompt)}
	for _, p := range initial {
		ps.prompts[p.Name] = p
	}
	return ps
}

// List returns a list of prompt names
func (ps *PromptService) Add(name string, content string) {
	ps.prompts[name] = Prompt{Name: name, Content: content}
}

// List returns a list of prompt names
func (ps *PromptService) List() []string {
	names := []string{}
	for name := range ps.prompts {
		names = append(names, name)
	}
	return names
}

// Get returns a prompt by name
func (ps *PromptService) Get(name string) (Prompt, bool) {
	p, ok := ps.prompts[name]
	return p, ok
}
