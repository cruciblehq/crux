package registry

import "net/url"

// Returns the path for the namespaces collection.
func namespacesPath() string {
	return "/namespaces"
}

// Returns the path for a single namespace.
func namespacePath(ns string) string {
	p, _ := url.JoinPath(namespacesPath(), ns)
	return p
}

// Returns the path for the resources collection within a namespace.
func resourcesPath(ns string) string {
	p, _ := url.JoinPath(namespacePath(ns), "resources")
	return p
}

// Returns the path for a single resource within a namespace.
func resourcePath(ns, res string) string {
	p, _ := url.JoinPath(resourcesPath(ns), res)
	return p
}

// Returns the path for the versions collection of a resource.
func versionsPath(ns, res string) string {
	p, _ := url.JoinPath(resourcePath(ns, res), "versions")
	return p
}

// Returns the path for a single version of a resource.
func versionPath(ns, res, ver string) string {
	p, _ := url.JoinPath(versionsPath(ns, res), ver)
	return p
}

// Returns the path for the archive of a specific version.
func archivePath(ns, res, ver string) string {
	p, _ := url.JoinPath(versionPath(ns, res, ver), "archive")
	return p
}

// Returns the path for the channels collection of a resource.
func channelsPath(ns, res string) string {
	p, _ := url.JoinPath(resourcePath(ns, res), "channels")
	return p
}

// Returns the path for a single channel of a resource.
func channelPath(ns, res, ch string) string {
	p, _ := url.JoinPath(channelsPath(ns, res), ch)
	return p
}
