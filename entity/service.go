package entity

import (
	"encoding/json"

	"github.com/pkg/errors"
)

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

// DecodeServices unmarshals services.
func DecodeServices(data []byte) (services []Service, err error) {

	services = []Service{}
	err = json.Unmarshal(data, &services)
	err = errors.Wrapf(err, "failed to unmarshal services from: %s", data)
	return
}
