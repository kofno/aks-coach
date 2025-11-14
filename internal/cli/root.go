package cli

import (
	"aks-coach/internal/compute"
	"aks-coach/internal/kube"
	"aks-coach/internal/render"
	"aks-coach/internal/version"
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	flagNamespace     string
	flagAllNamespaces bool
	flagOutput        string
	flagSelector      string
)

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "aks-coach",
		Short:   "aks-coach is a Kubernetes cluster health and performance assessment tool",
		Version: version.String(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			client, err := kube.NewClient()
			if err != nil {
				return err
			}

			scope := kube.Scope{
				Namespace:     flagNamespace,
				AllNamespaces: flagAllNamespaces,
				Selector:      flagSelector,
			}
			deps, err := kube.ListDeployments(ctx, client, scope)
			if err != nil {
				return err
			}
			hpas, err := kube.ListHPAs(ctx, client, scope)
			if err != nil {
				return err
			}

			rows := compute.BuildRows(deps, hpas)

			switch flagOutput {
			case "json":
				err = render.PrintJSON(rows)
				if err != nil {
					return err
				}
			case "table":
				render.PrintTable(scope, rows)
			default:
				return fmt.Errorf("unknown --output=%s", flagOutput)
			}

			return nil
		}}

	cmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "", "namespace scope for this request")
	cmd.Flags().BoolVarP(&flagAllNamespaces, "all-namespaces", "A", false, "if present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().StringVarP(&flagOutput, "output", "o", "table", "output format (table|json)")
	cmd.Flags().StringVarP(&flagSelector, "selector", "l", "", "Label selector (e.g. app=srv)")

	cmd.CompletionOptions.DisableDefaultCmd = false

	return cmd
}
