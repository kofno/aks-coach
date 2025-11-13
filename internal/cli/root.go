package cli

import (
	"aks-coach/internal/compute"
	"aks-coach/internal/kube"
	"aks-coach/internal/render"
	"aks-coach/internal/version"
	"context"
	"time"

	"github.com/spf13/cobra"
)

var (
	flagNamespace     string
	flagAllNamespaces bool
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
			render.PrintTable(scope, rows)

			return nil
		}}

	cmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "", "namespace scope for this request")
	cmd.Flags().BoolVarP(&flagAllNamespaces, "all-namespaces", "A", false, "if present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")

	cmd.CompletionOptions.DisableDefaultCmd = false

	return cmd
}
