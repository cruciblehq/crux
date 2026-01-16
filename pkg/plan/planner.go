package plan

import (
	"context"
	"fmt"
	"time"

	"github.com/cruciblehq/protocol/pkg/blueprint"
)

// Generates a plan from a blueprint.
// If state is provided, generates an incremental plan based on current deployment state.
func Build(ctx context.Context, bp *blueprint.Blueprint, blueprintPath string, state *State) (*Plan, error) {
	p := &Plan{
		CreatedAt: time.Now(),
		Services:  make([]ResolvedService, 0, len(bp.Services)),
	}

	// For now, just create a simple plan for local Docker deployment
	// TODO: Actually resolve references, fetch manifests, validate, etc.

	basePort := 8080
	for i, svc := range bp.Services {
		// Derive container name from prefix (strip leading slashes and convert to valid name)
		containerName := svc.Prefix[1:] // Remove leading /
		if containerName == "" {
			containerName = "service"
		}

		resolved := ResolvedService{
			Reference: svc.Reference,
			Resolved:  svc.Reference + " (unresolved)", // TODO: resolve version
			Prefix:    svc.Prefix,
			Deployment: &Deployment{
				Type: "docker",
				Container: &ContainerDeployment{
					Name:  containerName,
					Port:  basePort + i,
					Image: fmt.Sprintf("dist/%s-image.tar", containerName),
				},
			},
		}
		p.Services = append(p.Services, resolved)
	}

	// Build gateway routes
	if len(bp.Services) > 0 {
		p.Gateway = &Gateway{
			Type:   "reverse-proxy",
			Listen: ":8000",
			Routes: make([]GatewayRoute, 0, len(bp.Services)),
		}

		for _, svc := range p.Services {
			route := GatewayRoute{
				Prefix:   svc.Prefix,
				Upstream: fmt.Sprintf("http://localhost:%d", svc.Deployment.Container.Port),
				Service:  svc.Prefix, // Use prefix as service identifier
			}
			p.Gateway.Routes = append(p.Gateway.Routes, route)
		}
	}

	return p, nil
}
