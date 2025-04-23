package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/flox/internal/config"
	"github.com/kpiljoong/flox/internal/filters"
	"github.com/kpiljoong/flox/internal/input"
	"github.com/kpiljoong/flox/internal/input/file"
	"github.com/kpiljoong/flox/internal/metrics"
	"github.com/kpiljoong/flox/internal/output"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "flox",
	Short: "Flox is a fast and programmable log/event processor",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup signal handler for graceful shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			fmt.Println("\n[signal] Shutting down gracefully...")
			cancel()
		}()

		// Load configuration
		cfg, rawCfg, err := config.Load(cfgFile)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Initialize metrics server
		metrics.InitMetricsServer()

		jsonFilters := setupFilters(cfg.Filters)

		// Setup output
		out, err := setupOutput(ctx, cfg.Output.Type, rawCfg)
		if err != nil {
			fmt.Printf("Error setting up output: %v\n", err)
			os.Exit(1)
		}

		// Input handler
		handler := buildHandler(ctx, jsonFilters, out)

		// Setup input
		switch cfg.Input.Type {
		case "http":
			input.StartHTTP(cfg.Input.Address, handler)
		case "file":
			file.StartFile(ctx, cfg.Input.Path, handler, cfg.Input.Namespace, cfg.Input.TrackOffset, cfg.Input.StartFrom)
		default:
			log.Fatalf("Unsupported input type: %s", cfg.Input.Type)
		}

		// Wait for shutdown signal
		<-ctx.Done()
		fmt.Println("[shutdown] Flox stopped.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func setupFilters(filterConfigs []config.FilterConfig) []*filters.JSONFilter {
	var jsonFilters []*filters.JSONFilter
	for _, f := range filterConfigs {
		if f.Type == "json" {
			jsonFilters = append(jsonFilters, filters.NewJSONFilter(f.DropFields, f.RenameFields, f.AddFields))
		}
	}
	return jsonFilters
}

func setupOutput(ctx context.Context, outputType string, rawCfg map[string]interface{}) (output.Output, error) {
	fmt.Printf("Creating output: %s...\n", outputType)
	outputConfigRaw, ok := rawCfg["output"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid output config")
	}
	return output.NewOutput(ctx, outputType, outputConfigRaw)
}

func buildHandler(ctx context.Context, filters []*filters.JSONFilter, out output.Output) func(map[string]interface{}) {
	return func(event map[string]interface{}) {
		// log.Printf("[Processing] Received event: %v\n", event)
		metrics.EventReceived.Inc()

		for _, f := range filters {
			event = f.Process(event)
			metrics.EventFiltered.Inc()
		}

		if err := out.Send(event); err != nil {
			if ctx.Err() != nil {
				return
			}
			fmt.Printf("Error sending event: %v\n", err)
			metrics.OutputFailure.Inc()
		} else {
			metrics.OutputSuccess.Inc()
		}
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "pipeline.yaml", "config file (default is pipeline.yaml)")
}
