package manifest

import "time"

// Represents deployment metadata.
type Deployment struct {
	DeployedAt time.Time `json:"deployed_at"` // Timestamp of the most recent deployment.
}
