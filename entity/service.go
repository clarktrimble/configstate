package entity

// Capability represents something a service can do.
type Capability struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}

// Service is a service on the network.
type Service struct {
	Uri  string       `json:"uri"`
	Caps []Capability `json:"capabilities"`
}
