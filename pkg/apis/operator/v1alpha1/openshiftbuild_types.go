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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// The OpenShiftBuild desired-state fields (Shipwright and SharedResource) are
// copied from github.com/redhat-openshift-builds/operator (api/v1alpha1) so that
// this operator owns the definition instead of importing the external module.
// The surrounding object/spec/status follow this operator's component
// conventions (CommonSpec + duckv1.Status), so OpenShiftBuild is reconciled by a
// standard genreconciler controller like every other Tekton component.

var (
	_ TektonComponent     = (*OpenShiftBuild)(nil)
	_ TektonComponentSpec = (*OpenShiftBuildSpec)(nil)
)

// BuildState defines the desired state of a Builds component.
// +kubebuilder:validation:Enum="Enabled";"Disabled"
type BuildState string

const (
	// BuildEnabled will install the component, including any additional custom
	// resource definitions.
	BuildEnabled BuildState = "Enabled"

	// BuildDisabled will remove the component, but may leave behind any custom
	// resource definitions.
	BuildDisabled BuildState = "Disabled"
)

// OpenShiftBuild is the Schema for the OpenShiftBuild API
// +genclient
// +genreconciler:krshapedlogic=false
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.version`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].message`
type OpenShiftBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenShiftBuildSpec   `json:"spec,omitempty"`
	Status OpenShiftBuildStatus `json:"status,omitempty"`
}

// GetSpec implements TektonComponent
func (b *OpenShiftBuild) GetSpec() TektonComponentSpec {
	return &b.Spec
}

// GetStatus implements TektonComponent
func (b *OpenShiftBuild) GetStatus() TektonComponentStatus {
	return &b.Status
}

// OpenShiftBuildSpec defines the desired state of Builds for OpenShift components.
type OpenShiftBuildSpec struct {
	CommonSpec `json:",inline"`

	// Config holds the configuration for resources created by OpenShiftBuild
	// +optional
	Config Config `json:"config,omitempty"`

	// options holds additions fields and these fields will be updated on the manifests
	// +optional
	Options AdditionalOptions `json:"options"`

	// Shipwright defines the desired state of Shipwright components.
	// +optional
	Shipwright *Shipwright `json:"shipwright,omitempty"`

	// SharedResource defines the desired state of the Shared Resource CSI Driver
	// components.
	// +optional
	SharedResource *SharedResource `json:"sharedResource,omitempty"`
}

// Shipwright defines the desired state of Shipwright components.
type Shipwright struct {
	// Build defines the desired state of Shipwright Build APIs, controllers, and
	// related components.
	// +optional
	Build *ShipwrightBuild `json:"build,omitempty"`
}

// ShipwrightBuild defines the desired state of Shipwright Builds.
type ShipwrightBuild struct {
	// enable or disable ShipwrightBuild Component
	// +kubebuilder:default=True
	Enable bool `json:"enable,omitempty"`

	// State defines the desired state of the Shipwright Build controller, APIs,
	// and related components. Must be one of Enabled or Disabled.
	// +kubebuilder:default="Enabled"
	//State BuildState `json:"state"`
}

// SharedResource defines the desired state of the Shared Resource CSI Driver and
// components.
type SharedResource struct {
	// enable or disable SharedResource Component
	// +kubebuilder:default=True
	Enable bool `json:"enable,omitempty"`

	// State defines the desired state of SharedResource CSI Driver, APIs, and
	// related components. Must be one of Enabled or Disabled.
	// +kubebuilder:default="Enabled"
	//State BuildState `json:"state"`
}

// OpenShiftBuildStatus defines the observed state of OpenShiftBuild.
type OpenShiftBuildStatus struct {
	duckv1.Status `json:",inline"`

	// The version of the installed release
	// +optional
	Version string `json:"version,omitempty"`
}

// OpenShiftBuildList contains a list of OpenShiftBuild
// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type OpenShiftBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenShiftBuild `json:"items"`
}
