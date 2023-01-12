package namespaces

import v1 "k8s.io/api/core/v1"

type GetNamespaceMsg struct {
	Namespaces []v1.Namespace
}
