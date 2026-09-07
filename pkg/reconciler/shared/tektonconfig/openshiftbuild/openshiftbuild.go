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

// Package openshiftbuild reconciles the singleton OpenShiftBuild CR from the
// TektonConfig spec.platforms.openshift.builds field. TektonConfig owns the CR;
// the standalone OpenShiftBuild controller reconciles it and installs the
// components.
package openshiftbuild

import (
	"context"
	"fmt"
	"reflect"

	"github.com/tektoncd/operator/pkg/apis/operator/v1alpha1"
	operatorclient "github.com/tektoncd/operator/pkg/client/clientset/versioned/typed/operator/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// targetNamespace is the namespace where the OpenShift Builds components are
// installed. It matches the namespace used by the upstream openshift-builds
// operator.
const targetNamespace = "openshift-builds"

// EnsureOpenShiftBuildExists creates or updates the singleton OpenShiftBuild CR
// to match the desired spec carried by TektonConfig.
func EnsureOpenShiftBuildExists(ctx context.Context, client operatorclient.OpenShiftBuildInterface, config *v1alpha1.TektonConfig, version string, spec *v1alpha1.OpenShiftBuildSpec) (*v1alpha1.OpenShiftBuild, error) {
	ob, err := client.Get(ctx, v1alpha1.OpenShiftBuildResourceName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
		if _, err = createOpenShiftBuild(ctx, client, config, version, spec); err != nil {
			return nil, err
		}
		return nil, v1alpha1.RECONCILE_AGAIN_ERR
	}

	ob, err = updateOpenShiftBuild(ctx, ob, config, client, version, spec)
	if err != nil {
		return nil, err
	}

	if !ob.Status.IsReady() {
		return nil, v1alpha1.RECONCILE_AGAIN_ERR
	}

	return ob, nil
}

func createOpenShiftBuild(ctx context.Context, client operatorclient.OpenShiftBuildInterface, config *v1alpha1.TektonConfig, version string, spec *v1alpha1.OpenShiftBuildSpec) (*v1alpha1.OpenShiftBuild, error) {
	ownerReference := *metav1.NewControllerRef(config, config.GroupVersionKind())

	//spec := desiredSpec(openShiftBuildSpec)
	ob := &v1alpha1.OpenShiftBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:            v1alpha1.OpenShiftBuildResourceName,
			OwnerReferences: []metav1.OwnerReference{ownerReference},
			Labels: map[string]string{
				v1alpha1.ReleaseVersionKey: version,
			},
		},
		Spec: *spec,
	}

	if _, err := client.Create(ctx, ob, metav1.CreateOptions{}); err != nil {
		return nil, err
	}
	return ob, nil
}

func updateOpenShiftBuild(ctx context.Context, build *v1alpha1.OpenShiftBuild, config *v1alpha1.TektonConfig,
	client operatorclient.OpenShiftBuildInterface, version string, spec *v1alpha1.OpenShiftBuildSpec,
) (*v1alpha1.OpenShiftBuild, error) {
	updated := false

	//spec := desiredSpec(openShiftBuildSpec)
	if !reflect.DeepEqual(build.Spec.Shipwright, spec.Shipwright) {
		build.Spec.Shipwright = spec.Shipwright
		updated = true
	}
	if !reflect.DeepEqual(build.Spec.SharedResource, spec.SharedResource) {
		build.Spec.SharedResource = spec.SharedResource
		updated = true
	}
	if build.Spec.TargetNamespace != spec.TargetNamespace {
		build.Spec.TargetNamespace = spec.TargetNamespace
		updated = true
	}

	if build.ObjectMeta.OwnerReferences == nil {
		ownerRef := *metav1.NewControllerRef(config, config.GroupVersionKind())
		build.ObjectMeta.OwnerReferences = []metav1.OwnerReference{ownerRef}
		updated = true
	}

	if build.ObjectMeta.Labels == nil {
		build.ObjectMeta.Labels = map[string]string{}
	}
	if build.ObjectMeta.Labels[v1alpha1.ReleaseVersionKey] != version {
		build.ObjectMeta.Labels[v1alpha1.ReleaseVersionKey] = version
		updated = true
	}

	if updated {
		if _, err := client.Update(ctx, build, metav1.UpdateOptions{}); err != nil {
			return nil, err
		}
		return nil, v1alpha1.RECONCILE_AGAIN_ERR
	}

	return build, nil
}

// desiredSpec returns a copy of the TektonConfig-provided spec with the target
// namespace pinned to the openshift-builds namespace.
func desiredSpec(spec *v1alpha1.OpenShiftBuildSpec) v1alpha1.OpenShiftBuildSpec {
	obs := v1alpha1.OpenShiftBuildSpec{}
	if spec != nil {
		obs = *spec.DeepCopy()
	}
	obs.TargetNamespace = targetNamespace
	return obs
}

// EnsureOpenShiftBuildCRNotExists deletes the singleton OpenShiftBuild CR if it
// exists.
func EnsureOpenShiftBuildCRNotExists(ctx context.Context, client operatorclient.OpenShiftBuildInterface) error {
	if _, err := client.Get(ctx, v1alpha1.OpenShiftBuildResourceName, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := client.Delete(ctx, v1alpha1.OpenShiftBuildResourceName, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("OpenShiftBuild %q failed to delete: %v", v1alpha1.OpenShiftBuildResourceName, err)
	}
	return v1alpha1.RECONCILE_AGAIN_ERR
}
