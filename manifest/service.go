package manifest

// Holds configuration specific to service resources.
//
// Service resources are backend components that provide functionality to other
// systems by exposing an API. Developers build service images using their tool
// of choice (Docker, buildah, etc.) and provide them as image.tar. Service
// resources have no additional configuration fields at the moment. The OCI
// image is provided as a pre-built image.tar file, and the manifest serves
// mainly to identify the resource type.
type Service struct {
}
