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
	"errors"
	"fmt"

	"github.com/manifestival/manifestival"
	"github.com/tektoncd/operator/pkg/apis/operator/v1alpha1"
	operatorinformer "github.com/tektoncd/operator/pkg/client/informers/externalversions/operator/v1alpha1"
	openshiftbuildreconciler "github.com/tektoncd/operator/pkg/client/injection/reconciler/operator/v1alpha1/openshiftbuild"
	"github.com/tektoncd/operator/pkg/reconciler/common"
	tektoninstallersetclient "github.com/tektoncd/operator/pkg/reconciler/kubernetes/tektoninstallerset/client"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
)

// Reconciler implements controller.Reconciler for OpenShiftBuild resources. It
// installs the OpenShift Builds components (Shipwright Build and the Shared
// Resource CSI Driver) through TektonInstallerSets owned by the OpenShiftBuild
// instance.
type Reconciler struct {
	installerSetClient *tektoninstallersetclient.InstallerSetClient
	pipelineInformer   operatorinformer.TektonPipelineInformer
	shipwrightManifest *manifestival.Manifest
	sharedResourceManifest *manifestival.Manifest
	// version of OpenShift Builds which we are installing
	openshiftBuildsVersion string
}

// Check that our Reconciler implements controller.Reconciler
var _ openshiftbuildreconciler.Interface = (*Reconciler)(nil)

// ReconcileKind compares the actual state with the desired, and attempts to
// converge the two.
func (r *Reconciler) ReconcileKind(ctx context.Context, ob *v1alpha1.OpenShiftBuild) reconciler.Event {
	logger := logging.FromContext(ctx).With("name", ob.GetName())
	ob.Status.InitializeConditions()
	ob.Status.SetVersion(r.openshiftBuildsVersion)

	if ob.GetName() != v1alpha1.OpenShiftBuildResourceName {
		msg := fmt.Sprintf("Resource ignored, Expected Name: %s, Got Name: %s",
			v1alpha1.OpenShiftBuildResourceName, ob.GetName())
		logger.Error(msg)
		ob.Status.MarkNotReady(msg)
		return nil
	}

	//Make sure TektonPipeline is installed before proceeding with OpenShiftPipelinesAsCode
	if _, err := common.PipelineReady(r.pipelineInformer); err != nil {
		if err.Error() == common.PipelineNotReady || errors.Is(err, v1alpha1.DEPENDENCY_UPGRADE_PENDING_ERR) {
			ob.Status.MarkDependencyInstalling("tekton-pipelines is still installing")
			return v1alpha1.REQUEUE_EVENT_AFTER
		}
		ob.Status.MarkDependencyMissing("tekton-pipelines does not exist")
		return err
	}

	ob.Status.MarkDependenciesInstalled()
	ob.Status.MarkPreReconcilerComplete()

	if err := r.reconcileShipwright(ctx, ob); err != nil {
		return r.handleInstallError(ctx, ob, shipwrightBuildOperandName, err)
	}

	if err := r.reconcileSharedResource(ctx, ob); err != nil {
		return r.handleInstallError(ctx, ob, sharedResourceComponentName, err)
	}

	ob.Status.MarkInstallerSetAvailable()
	ob.Status.MarkInstallerSetReady()
	ob.Status.MarkPostReconcilerComplete()
	return nil
}

func (r *Reconciler) handleInstallError(ctx context.Context, ob *v1alpha1.OpenShiftBuild, name string, err error) reconciler.Event {
	if errors.Is(err, v1alpha1.REQUEUE_EVENT_AFTER) {
		return err
	}
	msg := fmt.Sprintf("%s installation failed: %s", name, err.Error())
	logging.FromContext(ctx).Error(msg)
	ob.Status.MarkInstallerSetNotReady(msg)
	return nil
}

func (r *Reconciler) reconcileShipwright(ctx context.Context, ob *v1alpha1.OpenShiftBuild) error {
	if !isShipwrightBuildEnabled(&ob.Spec) {
		return r.installerSetClient.CleanupCustomSet(ctx, shipwrightBuildOperandName)
	}
	return r.installerSetClient.CustomSet(ctx, ob, shipwrightBuildOperandName, r.shipwrightManifest, filterAndTransform(), componentLabels())
}

func (r *Reconciler) reconcileSharedResource(ctx context.Context, ob *v1alpha1.OpenShiftBuild) error {
	if !isSharedResourceEnabled(&ob.Spec) {
		return r.installerSetClient.CleanupCustomSet(ctx, sharedResourceComponentName)
	}
	return r.installerSetClient.CustomSet(ctx, ob, sharedResourceComponentName, r.sharedResourceManifest, filterAndTransform(), componentLabels())
}

func isShipwrightBuildEnabled(spec *v1alpha1.OpenShiftBuildSpec) bool {
	return spec != nil && spec.Shipwright != nil && spec.Shipwright.Build != nil &&
		spec.Shipwright.Build.Enable
}

func isSharedResourceEnabled(spec *v1alpha1.OpenShiftBuildSpec) bool {
	return spec != nil && spec.SharedResource != nil && spec.SharedResource.Enable
}

func componentLabels() map[string]string {
	return map[string]string{componentLabel: componentValue}
}
