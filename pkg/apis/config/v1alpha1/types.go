// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PostgresConfigSpec defines the desired state of [PostgresConfig]
type PostgresConfigSpec struct {
	// VolumeSize specifies the size of the persistent volume for the
	// cluster nodes.
	VolumeSize resource.Quantity `json:"volumeSize,omitZero"`

	// Replicas specifies the number of cluster instances.
	Replicas int `json:"replicas,omitzero"`

	// Users specifies the database users and their roles.
	Users map[string][]string `json:"users,omitzero"`

	// Databases specifies the database names and their owners.
	Databases map[string]string `json:"databases,omitzero"`

	// PostgresVersion specifies the PostgreSQL version of the cluster.
	PostgresVersion string `json:"postgresVersion,omitzero"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PostgresConfig is the schema for the PostgresConfig API
type PostgresConfig struct {
	metav1.TypeMeta `json:",inline"`

	// Spec provides the extension configuration spec.
	Spec PostgresConfigSpec `json:"spec,omitzero"`
}
