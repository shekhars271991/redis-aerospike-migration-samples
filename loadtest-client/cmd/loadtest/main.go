package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"loadtest-client/internal/config"
	"loadtest-client/internal/loadtest"
)

var (
	cfgFile     string
	baseURL     string
	duration    string
	qps         int
	workers     int
	outputFile  string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "loadtest",
	Short: "High-performance load testing client for quiz/game leaderboard API",
	Long: `A high-performance load testing client designed to generate massive loads (1M+ QPS)
for the quiz/game leaderboard API. Supports configurable throughput distribution,
real-time metrics, and detailed performance analysis.`,
	Run: runLoadTest,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./loadtest.yaml)")
	rootCmd.PersistentFlags().StringVar(&baseURL, "url", "http://localhost:8080", "target API base URL")
	rootCmd.PersistentFlags().StringVar(&duration, "duration", "60s", "test duration")
	rootCmd.PersistentFlags().IntVar(&qps, "qps", 1000, "target queries per second")
	rootCmd.PersistentFlags().IntVar(&workers, "workers", 100, "number of worker goroutines")
	rootCmd.PersistentFlags().StringVar(&outputFile, "output", "loadtest_results.json", "output file for results")

	// Bind flags to viper
	viper.BindPFlag("target.base_url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("load.duration", rootCmd.PersistentFlags().Lookup("duration"))
	viper.BindPFlag("load.target_qps", rootCmd.PersistentFlags().Lookup("qps"))
	viper.BindPFlag("load.workers", rootCmd.PersistentFlags().Lookup("workers"))
	viper.BindPFlag("reporting.output_file", rootCmd.PersistentFlags().Lookup("output"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("loadtest")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if cfgFile != "" {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
			os.Exit(1)
		}
		// If no config file specified and default not found, use default config
		log.Println("No config file found, using default configuration")
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
}

func runLoadTest(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Print configuration summary
	printConfigSummary(cfg)

	// Create load test runner
	runner, err := loadtest.NewRunner(cfg)
	if err != nil {
		log.Fatalf("Failed to create load test runner: %v", err)
	}
	defer runner.Close()

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Received interrupt signal, stopping load test...")
		cancel()
		runner.Stop()
	}()

	// Run the load test
	log.Println("Starting load test...")
	err = runner.Run(ctx)
	if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		log.Fatalf("Load test failed: %v", err)
	}

	// Export results if specified
	if cfg.Reporting.OutputFile != "" {
		if err := runner.ExportResults(cfg.Reporting.OutputFile); err != nil {
			log.Printf("Failed to export results: %v", err)
		}
	}

	log.Println("Load test completed successfully")
}

func loadConfig() (*config.Config, error) {
	// Start with default config
	cfg := config.GetDefaultConfig()

	// Unmarshal viper config into our struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func printConfigSummary(cfg *config.Config) {
	fmt.Printf("\n=== Load Test Configuration ===\n")
	fmt.Printf("Target URL: %s\n", cfg.Target.BaseURL)
	fmt.Printf("Duration: %s\n", cfg.Load.Duration)
	fmt.Printf("Target QPS: %d\n", cfg.Load.TargetQPS)
	fmt.Printf("Workers: %d\n", cfg.Load.Workers)
	fmt.Printf("Ramp-up: %s\n", cfg.Load.RampUpDuration)
	fmt.Printf("Max Connections: %d\n", cfg.HTTP.MaxConnsPerHost)
	
	fmt.Printf("\nAPI Distribution:\n")
	for _, api := range cfg.APIs {
		fmt.Printf("  %s (%s %s): %.1f%%\n", 
			api.Name, api.Method, api.Path, api.Weight*100)
	}
	
	fmt.Printf("\nReporting:\n")
	fmt.Printf("  Interval: %s\n", cfg.Reporting.Interval)
	fmt.Printf("  Output File: %s\n", cfg.Reporting.OutputFile)
	fmt.Printf("================================\n\n")
}
