package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"mcp-go/config"
	"mcp-go/protocol"
	"mcp-go/services"
	"mcp-go/transport"
	"net/http"
	"strings"
	"sync"
)

type Server struct {
	cfg       config.Config
	resources *services.ResourceService
	tools     *services.ToolService
	prompts   *services.PromptService
	sampling  *services.SamplingService
	transport *transport.SSETransport
	pendingMu sync.Mutex
	pending   map[string]chan *protocol.JSONRPCResponse
	nextID    int
}

func StartServer(cfg config.Config) error {
	// Set default values
	if cfg.Address == "" {
		cfg.Address = ":3000"
	}
	if cfg.SSEPath == "" {
		cfg.SSEPath = "/sse"
	}
	if cfg.MessagesPath == "" {
		cfg.MessagesPath = "/messages"
	}

	// Initialize the server
	server := &Server{
		cfg:       cfg,
		resources: services.NewResourceService(nil),
		tools:     services.NewToolService(),
		prompts:   services.NewPromptService(nil),
		sampling:  services.NewSamplingService(cfg.EnableSampling),
		pending:   make(map[string]chan *protocol.JSONRPCResponse),
	}
	// HTTP handler for JSON-RPC messages (POST)
	http.HandleFunc(cfg.SSEPath, func(w http.ResponseWriter, r *http.Request) {
		// Upgrade connection to SSE
		server.transport = transport.NewSSETransport(w)
		// Keep the connection open until the client disconnects
		<-r.Context().Done()
		// Client disconnected; close the SSE channel
		server.transport.Close()
	})
	// HTTP handler for JSON-RPC messages (POST)
	http.HandleFunc(cfg.MessagesPath, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			server.handleMessage(body)
		}
		// Respond with 200 OK (actual JSON-RPC responses are sent via SSE events)
		w.WriteHeader(http.StatusOK)
	})

	// Start HTTP server (using HTTP + SSE transport)&#8203;:contentReference[oaicite:21]{index=21}
	fmt.Printf("MCP server \"%s\" (v%s) listening at %s\n", cfg.Name, cfg.Version, cfg.Address)
	return http.ListenAndServe(cfg.Address, nil)
}

func (s *Server) handleMessage(jsonData []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(jsonData, &msg); err != nil {
		// TODO:
		resp := protocol.NewErrorResponse(nil, protocol.ParseError, "Parse error: "+err.Error(), nil)
		_ = s.transport.Send(resp)
		return
	}

	if _, hasMethod := msg["method"]; hasMethod {
		if _, hasID := msg["id"]; hasID {
			// It's a Request from client
			id := msg["id"]
			method := msg["method"].(string)
			params := msg["params"]
			res := s.handleRequest(method, params, id)
			// Send the response back via SSE
			if res != nil {
				_ = s.transport.Send(res)
			}
		} else {
			// It's a Notification (one-way message from client)
			method := msg["method"].(string)
			params := msg["params"]
			s.handleNotification(method, params)
		}
	} else if _, hasID := msg["id"]; hasID {
		// It's a Response to a server-initiated request (e.g. from sampling)
		idStr := fmt.Sprintf("%v", msg["id"])
		s.pendingMu.Lock()
		ch, ok := s.pending[idStr]
		if ok {
			delete(s.pending, idStr)
		}
		s.pendingMu.Unlock()
		if ok {
			// Unmarshal into JSONRPCResponse and notify the waiting call
			var resp protocol.JSONRPCResponse
			_ = json.Unmarshal(jsonData, &resp)
			ch <- &resp
			close(ch)
		}
	}
}

