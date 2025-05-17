package main

import (
	"fmt"
	"os"
	"path/filepath"
	"projectson/collector"
	"projectson/config"
	"projectson/utils"

	"github.com/spf13/cobra"
)

var (
	cfgFile         string
	projectRoot     string
	outputFile      string
	formats         []string
	includes        []string
	excludePatterns []string
	forceApply      bool
)

var rootCmd = &cobra.Command{
	Use:   "projectson-cli",
	Short: "ProjectSon CLI aggregates project files into a structured JSON output.",
	Long: `ProjectSon CLI is a command-line tool to scan project directories,
collect specified file types.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var initConfigCmd = &cobra.Command{
	Use:   "init",
	Short: "create a default configuration file (projectson_config.yaml)",
	Long: `creates a projectson_config.yaml file with default values in the current
directory or at the specified --config path. You can pre-fill values
using flags like --root, --output, and --formats.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.NewDefaultConfig()

		if projectRoot != "" {
			absPath, err := filepath.Abs(projectRoot)
			if err != nil {
				return fmt.Errorf("invalid root path: %w", err)
			}
			cfg.Root = absPath
		}
		if outputFile != "" {
			cfg.Output = outputFile
		}
		if len(formats) > 0 {
			cfg.Formats = formats
		}
		if len(includes) > 0 {
			cfg.Include = includes
		}

		targetCfgFile := cfgFile
		if targetCfgFile == "" {
			targetCfgFile = "projectson_config.yaml"
		}

		if _, err := os.Stat(targetCfgFile); err == nil && !forceApply {
			return fmt.Errorf("config file '%s' already exists. Use --force to overwrite", targetCfgFile)
		}

		if err := cfg.SaveConfig(targetCfgFile); err != nil {
			return fmt.Errorf("failed to save default config: %w", err)
		}
		fmt.Printf("Default configuration saved to %s\n", targetCfgFile)
		fmt.Println("Please review and edit this file, especially the 'root' and 'formats' fields.")
		return nil
	},
}

var validateConfigCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigWithOverrides(cmd)
		if err != nil {
			return err
		}
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}
		fmt.Println("Configuration is valid.")
		return nil
	},
}

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview files that will be collected",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigWithOverrides(cmd)
		if err != nil {
			return err
		}
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("configuration error: %w. Run 'validate' command for details", err)
		}

		fc, err := collector.NewFileCollector(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize collector: %w", err)
		}

		fmt.Println("Scanning for files to preview...")
		entries, err := fc.PreviewFiles()
		if err != nil {
			return fmt.Errorf("error during file preview: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("No files found matching the criteria.")
			return nil
		}

		fmt.Printf("Found %d files:\n", len(entries))
		fmt.Println("--------------------------------------------------")
		for _, entry := range entries {
			fmt.Printf("- Path: %s\n", entry.Path)
			fmt.Printf("  Original Path: %s\n", entry.OriginalPath)
			fmt.Printf("  Format: %s, Mode: %s, Size: %s\n", entry.Format, entry.Mode, utils.FormatSize(entry.Size))
		}
		fmt.Println("--------------------------------------------------")
		return nil
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run collection process",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigWithOverrides(cmd)
		if err != nil {
			return err
		}
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("configuration error: %w. Run 'validate' command for details", err)
		}

		fc, err := collector.NewFileCollector(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize collector: %w", err)
		}

		fmt.Println("Starting file collection process")

		progressCallback := func(current, total int) {
			fmt.Printf("\rProcessing: %d / %d files", current, total)
			if current == total && total > 0 {
				fmt.Println()
			}
		}
		if cfg.Output == "" {
			cfg.Output = "output.json"
		}

		count, sizeStr, err := fc.Run(progressCallback)
		if err != nil {
			return fmt.Errorf("error during file collection: %w", err)
		}

		fmt.Printf("\ncollection completed\n")
		fmt.Printf("--------------------------------------------------\n")
		fmt.Printf("files processed: %d\n", count)
		fmt.Printf("output size: %s\n", sizeStr)
		fmt.Printf("output written to: %s\n", cfg.Output)
		fmt.Println("--------------------------------------------------")
		return nil
	},
}

