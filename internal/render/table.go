package render

import (
	"aks-coach/internal/compute"
	"aks-coach/internal/kube"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// PrintTable renders a table of deployment information using the provided namespace scope and list of rows.
// scope defines the namespace or all namespaces for the data.
// rows is a slice of compute.Row, each representing data for one deployment.
// It outputs the table to standard output.
func PrintTable(scope kube.Scope, rows []compute.Row) {
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.SetStyle(table.StyleRounded)
	tw.Style().Options.SeparateRows = false
	tw.Style().Format.Header = text.FormatUpper

	tw.SetTitle(fmt.Sprintf("Scope %s", scope.Label()))
	tw.Style().Title.Align = text.AlignLeft

	tw.AppendHeader(table.Row{
		"Namespace", "Name", "Replicas", "CPU Req(m)", "CPU Limit(m)", "Mem Req(Mi)", "Mem Limit(Mi)", "HPA Min", "HPA Max", "HPA Target",
	})

	for _, r := range rows {
		tw.AppendRow(table.Row{
			r.Namespace,
			text.Trim(r.Name, 32),
			r.Replicas,
			int64(r.CPUReqMilli),
			int64(r.CPULimitMilli),
			int64(r.MemReqMi),
			int64(r.MemLimitMi),
			r.HPAMin, r.HPAMax, text.Trim(r.HPATarget, 32),
		})
	}

	// Right-align numbers
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Replicas", Align: text.AlignRight},
		{Name: "CPU Req(m)", Align: text.AlignRight},
		{Name: "CPU Limit(m)", Align: text.AlignRight},
		{Name: "Mem Req(Mi)", Align: text.AlignRight},
		{Name: "Mem Limit(Mi)", Align: text.AlignRight},
	})

	tw.Render()
}
