package kube

import "fmt"

type Scope struct {
	AllNamespaces bool
	Namespace     string
	Selector      string
}

func (s Scope) String() string {
	return s.Label()
}

func (s Scope) Label() string {
	base := "all namespaces"
	if !s.AllNamespaces {
		base = fmt.Sprintf("namespace %q", s.Namespace)
	}
	if s.Selector != "" {
		return fmt.Sprintf("%s (selector: %s)", base, s.Selector)
	}
	return base
}

func (s Scope) NS() string {
	if s.AllNamespaces {
		return ""
	}
	return s.Namespace
}
