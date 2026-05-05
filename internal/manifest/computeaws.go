package manifest

// AWS compute configuration.
//
// Specifies EC2 instance settings for AWS deployments.
type ComputeAWS struct {
	InstanceType string `json:"instance_type"`    // EC2 instance type (e.g. "t3.micro").
	Region       string `json:"region,omitempty"` // AWS region for the instance.
}
