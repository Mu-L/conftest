package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/open-policy-agent/conftest/downloader"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const pullDesc = `
This command downloads individual policies from a remote location.

Several locations are supported by the pull command. Under the hood
conftest leverages go-getter (https://github.com/hashicorp/go-getter).
The following protocols are supported for downloading policies:

	- OCI Registries
	- Local Files
	- Git
	- HTTP/HTTPS
	- Mercurial
	- Amazon S3
	- Google Cloud Storage

The location of the policies is specified by passing an URL, e.g.:

	$ conftest pull http://<my-policy-url>

Based on the protocol a different mechanism will be used to download the policy.
The pull command will also try to infer the protocol based on the URL if the
URL does not contain a protocol. For example, the OCI mechanism will be used if
an azure registry URL is passed, e.g.

	$ conftest pull instrumenta.azurecr.io/my-registry

The policy location defaults to the policy directory in the local folder.
The location can be overridden with the '--policy' flag, e.g.:

	$ conftest pull --policy <my-directory> <oci-url>

When using absolute paths, you can enable the '--absolute-paths' flag to preserve them:

	$ conftest pull --absolute-paths --policy /absolute/path/to/policies <oci-url>
`

// NewPullCommand creates a new pull command to allow users
// to download individual policies.
func NewPullCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "pull <repository>",
		Short: "Download individual policies",
		Long:  pullDesc,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := viper.BindPFlag("policy", cmd.Flags().Lookup("policy")); err != nil {
				return fmt.Errorf("bind flag: %w", err)
			}
			if err := viper.BindPFlag("tls", cmd.Flags().Lookup("tls")); err != nil {
				return fmt.Errorf("bind flag: %w", err)
			}
			if err := viper.BindPFlag("absolute-paths", cmd.Flags().Lookup("absolute-paths")); err != nil {
				return fmt.Errorf("bind flag: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.Usage() //nolint
				return fmt.Errorf("missing required arguments")
			}

			policyPath := viper.GetString("policy")
			var policyDir string
			if viper.GetBool("absolute-paths") && filepath.IsAbs(policyPath) {
				policyDir = policyPath
			} else {
				policyDir = filepath.Join(".", policyPath)
			}

			if err := downloader.Download(ctx, policyDir, args); err != nil {
				return fmt.Errorf("download policies: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringP("policy", "p", "policy", "Path to download the policies to")
	cmd.Flags().BoolP("tls", "s", true, "Use TLS to access the registry")
	cmd.Flags().Bool("absolute-paths", false, "Preserve absolute paths in policy flag")

	return &cmd
}
