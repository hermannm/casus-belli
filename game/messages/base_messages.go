package messages

// Embedded struct in all message types.
type Base struct {
	// Allows for correctly identifying incoming messages.
	Type string `json:"type"`
}
