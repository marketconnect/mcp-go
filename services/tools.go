package services

import "fmt"

// https://modelcontextprotocol.io/docs/concepts/tools#tool-definition-structure
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Fn          ToolFunc               `json:"-"`
}

// ToolFunc is a function type for implementing a tool that can be called by the LLM.
// It takes a generic map of parameters and returns a result or an error.
type ToolFunc func(params map[string]interface{}) (interface{}, error)

// ToolService provides access to a list of tools
type ToolService struct {
	tools map[string]Tool
}

// NewToolService creates a new tool service
func NewToolService() *ToolService {
	return &ToolService{tools: make(map[string]Tool)}
}

// Register registers a new tool
func (ts *ToolService) Register(name string, description string, inputSchema map[string]interface{}, fn ToolFunc) {
	ts.tools[name] = Tool{Name: name, Description: description, InputSchema: inputSchema, Fn: fn}
}

// List returns a list of tool names
func (ts *ToolService) List() []string {
	names := []string{}
	for name := range ts.tools {
		names = append(names, name)
	}
	return names
}

// AllTools returns a list of all tools
func (ts *ToolService) AllTools() []Tool {
	list := []Tool{}
	for _, tool := range ts.tools {
		list = append(list, tool)
	}
	return list
}

// Call calls a tool by name and returns its result or an error
func (ts *ToolService) Call(name string, params map[string]interface{}) (interface{}, error) {
	tool, exists := ts.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found")
	}
	// Invoke the tool's function
	return tool.Fn(params)
}
