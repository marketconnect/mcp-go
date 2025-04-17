package protocol

type MessageType int

const (
	Unknown MessageType = iota
	Request
	Notification
	Response
	Batch
)

func (m MessageType) String() string {
	switch m {
	case Request:
		return "request"
	case Notification:
		return "notification"
	case Response:
		return "response"
	case Batch:
		return "batch"
	default:
		return "unknown"
	}
}

type TypedMessage[T IDConstraint] struct {
	Type  MessageType
	Value any
}

func (m *TypedMessage[T]) AsRequest() (*JSONRPCRequest[T], bool) {
	if m.Type != Request {
		return nil, false
	}
	req, ok := m.Value.(*JSONRPCRequest[T])
	return req, ok
}

func (m *TypedMessage[T]) AsResponse() (*JSONRPCResponse[T, any], bool) {
	if m.Type != Response {
		return nil, false
	}
	resp, ok := m.Value.(*JSONRPCResponse[T, any])
	return resp, ok
}

func (m *TypedMessage[T]) AsNotification() (*JSONRPCNotification, bool) {
	if m.Type != Notification {
		return nil, false
	}
	notif, ok := m.Value.(*JSONRPCNotification)
	return notif, ok
}

func (m *TypedMessage[T]) AsBatch() ([]*TypedMessage[T], bool) {
	if m.Type != Batch {
		return nil, false
	}
	batch, ok := m.Value.([]*TypedMessage[T])
	return batch, ok
}
