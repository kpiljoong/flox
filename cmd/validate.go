package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kpiljoong/flox/internal/config"
)

var validateCmd = &cobra.Command{
	Use:   "validate [config.yaml]",
	Short: "Validate a Flox pipeline config",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfgFile := args[0]
		fmt.Printf("Validating pipeline config: %s\n", cfgFile)

		v := viper.New()
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			fmt.Printf("Error reading config file: %s\n", err)
			os.Exit(1)
		}

		var pipeline config.PipelineConfig
		if err := v.Unmarshal(&pipeline); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse config: %v\n", err)
			os.Exit(1)
		}

		// Basic validation
		if pipeline.Input.Type == "" {
			fmt.Fprintln(os.Stderr, "'input.type' is required")
			os.Exit(1)
		}
		if pipeline.Output.Type == "" {
			fmt.Fprintln(os.Stderr, "'output.type' is required")
			os.Exit(1)
		}
		fmt.Println("Config is valid")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
