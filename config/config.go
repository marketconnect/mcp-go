package config

// Configuration for the MCP server, including which capabilities tools, resources, and prompts to enable.
// By default all capabilities are enabled (true), but they can be disabled via these flags.

type Config struct {
	Name    string // Server name (for identification in handshake)
	Version string // Server version (for identification in handshake)
	Address string // Listen address (e.g. ":3000" for port 3000)
	// https://spec.modelcontextprotocol.io/specification/2025-03-26/basic/transports/
	SSEPath         string // HTTP endpoint for SSE (server-to-client events)
	MessagesPath    string // HTTP endpoint for client-to-server JSON-RPC messages
	EnableResources bool   // If true, enable Resources capability
	EnableTools     bool   // If true, enable Tools capability
	EnablePrompts   bool   // If true, enable Prompts capability
	EnableSampling  bool   // If true, allow server-initiated sampling requests
}
