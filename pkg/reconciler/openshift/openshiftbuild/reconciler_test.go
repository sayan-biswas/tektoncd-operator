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
	"testing"

	"github.com/tektoncd/operator/pkg/apis/operator/v1alpha1"
)

func shipwrightSpec(state bool) *v1alpha1.OpenShiftBuildSpec {
	return &v1alpha1.OpenShiftBuildSpec{
		Shipwright: &v1alpha1.Shipwright{
			Build: &v1alpha1.ShipwrightBuild{Enable: state},
		},
	}
}

func sharedResourceSpec(state bool) *v1alpha1.OpenShiftBuildSpec {
	return &v1alpha1.OpenShiftBuildSpec{
		SharedResource: &v1alpha1.SharedResource{Enable: state},
	}
}

func TestShipwrightEnabled(t *testing.T) {
	tests := []struct {
		name string
		spec *v1alpha1.OpenShiftBuildSpec
		want bool
	}{
		{"nil spec", nil, false},
		{"nil shipwright", &v1alpha1.OpenShiftBuildSpec{}, false},
		{"nil build", &v1alpha1.OpenShiftBuildSpec{Shipwright: &v1alpha1.Shipwright{}}, false},
		{"enabled", shipwrightSpec(true), true},
		{"disabled", shipwrightSpec(false), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isShipwrightBuildEnabled(tt.spec); got != tt.want {
				t.Errorf("shipwrightEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSharedResourceEnabled(t *testing.T) {
	tests := []struct {
		name string
		spec *v1alpha1.OpenShiftBuildSpec
		want bool
	}{
		{"nil spec", nil, false},
		{"nil sharedResource", &v1alpha1.OpenShiftBuildSpec{}, false},
		{"enabled", sharedResourceSpec(true), true},
		{"disabled", sharedResourceSpec(false), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSharedResourceEnabled(tt.spec); got != tt.want {
				t.Errorf("sharedResourceEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
