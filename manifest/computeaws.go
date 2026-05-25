package manifest

import "github.com/cruciblehq/crux/crex"

// AWS EC2 instance configuration.
//
// This is the configuration for a compute unit with AWS as the provider. The
// compute unit is backed by an EC2 instance with the specified configuration.
// The executor provisions the instance and schedules services onto it.
type ComputeAWS struct {

	// EC2 instance type (e.g. "t3.micro").
	//
	// The instance type determines the hardware resources allocated to the
	// compute unit. A list of available instance types can be found in the
	// AWS documentation: https://aws.amazon.com/ec2/instance-types/.
	InstanceType string `codec:"instance_type"`

	// AWS region (e.g. "us-east-1").
	//
	// The region determines the geographic location of the compute unit. A
	// list of available regions can be found in the AWS documentation:
	// https://aws.amazon.com/about-aws/global-infrastructure/regions_az/.
	Region string `codec:"region,omitempty"`

	// Availability zone within the region (e.g. "us-east-1a").
	//
	// The availability zone determines the specific data center where the
	// compute unit is provisioned. If not specified, the executor may choose
	// any availability zone within the specified region. A list of available
	// availability zones can be found in the AWS documentation:
	// https://aws.amazon.com/about-aws/global-infrastructure/regions_az/.
	AvailabilityZone string `codec:"availability_zone,omitempty"`

	// Hardware tenancy of the instance.
	//
	// One of "default" (shared hardware), "dedicated" (single-tenant hardware),
	// or "host" (dedicated physical host). Empty means "default".
	Tenancy string `codec:"tenancy,omitempty"`
}

// Validates the instance type and tenancy.
func (c *ComputeAWS) Validate() error {
	if c.InstanceType == "" {
		return ErrMissingComputeInstanceType
	}
	switch c.Tenancy {
	case "", "default", "dedicated", "host":
		return nil
	default:
		return crex.Wrapf(ErrInvalidTenancy, "unknown tenancy %q", c.Tenancy)
	}
}
