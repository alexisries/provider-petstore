/*
Copyright 2022 The Crossplane Authors.

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
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type PetCategory struct {
	// The id of the pet category
	Id int64 `json:"id"`

	// The name of the pet category
	Name string `json:"name"`
}

type PetTag struct {
	// The id of the pet tag
	Id int64 `json:"id"`

	// The name of the pet tag
	Name string `json:"name"`
}

// PetParameters define the desired state of an pet
type PetParameters struct {
	// The name of the Pet
	// +optional
	Name string `json:"name,omitempty"`

	// Category og the Pet
	// +optional
	Category *PetCategory `json:"category,omitempty"`

	// List of the pet tags
	// +optional
	Tags []PetTag `json:"tags,omitempty"`

	// List of pet photos url
	// +optional
	PhotoUrls []string `json:"photosUrls,omitempty"`
}

// PetObservation keeps the state of external resource
type PetObservation struct {
	// Id of the pet
	Id int64 `json:"id,omitempty"`

	// Status of the pet
	// +kubebuilder:validation:Enum=AVAILABLE;INPROGRESS;INACTIVE;PENDING;FAILED
	Status string `json:"status,omitempty"`
}

// A PetSpec defines the desired state of a Pet.
type PetSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       PetParameters `json:"forProvider"`
}

// A PetStatus represents the observed state of a Pet.
type PetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          PetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Pet is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,petstore}
type Pet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PetSpec   `json:"spec"`
	Status PetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PetList contains a list of Pet
type PetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pet `json:"items"`
}

// Pet type metadata.
var (
	PetKind             = reflect.TypeOf(Pet{}).Name()
	PetGroupKind        = schema.GroupKind{Group: Group, Kind: PetKind}.String()
	PetKindAPIVersion   = PetKind + "." + SchemeGroupVersion.String()
	PetGroupVersionKind = SchemeGroupVersion.WithKind(PetKind)
)

func init() {
	SchemeBuilder.Register(&Pet{}, &PetList{})
}
