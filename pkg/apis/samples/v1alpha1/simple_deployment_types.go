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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// SimpleDeployment is a Knative abstraction that encapsulates the interface by which Knative
// components express a desire to have a particular image cached.
//
// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SimpleDeployment struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the SimpleDeployment (from the client).
	// +optional
	Spec SimpleDeploymentSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the SimpleDeployment (from the controller).
	// +optional
	Status SimpleDeploymentStatus `json:"status,omitempty"`
}

var (
	// Check that AddressableService can be validated and defaulted.
	_ apis.Validatable   = (*SimpleDeployment)(nil)
	_ apis.Defaultable   = (*SimpleDeployment)(nil)
	_ kmeta.OwnerRefable = (*SimpleDeployment)(nil)
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*SimpleDeployment)(nil)
)

// SimpleDeploymentSpec holds the desired state of the SimpleDeployment (from the client).
type SimpleDeploymentSpec struct {
}

const (
	TaskrunSucceeded               apis.ConditionType = "TaskRunSucceeded"
	SimpleDeploymentConditionReady                    = apis.ConditionReady

	PHASEPENDING   string = "PENDING"
	PHASERUNNING   string = "RUNNING"
	PHASECOMPLETED string = "COMPLETED"
)

// SimpleDeploymentStatus communicates the observed state of the SimpleDeployment (from the controller).
type SimpleDeploymentStatus struct {
	duckv1.Status `json:",inline"`
	TaskRunName   string `json:"taskRun,inline"`
	Phase         string `json:"phase,inline"`
}

// SimpleDeploymentList is a list of AddressableService resources
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SimpleDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SimpleDeployment `json:"items"`
}

// GetStatus retrieves the status of the resource. Implements the KRShaped interface.
func (d *SimpleDeployment) GetStatus() *duckv1.Status {
	return &d.Status.Status
}

func (d *SimpleDeployment) GetNamespace() string {
	return d.Namespace
}

func (d *SimpleDeployment) SetNamespace(namespace string) {
	d.Namespace = namespace
}

func (d *SimpleDeployment) GetName() string {
	return d.Name
}
