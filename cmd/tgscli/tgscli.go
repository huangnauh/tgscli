package main

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/huangnauh/tgscli/pkg/version"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

var (
	outWriter io.Writer = os.Stdout
	errWriter io.Writer = os.Stderr
	inReader  io.Reader = os.Stdin

	colorableOut io.Writer = colorable.NewColorableStdout()
)

var rootCmd = &cobra.Command{
	Use: version.APP,
	Version: fmt.Sprintf("%s (%s), runtime:%s/%s %s", version.GitDescribe,
		version.GitCommit, runtime.GOOS, runtime.GOARCH, runtime.Version()),
	Short: "Storage Command Line utility for cluster management",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		outWriter = cmd.OutOrStdout()
		errWriter = cmd.ErrOrStderr()
		inReader = cmd.InOrStdin()

		if outWriter != os.Stdout {
			colorableOut = outWriter
		}
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

func errorExit(err error) {
	fmt.Fprintf(errWriter, err.Error())
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
