package plan

import (
	"time"

	"github.com/cruciblehq/protocol/pkg/codec"
)

// State represents the current state of a deployment.
type State struct {
	Plan       string            `field:"plan"`        // Path or ID of the plan that was deployed
	DeployedAt time.Time         `field:"deployed_at"` // When this state was created
	Services   []DeployedService `field:"services"`    // Actually deployed services
	Widgets    []DeployedWidget  `field:"widgets"`     // Actually deployed widgets
	Gateway    *DeployedGateway  `field:"gateway"`     // Actually deployed gateway
}

// DeployedService represents a service that has been deployed.
type DeployedService struct {
	// From plan
	Reference   string `field:"reference"`
	Resolved    string `field:"resolved"`
	Prefix      string `field:"prefix"`
	ImageDigest string `field:"image_digest"`

	// Runtime info
	Provider     string            `field:"provider"`      // aws, docker, kubernetes
	ResourceID   string            `field:"resource_id"`   // ECS task ARN, container ID, etc.
	ResourceType string            `field:"resource_type"` // ecs-service, docker-container, etc.
	Status       string            `field:"status"`        // running, stopped, failed
	Metadata     map[string]string `field:"metadata"`      // Provider-specific metadata
}

// DeployedWidget represents a widget that has been deployed.
type DeployedWidget struct {
	Reference string            `field:"reference"`
	Resolved  string            `field:"resolved"`
	Name      string            `field:"name"`
	Bundle    string            `field:"bundle"`
	Metadata  map[string]string `field:"metadata"`
}

// DeployedGateway represents a gateway that has been deployed.
type DeployedGateway struct {
	Type         string            `field:"type"`
	Listen       string            `field:"listen"`
	Provider     string            `field:"provider"`      // aws-alb, nginx, etc.
	ResourceID   string            `field:"resource_id"`   // ALB ARN, container ID, etc.
	ResourceType string            `field:"resource_type"` // alb, nginx-container, etc.
	Status       string            `field:"status"`
	Routes       []DeployedRoute   `field:"routes"`
	Metadata     map[string]string `field:"metadata"`
}

// DeployedRoute represents a deployed gateway route.
type DeployedRoute struct {
	Prefix   string `field:"prefix"`
	Upstream string `field:"upstream"`
	Service  string `field:"service"`
	RuleID   string `field:"rule_id"` // Provider-specific rule identifier
}

// Write saves the state to a file.
func (s *State) Write(path string) error {
	return codec.EncodeFile(path, "field", s)
}

// ReadState loads a state from a file.
func ReadState(path string) (*State, error) {
	var s State
	if _, err := codec.DecodeFile(path, "field", &s); err != nil {
		return nil, err
	}
	return &s, nil
}
