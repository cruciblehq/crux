package manifest

// AWS compute configuration.
//
// Specifies EC2 instance settings for AWS deployments.
type ComputeAWS struct {

	// EC2 instance type (e.g. "t3.micro").
	InstanceType string `codec:"instance_type"`

	// AWS region for the instance.
	Region string `codec:"region,omitempty"`
}
