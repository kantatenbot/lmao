package command

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gosimple/slug"

	pl "github.com/kantatenbot/mass-exec/internal/payload"
	"github.com/spf13/cobra"
)

var (
	outputDirectory string
	decode          bool
)

var receiveCommand = &cobra.Command{
	Use:   "receive [script]",
	Short: "Receive output from run",
	Long:  "Receive output from run. Specifiy --decode to print the output to stdout. --output-dir and --decode can be used together to save output and print at the same time",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		input, err := getInputCh(cmd)
		if err != nil {
			fatal("Couldn't get inputs,", err.Error())
		}

		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			fatal("Unable to load AWS credentials,", err.Error())
		}
		s3client := s3.NewFromConfig(cfg)

		count := 0
		// TODO refactor
		for line := range input {
			func() {
				defer func() {
					count += 1
					if !decode {
						fmt.Fprintln(os.Stderr, "\rReceived", count)
					}
				}()
				output := &pl.Output{}
				err = json.Unmarshal([]byte(line), output)
				if err != nil {
					errorln(fmt.Printf("error decoding line,%s, %s", line, err.Error()))
				}
				result, err := s3client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: &output.Bucket,
					Key:    &output.ObjectKey,
				})
				if err != nil {
					errorln(fmt.Printf("error downloading line,%s", err.Error()))
				}

				if outputDirectory != "" {
					func() {
						fileName := slug.Make(strings.Join(output.Argv, " "))
						outPath := path.Join(outputDirectory, fileName)
						file, err := os.Create(outPath)
						if err != nil {
							fatal(fmt.Sprintf("unable to create output file %s, %s", fileName, err.Error()))
						}
						defer func() {
							if err := file.Close(); err != nil {
								panic(err)
							}
						}()

						_, err = io.Copy(file, result.Body)
						if err != nil {
							fatal(fmt.Sprintf("unable to write output file %s, %s", fileName, err.Error()))
						}
					}()
				}
				if decode {
					po := &pl.ProcessOutput{}
					json.NewDecoder(result.Body).Decode(po)
					os.Stdout.Write(po.Stdout)
					os.Stderr.Write(po.Stderr)
				} else {
					io.Copy(os.Stdout, result.Body)

				}
			}()
		}
	},
}

func init() {
	useInputFileFlag(receiveCommand)
	receiveCommand.Flags().StringVarP(&outputDirectory, "output-directory", "o", "", "save files to directory")
	receiveCommand.Flags().BoolVarP(&decode, "decode", "D", false, "decode output (stdout goes to &1, stderr to &2")

	rootCommand.AddCommand(receiveCommand)
}
