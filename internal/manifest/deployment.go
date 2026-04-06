package manifest

import "time"

// Represents deployment metadata.
type Deployment struct {
	DeployedAt time.Time `codec:"deployed_at"` // Timestamp of the most recent deployment.
}
