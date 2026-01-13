package plan

import (
	"time"

	"github.com/cruciblehq/protocol/pkg/codec"
)

// Plan represents a resolved deployment plan.
type Plan struct {
	Blueprint string            `field:"blueprint"`
	Target    string            `field:"target"`
	CreatedAt time.Time         `field:"created_at"`
	Services  []ResolvedService `field:"services"`
	Widgets   []ResolvedWidget  `field:"widgets"`
	Gateway   *Gateway          `field:"gateway"`
}

// ResolvedService is a service with all references resolved.
type ResolvedService struct {
	Reference   string      `field:"reference"`
	Resolved    string      `field:"resolved"`
	Prefix      string      `field:"prefix"`
	ImageDigest string      `field:"image_digest"`
	Deployment  *Deployment `field:"deployment"`
}

// Deployment contains target-specific deployment info.
type Deployment struct {
	Type      string               `field:"type"`
	Container *ContainerDeployment `field:"container"`
}

// ContainerDeployment defines Docker container deployment.
type ContainerDeployment struct {
	Name  string `field:"name"`
	Port  int    `field:"port"`
	Image string `field:"image"`
}

// ResolvedWidget is a widget with references resolved.
type ResolvedWidget struct {
	Reference string `field:"reference"`
	Resolved  string `field:"resolved"`
	Name      string `field:"name"`
	Bundle    string `field:"bundle"`
}

// Gateway defines API gateway configuration.
type Gateway struct {
	Type   string         `field:"type"`
	Listen string         `field:"listen"`
	Routes []GatewayRoute `field:"routes"`
}

// GatewayRoute maps a prefix to a service.
type GatewayRoute struct {
	Prefix   string `field:"prefix"`
	Upstream string `field:"upstream"`
	Service  string `field:"service"`
}

// Write saves the plan to a file.
//
// The file format is inferred from the path extension (.json, .yaml, .toml).
func (p *Plan) Write(path string) error {
	return codec.EncodeFile(path, "field", p)
}

// Read loads a plan from a file.
//
// The file format is inferred from the path extension (.json, .yaml, .toml).
func Read(path string) (*Plan, error) {
	var p Plan
	if _, err := codec.DecodeFile(path, "field", &p); err != nil {
		return nil, err
	}
	return &p, nil
}
