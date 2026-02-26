package resource

import (
	"context"

	"github.com/cruciblehq/crex"
	"github.com/cruciblehq/spec/manifest"
)

// [Runner] for resource types that are not recognized.
type runnerStub struct {
	resourceType string
}

func (u *runnerStub) Build(_ context.Context, _ manifest.Manifest, _ string) (*BuildResult, error) {
	return nil, runnerStubNotSupportedError(u.resourceType)
}

func (u *runnerStub) Start(_ context.Context, _ manifest.Manifest, _ string) error {
	return runnerStubNotSupportedError(u.resourceType)
}

func (u *runnerStub) Stop(_ context.Context, _ manifest.Manifest) error {
	return runnerStubNotSupportedError(u.resourceType)
}

func (u *runnerStub) Restart(_ context.Context, _ manifest.Manifest, _ string) error {
	return runnerStubNotSupportedError(u.resourceType)
}

func (u *runnerStub) Reset(_ context.Context, _ manifest.Manifest, _ string) error {
	return runnerStubNotSupportedError(u.resourceType)
}

func (u *runnerStub) Destroy(_ context.Context, _ manifest.Manifest) error {
	return runnerStubNotSupportedError(u.resourceType)
}

func (u *runnerStub) Exec(_ context.Context, _ manifest.Manifest, _ []string) (*ExecResult, error) {
	return nil, runnerStubNotSupportedError(u.resourceType)
}

func (u *runnerStub) Status(_ context.Context, _ manifest.Manifest) (*StatusResult, error) {
	return nil, runnerStubNotSupportedError(u.resourceType)
}

func (u *runnerStub) Pack(_ context.Context, _ manifest.Manifest, _, _, _ string) (*PackResult, error) {
	return nil, runnerStubNotSupportedError(u.resourceType)
}

func (u *runnerStub) Push(_ context.Context, _ manifest.Manifest, _ string) error {
	return runnerStubNotSupportedError(u.resourceType)
}

func runnerStubNotSupportedError(resourceType string) error {
	return crex.Wrapf(ErrInvalidResourceType, "resource type %q is not supported", resourceType)
}
