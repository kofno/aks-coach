package kube

import "fmt"

type Scope struct {
	AllNamespaces bool
	Namespace     string
}

func (s Scope) String() string {
	if s.AllNamespaces {
		return "all namespaces"
	}
	return s.Namespace
}

func (s Scope) Label() string {
	if s.AllNamespaces {
		return "all namespaces"
	}
	return fmt.Sprintf("namespace %q", s.Namespace)
}

func (s Scope) NS() string {
	if s.AllNamespaces {
		return ""
	}
	return s.Namespace
}
