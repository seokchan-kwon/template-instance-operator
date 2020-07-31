/*
 * Kubernetes
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: v1.10.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package client

// NamespaceSpec describes the attributes on a Namespace.
type V1NamespaceSpec struct {

	// Finalizers is an opaque list of values that must be empty to permanently remove object from storage. More info: https://kubernetes.io/docs/tasks/administer-cluster/namespaces/
	Finalizers []string `json:"finalizers,omitempty"`
}
