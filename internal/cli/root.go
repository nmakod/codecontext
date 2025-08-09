package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	// Version information
	appVersion = "2.4.0"
	buildDate  = "unknown"
	gitCommit  = "unknown"

	rootCmd = &cobra.Command{
		Use:   "codecontext",
		Short: "CodeContext - Intelligent context maps for AI-powered development",
		Long: `CodeContext is an automated repository mapping system that generates
intelligent context maps for AI-powered development tools, with a focus on
token optimization and incremental updates.`,
		Version: appVersion,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

// SetVersion sets the version information from build time
func SetVersion(version, date, commit string) {
	if version != "" {
		appVersion = version
		rootCmd.Version = version
	}
	if date != "" {
		buildDate = date
	}
	if commit != "" {
		gitCommit = commit
	}

	// Update version template to include build info
	rootCmd.SetVersionTemplate(fmt.Sprintf(`{{with .Name}}{{printf "%%s " .}}{{end}}{{printf "version %%s" .Version}}
Build Date: %s
Git Commit: %s
`, buildDate, gitCommit))
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .codecontext/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringP("output", "o", "CLAUDE.md", "output file")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".codecontext")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	// Skip reading config file for init command to avoid hanging in large repos
	if len(os.Args) > 1 && os.Args[1] == "init" {
		return
	}

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
		}
	}
}
