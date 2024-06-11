/*
SKCF-API

The SKCF API provides endpoints to create and delete clusters within STACKIT portal projects and to trigger further cluster management tasks.

API version: 1alpha1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package skcf

type Kubeconfig struct {
	ExpirationTimestamp *string `json:"expirationTimestamp,omitempty"`
	Kubeconfig          *string `json:"kubeconfig,omitempty"`
}
