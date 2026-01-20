package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/wcx0206/hermes/internal/config"
)

type overview struct {
	Logging  config.Logging  `yaml:"logging"`
	Defaults config.Defaults `yaml:"defaults"`
}

type configOpts struct {
	configPath string
}

func NewConfigCmd() *cobra.Command {
	opts := &configOpts{}
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect or edit hermes configuration",
	}
	cmd.PersistentFlags().StringVar(&opts.configPath, "config", "config.yaml", "config file path")

	cmd.AddCommand(
		newConfigShowCmd(opts),
		newConfigEditCmd(opts),
	)
	return cmd
}

func newConfigShowCmd(opts *configOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print current config.yaml",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.LoadConfig(opts.configPath)
			if err != nil {
				return err
			}
			ov := overview{
				Logging:  cfg.Logging,
				Defaults: cfg.Defaults,
			}
			b, err := yaml.Marshal(&ov)
			if err != nil {
				return err
			}
			cmd.Print(string(b))
			return nil
		},
	}
}

func newConfigEditCmd(opts *configOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Interactively update logging/defaults",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.LoadConfig(opts.configPath)
			if err != nil {
				return err
			}

			reader := bufio.NewReader(os.Stdin)
			cfg.Logging.Path = promptDefault(reader, "Logging path", cfg.Logging.Path)
			cfg.Logging.Debug = strings.ToLower(promptDefault(reader, "Logging debug (true/false)", fmt.Sprintf("%t", cfg.Logging.Debug))) == "true"

			cfg.Defaults.RcloneRemote = promptDefault(reader, "Defaults rclone_remote", cfg.Defaults.RcloneRemote)
			cfg.Defaults.Bucket = promptDefault(reader, "Defaults bucket", cfg.Defaults.Bucket)
			cfg.Defaults.Cron = promptDefault(reader, "Defaults cron", cfg.Defaults.Cron)

			return config.SaveConfig(opts.configPath, cfg)
		},
	}
}
