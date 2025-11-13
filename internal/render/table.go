package render

import (
	"aks-coach/internal/compute"
	"aks-coach/internal/kube"
	"fmt"
	"os"
	"text/tabwriter"
)

// PrintTable displays a formatted table of deployment details and resource metrics, grouped by the specified scope.
// scope defines the label of the namespace or clusters being analyzed.
// rows is a slice of Row objects containing the deployment and resource data to display.
func PrintTable(scope kube.Scope, rows []compute.Row) {
	fmt.Printf("Scope: %s\n\n", scope.Label())

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, err := fmt.Fprintln(w, "NAMESPACE\tNAME\tREPLICAS\tCPU_REQ(m)\tCPU_LIMIT(m)\tMEM_REQ(Mi)\tMEM_LIMIT(Mi)\tHPA_MIN\tHPA_MAX\tHPA_TARGET")
	if err != nil {
		return
	}
	for _, r := range rows {
		_, err := fmt.Fprintf(w, "%s\t%s\t%d\t%.0f\t%.0f\t%.0f\t%.0f\t%s\t%s\t%s\n",
			trunc(r.Namespace, 16),
			trunc(r.Name, 32),
			r.Replicas,
			r.CPUReqMilli, r.CPULimitMilli, r.MemReqMi, r.MemLimitMi,
			r.HPAMin, r.HPAMax, trunc(r.HPATarget, 20))
		if err != nil {
			return
		}
	}
	err = w.Flush()
	if err != nil {
		return
	}
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
