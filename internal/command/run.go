package command

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"

	pl "github.com/kantatenbot/mass-exec/internal/payload"
	"github.com/spf13/cobra"
)

var (
	functionName     string
	functionNameFlag = "function-name"
)

var runCommand = &cobra.Command{
	Use:   "run [script]",
	Short: "Run scripts in AWS Lambda",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		runId := fmt.Sprintf("%d", time.Now().Unix())
		errorln("run id", runId)

		functionName, _ := cmd.Flags().GetString(functionNameFlag)
		script, err := getScript(cmd, args)
		if err != nil {
			fatal(err.Error())
		}

		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			fatal("couldn't load aws config,", err)
		}
		client := lambda.NewFromConfig(cfg)

		// since we use synchronous calls, limit how many network connections we make
		sem := make(chan bool, 50)
		wg := sync.WaitGroup{}

		input, err := getInputCh(cmd)
		if err != nil {
			fatal("Couldn't get inputs,", err.Error())
		}
		for line := range input {
			sem <- true
			wg.Add(1)
			go func(argvString string) {
				defer func() {
					wg.Done()
					<-sem
				}()

				// invoke the lambda
				p := pl.NewPayload(runId, script, argvString)
				resp, err := client.Invoke(cmd.Context(), &lambda.InvokeInput{
					FunctionName:   &functionName,
					InvocationType: types.InvocationTypeRequestResponse,
					Payload:        []byte(p.String()),
				})
				if err != nil {
					errorln("error invoking function,", err.Error())
				} else {
					// TODO: probably you want to write to a channel instead, this is a race condition
					os.Stdout.Write(append(resp.Payload, []byte("\n")...))
				}
			}(line)
		}

		wg.Wait()
	},
}

func init() {
	useScriptFileFlag(runCommand)
	useInputFileFlag(runCommand)
	runCommand.Flags().StringVar(&functionName, functionNameFlag, "mass-exec", "name of lambda")

	rootCommand.AddCommand(runCommand)
}