func (s *Server) handleRequest(method string, params interface{}, id interface{}) *protocol.JSONRPCResponse {
	// initialization:
	// https://modelcontextprotocol.io/docs/concepts/architecture#connection-lifecycle
	// Client sends initialize request with protocol version and capabilities
	if method == "initialize" {
		// Check capabilities
		caps := make(map[string]interface{})
		if s.cfg.EnableResources {
			caps["resources"] = struct{}{}
		}
		if s.cfg.EnableTools {
			caps["tools"] = struct{}{}
		}
		if s.cfg.EnablePrompts {
			caps["prompts"] = struct{}{}
		}
		// Server responds with its protocol version and capabilities
		result := map[string]interface{}{
			"name":         s.cfg.Name,
			"version":      s.cfg.Version,
			"capabilities": caps,
		}
		return protocol.NewResponse(id, result)
	}

	// Resources:
	if method == "resources/list" {
		if !s.cfg.EnableResources {
			return protocol.NewErrorResponse(id, protocol.CapabilityDisabled, "Resources capability is disabled", nil)
		}
		// List available resources
		resources := s.resources.List()
		return protocol.NewResponse(id, map[string]interface{}{
			"resources": resources,
		})
	}
	if method == "resources/read" {
		if !s.cfg.EnableResources {
			return protocol.NewErrorResponse(id, protocol.CapabilityDisabled, "Resources capability is disabled", nil)
		}

		var uri string
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if u, ok2 := paramsMap["uri"].(string); ok2 {
				uri = u
			}
		}
		if uri == "" {
			return protocol.NewErrorResponse(id, protocol.InvalidParams, "Missing 'uri' parameter", nil)
		}
		res, found := s.resources.Get(uri)
		if !found {
			return protocol.NewErrorResponse(id, protocol.CapabilityDisabled, fmt.Sprintf("Resource '%s' not found", uri), nil)
		}
		return protocol.NewResponse(id, map[string]interface{}{
			"content": res.Content,
		})
	}

	if method == "tools/list" {
		if !s.cfg.EnableTools {
			return protocol.NewErrorResponse(id, protocol.CapabilityDisabled, "Tools capability is disabled", nil)
		}
		// List tools with names and descriptions
		tools := s.tools.AllTools()
		return protocol.NewResponse(id, map[string]interface{}{
			"tools": tools,
		})
	}
	if strings.HasPrefix(method, "tools/") {
		if !s.cfg.EnableTools {
			return protocol.NewErrorResponse(id, protocol.CapabilityDisabled, "Tools capability is disabled", nil)
		}
		toolName := strings.TrimPrefix(method, "tools/")
		if toolName == "" {
			return protocol.NewErrorResponse(id, protocol.MethodNotFound, "Invalid tool method", nil)
		}
		// Parameters for the tool call
		var paramsMap map[string]interface{}
		if params != nil {
			if pm, ok := params.(map[string]interface{}); ok {
				paramsMap = pm
			}
		}
		result, err := s.tools.Call(toolName, paramsMap)
		if err != nil {
			// Tool not found or execution error
			return protocol.NewErrorResponse(id, protocol.InternalError, err.Error(), nil)
		}
		return protocol.NewResponse(id, result)
	}
	if method == "prompts/list" {
		if !s.cfg.EnablePrompts {
			return protocol.NewErrorResponse(id, protocol.CapabilityDisabled, "Prompts capability is disabled", nil)
		}
		names := s.prompts.List()
		return protocol.NewResponse(id, map[string]interface{}{
			"prompts": names,
		})
	}
	if method == "prompts/get" {
		if !s.cfg.EnablePrompts {
			return protocol.NewErrorResponse(id, protocol.CapabilityDisabled, "Prompts capability is disabled", nil)
		}
		var name string
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if n, ok2 := paramsMap["name"].(string); ok2 {
				name = n
			}
		}
		if name == "" {
			return protocol.NewErrorResponse(id, protocol.InvalidParams, "Missing 'name' parameter", nil)
		}
		prompt, ok := s.prompts.Get(name)
		if !ok {
			return protocol.NewErrorResponse(id, protocol.CapabilityDisabled, fmt.Sprintf("Prompt '%s' not found", name), nil)
		}
		return protocol.NewResponse(id, map[string]interface{}{
			"content": prompt.Content,
		})
	}

	// If method is unrecognized or not implemented
	return protocol.NewErrorResponse(id, protocol.MethodNotFound, fmt.Sprintf("Method '%s' not found", method), nil)
}
