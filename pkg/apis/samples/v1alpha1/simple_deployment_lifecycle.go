/*
Copyright 2020 The Knative Authors

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

var simpleDeploymentCondSet = apis.NewLivingConditionSet(TaskrunSucceeded)

// GetGroupVersionKind implements kmeta.OwnerRefable
func (*SimpleDeployment) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("SimpleDeployment")
}

func (*SimpleDeployment) GroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("SimpleDeployment")
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (d *SimpleDeployment) GetConditionSet() apis.ConditionSet {
	return simpleDeploymentCondSet
}

// InitializeConditions sets the initial values to the conditions.
func (ds *SimpleDeploymentStatus) InitializeConditions() {
	simpleDeploymentCondSet.Manage(ds).InitializeConditions()
}

// MarkPodsNotReady makes the SimpleDeployment be not ready.
func (ds *SimpleDeploymentStatus) MarkTRFailed(msg string) {
	ds.MarkNotReady("something is fishy !")
	simpleDeploymentCondSet.Manage(ds).MarkFalse(
		TaskrunSucceeded,
		"TaskRun failed",
		msg)
}

func (ds *SimpleDeploymentStatus) MarkTRReady() {
	simpleDeploymentCondSet.Manage(ds).MarkTrue(TaskrunSucceeded)
}

func (ds *SimpleDeploymentStatus) MarkReady() {
	simpleDeploymentCondSet.Manage(ds).MarkTrue(SimpleDeploymentConditionReady)
}

func (ds *SimpleDeploymentStatus) MarkNotReady(msg string) {
	simpleDeploymentCondSet.Manage(ds).MarkFalse(
		apis.ConditionReady,
		"Error",
		"Ready: %s", msg)
}
