// provides a standardized way for servers to request LLM sampling
// (“completions” or “generations”) from language models via clients.
// Clients that support sampling MUST declare the sampling capability during initialization:
// {
// 	"capabilities": {
// 	  "sampling": {}
// 	}
// }
// To request a language model generation, servers send a sampling/createMessage request to the client.

// https://spec.modelcontextprotocol.io/specification/2024-11-05/client/sampling/
package services

type SamplingMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type SamplingService struct {
	Enabled bool
}

func NewSamplingService(enabled bool) *SamplingService {
	return &SamplingService{Enabled: enabled}
}
