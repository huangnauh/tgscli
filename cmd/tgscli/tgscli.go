package main

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/huangnauh/tgscli/pkg/version"
	"github.com/spf13/cobra"
)

var (
	outWriter io.Writer = os.Stdout
	errWriter io.Writer = os.Stderr
)

var rootCmd = &cobra.Command{
	Use:               version.APP,
	DisableAutoGenTag: true,
	Version: fmt.Sprintf("%s (%s), runtime:%s/%s %s", version.GitDescribe,
		version.GitCommit, runtime.GOOS, runtime.GOARCH, runtime.Version()),
	Short: "Storage Command Line utility for telegram files management",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		outWriter = cmd.OutOrStdout()
		errWriter = cmd.ErrOrStderr()
	},
}

func init() {
	cobra.MousetrapHelpText = `This is a command line tool.

You need to open pwsh.exe or cmd.exe and run it from there.
`
	cobra.OnInitialize(onInit)
}

func errorExitf(format string, a ...interface{}) {
	fmt.Fprintf(errWriter, format+"\n", a...)
	os.Exit(1)
}

func onInit() {

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
