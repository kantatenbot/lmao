package command

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	inputFlagName  = "input"
	scriptFlagName = "file"
)

func errorln(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

func fatal(a ...interface{}) {
	errorln(a...)
	os.Exit(1)
}

func useScriptFileFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(scriptFlagName, "f", "", "file containing a script")
}

func getScript(cmd *cobra.Command, args []string) (string, error) {
	// explicit inline script overrides the file flag
	if len(args) > 0 {
		return args[0], nil
	}

	// look for a file
	scriptFile, _ := cmd.Flags().GetString(scriptFlagName)
	if scriptFile != "" {
		contents, err := os.ReadFile(scriptFile)
		if err != nil {
			return "", fmt.Errorf("couldn't read script file, %s", err.Error())
		}
		return string(contents), nil
	} else {
		return "", fmt.Errorf("no file or script provided")
	}
}

func useInputFileFlag(cmd *cobra.Command) {
	cmd.Flags().String(inputFlagName, "-", "input file. use - for stdin. use empty string for no inputs")
}

// getInputCh returns a chan string of the inputs, one input per message
func getInputCh(cmd *cobra.Command) (<-chan string, error) {
	inputFilename, _ := cmd.Flags().GetString(inputFlagName)
	switch inputFilename {
	case "":
		// run without args. we fake this using a chan that contains a single empty string.
		ch := make(chan string)
		go func() {
			ch <- ""
			close(ch)
		}()
		return ch, nil
	case "-":
		return makeScannerChan(os.Stdin), nil
	default:
		f, err := os.Open(inputFilename)
		if err != nil {
			return nil, fmt.Errorf("couldn't read input file, %s", err.Error())
		}
		return makeScannerChan(f), nil
	}
}

// makeScannerChan returns a channel yielding lines from an open file handle.
func makeScannerChan(f *os.File) chan string {
	ch := make(chan string)
	scanner := bufio.NewScanner(f)
	go func() {
		defer f.Close() // odd place to put this i think lol
		for scanner.Scan() {
			ch <- scanner.Text()
		}
		close(ch)
	}()

	return ch
}
