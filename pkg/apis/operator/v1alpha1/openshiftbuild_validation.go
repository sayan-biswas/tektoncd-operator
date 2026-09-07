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
	"context"
	"fmt"

	"knative.dev/pkg/apis"
)

func (b *OpenShiftBuild) Validate(ctx context.Context) (errs *apis.FieldError) {
	if apis.IsInDelete(ctx) {
		return nil
	}

	if b.GetName() != OpenShiftBuildResourceName {
		errMsg := fmt.Sprintf("metadata.name, Only one instance of OpenShiftBuild is allowed by name, %s", OpenShiftBuildResourceName)
		return errs.Also(apis.ErrInvalidValue(b.GetName(), errMsg))
	}

	//errs = errs.Also(validateBuildState(b.Spec.Shipwright, "spec.shipwright"))
	//errs = errs.Also(validateSharedResourceState(b.Spec.SharedResource, "spec.sharedResource"))
	return errs
}

//func validateBuildState(shipwright *Shipwright, path string) (errs *apis.FieldError) {
//	if shipwright == nil || shipwright.Build == nil {
//		return nil
//	}
//	return validateState(shipwright.Build.State, path+".build.state")
//}

//func validateSharedResourceState(sharedResource *SharedResource, path string) (errs *apis.FieldError) {
//	if sharedResource == nil {
//		return nil
//	}
//	return validateState(sharedResource.State, path+".state")
//}

//func validateState(state BuildState, path string) *apis.FieldError {
//	switch state {
//	case "", BuildEnabled, BuildDisabled:
//		return nil
//	default:
//		return apis.ErrInvalidValue(state, path)
//	}
//}