func loadConfigWithOverrides(cmd *cobra.Command) (*config.Config, error) {
	var cfg *config.Config
	var err error

	configPathToLoad := cfgFile
	if configPathToLoad != "" {
		if _, errStat := os.Stat(configPathToLoad); os.IsNotExist(errStat) {

			if cmd.Name() != "init" {
				return nil, fmt.Errorf("config file not found: %s", configPathToLoad)
			}
			cfg = config.NewDefaultConfig()
		} else if errStat == nil {
			cfg, err = config.LoadConfig(configPathToLoad)
			if err != nil {
				return nil, fmt.Errorf("error loading config file %s: %w", configPathToLoad, err)
			}
		} else {
			return nil, fmt.Errorf("error stating config file %s: %w", configPathToLoad, err)
		}
	} else {
		if _, errStat := os.Stat("projectson_config.yaml"); errStat == nil {
			cfg, err = config.LoadConfig("projectson_config.yaml")
			if err != nil {
				return nil, fmt.Errorf("error loading default config file projectson_config.yaml: %w", err)
			}
		} else {
			cfg = config.NewDefaultConfig()
		}
	}

	if cmd.Flags().Changed("root") && projectRoot != "" {
		absPath, err := filepath.Abs(projectRoot)
		if err != nil {
			return nil, fmt.Errorf("invalid root path from flag: %w", err)
		}
		cfg.Root = absPath
	}
	if cmd.Flags().Changed("output") && outputFile != "" {
		cfg.Output = outputFile
	}
	if cmd.Flags().Changed("formats") && len(formats) > 0 {
		cfg.Formats = formats
	}
	if cmd.Flags().Changed("include") && len(includes) > 0 {
		cfg.Include = includes
	}
	if cmd.Flags().Changed("exclude") && len(excludePatterns) > 0 {
		cfg.ExcludePatterns = excludePatterns
	}

	if cfg.Output != "" {
		outputDir := filepath.Dir(cfg.Output)
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			if mkErr := os.MkdirAll(outputDir, 0755); mkErr != nil {
				fmt.Printf("warning: could not create output directory %s: %v\n", outputDir, mkErr)
			}
		}
	} else if cmd.Name() == "run" {
		cfg.Output = "output.json"
		fmt.Println("Output path not specified, defaulting to 'output.json'")
	}

	return cfg, nil
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file (default is projectson_config.yaml in current dir)")

	overrideFlags := []*cobra.Command{runCmd, previewCmd, validateConfigCmd, initConfigCmd}
	for _, cmd := range overrideFlags {
		cmd.Flags().StringVarP(&projectRoot, "root", "r", "", "Project root directory (overrides config)")
		cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output JSON file path (overrides config)")
		cmd.Flags().StringSliceVarP(&formats, "formats", "f", []string{}, "File formats/extensions, comma-separated (e.g., go,vue,ts) (overrides config)")

		if cmd == initConfigCmd {
			cmd.Flags().StringSliceVar(&includes, "include", []string{}, "Paths to include relative to root, comma-separated (e.g. src,docs/api.md) (for init)")
		}
		if cmd != initConfigCmd {
			cmd.Flags().StringSliceVarP(&excludePatterns, "exclude", "e", []string{}, "Exclude patterns, comma-separated (e.g., node_modules,*.log) (overrides config)")
		}
	}

	initConfigCmd.Flags().BoolVar(&forceApply, "force", false, "force overwrite if config file already exists")

	rootCmd.AddCommand(initConfigCmd)
	rootCmd.AddCommand(validateConfigCmd)
	rootCmd.AddCommand(previewCmd)
	rootCmd.AddCommand(runCmd)
}

func main() {
	Execute()
}
