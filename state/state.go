package state

import "time"

// Represents the current state of a deployment.
//
// Records what resources have been deployed and their runtime identifiers.
// Used for incremental deployments and resource lifecycle management.
type State struct {
	Version    int        `json:"version"`
	Deployment Deployment `json:"deployment"`
	Services   []Service  `json:"services"`
}

// Represents deployment metadata.
type Deployment struct {
	DeployedAt time.Time `json:"deployed_at"`
}

// Represents a service that has been deployed.
type Service struct {
	ID         string `json:"id"`          // Service identifier, assigned at composition time.
	Reference  string `json:"reference"`   // Frozen service reference with exact version and digest.
	ResourceID string `json:"resource_id"` // Runtime resource identifier assigned during deployment.
}
