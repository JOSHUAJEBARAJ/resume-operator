/*
Copyright 2025.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ResumeSpec defines the desired state of Resume.
type ResumeSpec struct {
	// Name of the individual
	Name string `json:"name"`

	// Contact information
	Contact ContactInfo `json:"contact"`

	// Work experience
	Experience []Experience `json:"experience"`

	// Skills
	Skills map[string][]string `json:"skills"`

	// Education details
	Education []Education `json:"education"`

	// Projects
	Projects []Project `json:"projects"`
}

// ContactInfo defines the contact details of the individual
type ContactInfo struct {
	Website  string `json:"website"`
	LinkedIn string `json:"linkedin"`
	GitHub   string `json:"github"`
	Email    string `json:"email"`
}

// Experience defines a single work experience entry
type Experience struct {
	Title            string   `json:"title"`
	Company          string   `json:"company"`
	Location         string   `json:"location"`
	Period           string   `json:"period"`
	Responsibilities []string `json:"responsibilities"`
}

// Skills defines the skills of the individua

// Education defines a single education entry
type Education struct {
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
	Location    string `json:"location"`
	Period      string `json:"period"`
}

// Project defines a single project entry
type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Link        string `json:"link"`
}

// ResumeStatus defines the observed state of Resume.
type ResumeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Resume is the Schema for the resumes API.
type Resume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResumeSpec   `json:"spec,omitempty"`
	Status ResumeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResumeList contains a list of Resume.
type ResumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Resume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Resume{}, &ResumeList{})
}
