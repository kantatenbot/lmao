package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/gosimple/slug"

	pl "github.com/kantatenbot/mass-exec/internal/payload"
)

func main() {
	lambda.Start(handler)
}

var (
	Bucket string
	Prefix string
)

func init() {
	b, ok := os.LookupEnv("MASS_EXEC_BUCKET_NAME")
	if !ok {
		log.Fatalf("Missing MASS_EXEC_BUCKET_NAME in env")
	}
	Bucket = b

	p, ok := os.LookupEnv("MASS_EXEC_OBJECT_PREFIX")
	if !ok {
		log.Fatalf("Missing MASS_EXEC_OBJECT_PREFIX in env")
	}
	Prefix = p
}

func handler(ctx context.Context, payload *pl.Payload) (*pl.Output, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Unable to load AWS credentials, %s", err.Error())
	}
	s3client := s3.NewFromConfig(cfg)
	processOuput := process(ctx, payload)
	// TODO(kantatenbot) fix names and refactor when sober
	return saveOutput(ctx, s3client, payload, processOuput)
}

func process(ctx context.Context, payload *pl.Payload) *pl.ProcessOutput {
	log.Printf("processing %s with args %s\n", payload.Script, payload.Argv)

	scriptfile, err := ioutil.TempFile("/tmp", "script-*")
	if err != nil {
		return &pl.ProcessOutput{
			Stderr: []byte(fmt.Sprintf("error writing script file, %s", err.Error())),
		}
	}
	defer os.Remove(scriptfile.Name())

	// write the script to a file and make it executable
	os.WriteFile(scriptfile.Name(), []byte(payload.Script), 0755)
	os.Chmod(scriptfile.Name(), 0755)
	scriptfile.Close() // you can't exec an open file

	// check for a shebang. if there is one, the script is the
	// executable. if not, treat it as a bash script
	var executable string
	var argv []string
	if strings.HasPrefix(payload.Script, "#!") {
		executable = scriptfile.Name()
	} else {
		executable = "bash"
		argv = append(argv, scriptfile.Name())
	}
	if err != nil {
		return &pl.ProcessOutput{
			Stderr: []byte(fmt.Sprintf("error getting script arguments, %s", err.Error())),
		}
	}
	argv = append(argv, payload.Argv...)

	// exec and collect the output
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(executable, argv...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err = cmd.Run()
	status := 0
	if err != nil {
		// errors at this point are probably important to the user, so we don't early return
		log.Printf("command err, %s", err.Error())
		if exitError, ok := err.(*exec.ExitError); ok {
			status = exitError.ExitCode()
		}
	}

	return &pl.ProcessOutput{
		Argv:   payload.Argv,
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
		Status: status,
	}
}

// TODO rename
func saveOutput(ctx context.Context, client *s3.Client, p *pl.Payload, po *pl.ProcessOutput) (*pl.Output, error) {
	output := pl.NewOutputFromPayload(p, Bucket)
	uploader := manager.NewUploader(client)
	m, err := json.Marshal(*po)
	if err != nil {
		return nil, err
	}

	s := slug.Make(strings.Join(p.Argv, " "))
	key := path.Join(Prefix, p.RunId, s)
	err = saveProcessOutput(ctx, uploader, p, bytes.NewBuffer(m), key)
	if err != nil {
		output.AddError(err)
		return output, err
	}
	output.SetObjectKey(key)

	return output, nil
}

// TODO rename
func saveProcessOutput(ctx context.Context, uploader *manager.Uploader, p *pl.Payload, body io.Reader, key string) error {
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		return fmt.Errorf("error saving stdout, %s", err.Error())
	}
	return nil
}
