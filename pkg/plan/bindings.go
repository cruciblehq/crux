package plan

import (
	"github.com/cruciblehq/protocol/pkg/plan"
)

// Creates bindings between services and compute resources.
//
// Binds all services to the single shared compute instance. Each service gets
// its own binding entry, enabling the deployment system to provision all
// services on the same infrastructure.
func bind(p *plan.Plan) {
	computeID := "main-compute"

	for _, service := range p.Services {
		binding := plan.Binding{
			Service: service.ID,
			Compute: computeID,
		}
		p.Bindings = append(p.Bindings, binding)
	}
}
