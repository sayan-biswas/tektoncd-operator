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

// Package openshiftbuild installs the OpenShift Builds components (Shipwright Build and
// the Shared Resource CSI Driver) whose desired state is expressed through the
// OpenShiftBuild CRD.
//
// TektonConfig's OpenShift extension creates and updates a singleton
// OpenShiftBuild CR from spec.platforms.openshift.builds; this controller
// reconciles that CR and installs the components through TektonInstallerSets
// owned by the OpenShiftBuild instance.
//
// The desired-state API types (OpenShiftBuildSpec and friends) are owned by this
// operator's own API package (copied from the upstream
// github.com/redhat-openshift-builds/operator), and the manifests that are
// installed are the ones shipped by that same operator (bundled into the
// operator payload under KO_DATA_PATH/openshift-builds). The install logic itself
// cannot be reused from that operator because it lives in an internal/ package
// and is built on controller-runtime with direct manifestival applies; this
// operator applies everything through TektonInstallerSets instead.
package openshiftbuild

import (
	"context"
	"os"
	"path/filepath"

	manifestivalclient "github.com/manifestival/client-go-client"
	"github.com/manifestival/manifestival"
	"github.com/tektoncd/operator/pkg/apis/operator/v1alpha1"
	operatorclient "github.com/tektoncd/operator/pkg/client/injection/client"
	openshiftbuildinformer "github.com/tektoncd/operator/pkg/client/injection/informers/operator/v1alpha1/openshiftbuild"
	tektoninstallersetinformer "github.com/tektoncd/operator/pkg/client/injection/informers/operator/v1alpha1/tektoninstallerset"
	openshiftbuildreconciler "github.com/tektoncd/operator/pkg/client/injection/reconciler/operator/v1alpha1/openshiftbuild"
	"github.com/tektoncd/operator/pkg/reconciler/common"
	tektoninstallersetclient "github.com/tektoncd/operator/pkg/reconciler/kubernetes/tektoninstallerset/client"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/logging"
)

// NewController initializes the controller and is called by the generated code.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := logging.FromContext(ctx)

	mfClient, err := manifestivalclient.NewClient(injection.GetConfig(ctx))
	if err != nil {
		logger.Fatalw("error creating manifestival client from injected config", zap.Error(err))
	}

	operatorVer, err := common.OperatorVersion(ctx)
	if err != nil {
		logger.Fatal(err)
	}

	base := filepath.Join(os.Getenv(common.KoEnvKey), ManifestPath)
	shipwrightManifest := loadManifest(logger, mfClient, filepath.Join(base, ShipwrightBuildManifestPath))
	sharedResourceManifest := loadManifest(logger, mfClient, filepath.Join(base, SharedResourceManifestPath))

	tisClient := operatorclient.Get(ctx).OperatorV1alpha1().TektonInstallerSets()

	c := &Reconciler{
		installerSetClient:     tektoninstallersetclient.NewInstallerSetClient(tisClient, operatorVer, operatorVer, v1alpha1.KindOpenShiftBuild, nil),
		shipwrightManifest:     shipwrightManifest,
		sharedResourceManifest: sharedResourceManifest,
		openshiftBuildsVersion: operatorVer,
	}
	impl := openshiftbuildreconciler.NewImpl(ctx, c)

	logger.Debug("Setting up event handlers for OpenShiftBuild")

	if _, err := openshiftbuildinformer.Get(ctx).Informer().AddEventHandler(controller.HandleAll(impl.Enqueue)); err != nil {
		logger.Panicf("Couldn't register OpenShiftBuild informer event handler: %w", err)
	}

	if _, err := tektoninstallersetinformer.Get(ctx).Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterController(&v1alpha1.OpenShiftBuild{}),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	}); err != nil {
		logger.Panicf("Couldn't register TektonInstallerSet informer event handler: %w", err)
	}

	return impl
}

// loadManifest reads all YAML under dir into a manifest bound to mfClient. A
// missing directory yields an empty manifest instead of a fatal error.
func loadManifest(logger *zap.SugaredLogger, mfClient manifestival.Client, dir string) *manifestival.Manifest {
	manifest, err := manifestival.ManifestFrom(manifestival.Slice{}, manifestival.UseClient(mfClient))
	if err != nil {
		logger.Fatalw("error creating initial manifest", zap.Error(err))
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Warnf("openshift-builds manifest directory not found, skipping: %s", dir)
		return &manifest
	}
	if err := common.AppendManifest(&manifest, dir); err != nil {
		logger.Fatalf("failed to load openshift-builds manifest from %s: %v", dir, err)
	}
	return &manifest
}
