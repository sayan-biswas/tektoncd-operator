package openshiftbuild

import "path/filepath"

const (
	// TargetNamespace is the namespace where the OpenShift Builds components are
	// installed. It matches the namespace used by the upstream openshift-builds
	// operator.
	TargetNamespace = "openshift-builds"

	// ManifestPath is the KO_DATA_PATH subdirectory that holds the manifests
	// bundled from github.com/redhat-openshift-builds/operator.
	ManifestPath = "openshift-builds"

	// shipwrightBuildOperandName and sharedResourceComponentName are the CustomSet names for
	// each component; the resulting installer sets are named
	// openshiftbuild-shipwright-<rand> and openshiftbuild-sharedresource-<rand>.
	shipwrightBuildOperandName  = "shipwrightbuild"
	sharedResourceComponentName = "sharedresource"
)


var (
	ShipwrightBuildManifestPath         = filepath.Join(ManifestPath, "shipwright")
	SharedResourceManifestPath = filepath.Join(ManifestPath, "sharedresource")
)