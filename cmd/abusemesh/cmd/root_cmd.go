package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

//The value of the 'output' flag
var outputFormatterFlag string

func init() {
	rootCmd.AddCommand(getCmd, watchCmd)

	getCmd.PersistentFlags().StringVarP(&outputFormatterFlag, "output", "o", "human", "Output format, one of: human, json")
}

//The root command
var rootCmd = &cobra.Command{
	Use:   "abusemesh",
	Short: "This is the abusemesh deamon CLI client",
	Long:  `Abusemesh is a application which can exchange internet abuse reports with other Abusemesh instances in a 'mesh' topology`,
}

//Get subcommand which has other children
// so we have semantic commands like: 'abusemesh get node' and 'abusemesh get events'
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get and show information",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		switch strings.ToLower(outputFormatterFlag) {
		case "human":
			ConsoleOutputFormatter = nil
		case "json":
			ConsoleOutputFormatter = JSONFormatter
		default:
			return errors.Errorf("'%s' is not a valid output format", outputFormatterFlag)
		}

		return nil
	},
}

//Watch subcommand which has other children
// so we have semantic commands like: 'abusemesh watch events' and 'abusemesh watch reports'
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for changes and show them",
	Long:  "Opens a stream to the node and will report changes until instructed to stop with Ctrl-D or Ctrl-C",
}

//Execute executes the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
