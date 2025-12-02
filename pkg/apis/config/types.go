// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PostgresConfigSpec defines the desired state of [PostgresConfig]
type PostgresConfigSpec struct {
	// VolumeSize specifies the size of the persistent volume for the
	// cluster nodes.
	VolumeSize resource.Quantity

	// Replicas specifies the number of cluster instances.
	Replicas int32

	// Users specifies the database users and their roles.
	Users map[string][]string

	// Databases specifies the database names and their owners.
	Databases map[string]string

	// PostgresVersion specifies the PostgreSQL version of the cluster.
	PostgresVersion string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PostgresConfig is the schema for the PostgresConfig API
type PostgresConfig struct {
	metav1.TypeMeta

	// Spec provides the extension configuration spec.
	Spec PostgresConfigSpec
}
