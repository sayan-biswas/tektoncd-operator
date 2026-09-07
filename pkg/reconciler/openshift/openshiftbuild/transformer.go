/*
Copyright 2026 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package openshiftbuild

import (
	"context"

	"github.com/manifestival/manifestival"
	"github.com/tektoncd/operator/pkg/apis/operator/v1alpha1"
	"github.com/tektoncd/operator/pkg/reconciler/common"
	openshift "github.com/tektoncd/operator/pkg/reconciler/openshift/common"
	"github.com/tektoncd/operator/pkg/reconciler/kubernetes/tektoninstallerset/client"
)

const (
	// componentLabel marks the installer sets created by this reconciler so they
	// can be identified independently of the owning OpenShiftBuild.
	componentLabel = "operator.tekton.dev/component"
	componentValue = "openshift-builds"
)

// filterAndTransform installs the Builds manifests into the openshift-builds
// namespace. The upstream operator applies the same InjectNamespace transform;
// removeRunAsUserRunAsGroup mirrors its behaviour of letting the OpenShift SCC
// controller assign the runAsUser/runAsGroup so the pods pass the restricted
// SCC admission.
func filterAndTransform() client.FilterAndTransform {
	return func(ctx context.Context, manifest *manifestival.Manifest, tektonComponent v1alpha1.TektonComponent) (*manifestival.Manifest, error) {
		images := common.ImageRegistryDomainOverride(common.ToLowerCaseKeys(common.ImagesFromEnv(common.BuildsImagePrefix)))
		transformed, err := manifest.Transform(
			manifestival.InjectNamespace(TargetNamespace),
			common.InjectOperandNameLabelOverwriteExisting(""),
			common.DeploymentImages(images),
			common.DeploymentEnvVarKubernetesMinVersion(),
			openshift.RemoveRunAsGroup(),
			openshift.RemoveRunAsUser(),
		)
		if err != nil {
			return nil, err
		}
		return &transformed, nil
	}
}

/*
// removeRunAsUserRunAsGroup drops runAsUser/runAsGroup from Pod and container
// security contexts so that OpenShift assigns them from the namespace's SCC
// range. It mirrors RemoveRunAsUserRunAsGroup in the upstream openshift-builds
// operator's internal/common transformer set.
func removeRunAsUserRunAsGroup(u *unstructured.Unstructured) error {
	switch u.GetKind() {
	case "Deployment", "StatefulSet", "DaemonSet", "Job":
	default:
		return nil
	}

	podSpec, found, err := unstructured.NestedMap(u.Object, "spec", "template", "spec")
	if err != nil || !found {
		return err
	}

	// Pod-level security context.
	if sc, ok := podSpec["securityContext"].(map[string]interface{}); ok {
		delete(sc, "runAsUser")
		delete(sc, "runAsGroup")
	}

	// Container-level security contexts (containers + initContainers).
	for _, key := range []string{"containers", "initContainers"} {
		containers, ok := podSpec[key].([]interface{})
		if !ok {
			continue
		}
		for _, c := range containers {
			container, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			if sc, ok := container["securityContext"].(map[string]interface{}); ok {
				delete(sc, "runAsUser")
				delete(sc, "runAsGroup")
			}
		}
	}

	return unstructured.SetNestedMap(u.Object, podSpec, "spec", "template", "spec")
}

 */
