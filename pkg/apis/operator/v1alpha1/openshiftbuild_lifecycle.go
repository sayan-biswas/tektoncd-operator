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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

var (
	_ TektonComponentStatus = (*OpenShiftBuildStatus)(nil)

	openShiftBuildCondSet = apis.NewLivingConditionSet(
		PreReconciler,
		DependenciesInstalled,
		InstallerSetAvailable,
		InstallerSetReady,
		PostReconciler,
	)
)

// GroupVersionKind returns SchemeGroupVersion of an OpenShiftBuild
func (b *OpenShiftBuild) GroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind(KindOpenShiftBuild)
}

func (b *OpenShiftBuild) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind(KindOpenShiftBuild)
}

// GetCondition returns the current condition of a given condition type
func (bs *OpenShiftBuildStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return openShiftBuildCondSet.Manage(bs).GetCondition(t)
}

// InitializeConditions initializes conditions of an OpenShiftBuildStatus
func (bs *OpenShiftBuildStatus) InitializeConditions() {
	openShiftBuildCondSet.Manage(bs).InitializeConditions()
}

// IsReady looks at the conditions returns true if they are all true.
func (bs *OpenShiftBuildStatus) IsReady() bool {
	return openShiftBuildCondSet.Manage(bs).IsHappy()
}

func (bs *OpenShiftBuildStatus) MarkNotReady(msg string) {
	openShiftBuildCondSet.Manage(bs).MarkFalse(
		apis.ConditionReady,
		"Error",
		"Ready: %s", msg)
}

func (bs *OpenShiftBuildStatus) MarkPreReconcilerComplete() {
	openShiftBuildCondSet.Manage(bs).MarkTrue(PreReconciler)
}

func (bs *OpenShiftBuildStatus) MarkInstallerSetAvailable() {
	openShiftBuildCondSet.Manage(bs).MarkTrue(InstallerSetAvailable)
}

func (bs *OpenShiftBuildStatus) MarkInstallerSetReady() {
	openShiftBuildCondSet.Manage(bs).MarkTrue(InstallerSetReady)
}

func (bs *OpenShiftBuildStatus) MarkInstallerSetNotAvailable(msg string) {
	bs.MarkNotReady("InstallerSet not ready")
	openShiftBuildCondSet.Manage(bs).MarkFalse(
		InstallerSetAvailable,
		"Error",
		"Installer set not ready: %s", msg)
}

func (bs *OpenShiftBuildStatus) MarkInstallerSetNotReady(msg string) {
	bs.MarkNotReady("InstallerSet not ready")
	openShiftBuildCondSet.Manage(bs).MarkFalse(
		InstallerSetReady,
		"Error",
		"Installer set not ready: %s", msg)
}

func (bs *OpenShiftBuildStatus) MarkPostReconcilerComplete() {
	openShiftBuildCondSet.Manage(bs).MarkTrue(PostReconciler)
}

// MarkDependenciesInstalled marks the DependenciesInstalled status as true.
func (bs *OpenShiftBuildStatus) MarkDependenciesInstalled() {
	openShiftBuildCondSet.Manage(bs).MarkTrue(DependenciesInstalled)
}

// MarkDependencyInstalling marks the DependenciesInstalled status as false
func (bs *OpenShiftBuildStatus) MarkDependencyInstalling(msg string) {
	openShiftBuildCondSet.Manage(bs).MarkFalse(
		DependenciesInstalled,
		"Installing",
		"Dependency installing: %s", msg)
}

// MarkDependencyMissing marks the DependenciesInstalled status as false
func (bs *OpenShiftBuildStatus) MarkDependencyMissing(msg string) {
	openShiftBuildCondSet.Manage(bs).MarkFalse(
		DependenciesInstalled,
		"Error",
		"Dependency missing: %s", msg)
}

func (bs *OpenShiftBuildStatus) MarkPreReconcilerFailed(msg string) {
	bs.MarkNotReady("PreReconciliation failed")
	openShiftBuildCondSet.Manage(bs).MarkFalse(
		PreReconciler,
		"Error",
		msg,
	)
}

func (bs *OpenShiftBuildStatus) MarkPostReconcilerFailed(msg string) {
	bs.MarkNotReady("PostReconciliation failed")
	openShiftBuildCondSet.Manage(bs).MarkFalse(
		PostReconciler,
		"Error",
		msg,
	)
}

// GetVersion gets the currently installed version of the component.
func (bs *OpenShiftBuildStatus) GetVersion() string {
	return bs.Version
}

// SetVersion sets the currently installed version of the component.
func (bs *OpenShiftBuildStatus) SetVersion(version string) {
	bs.Version = version
}
