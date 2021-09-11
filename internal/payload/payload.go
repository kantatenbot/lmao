package payload

import (
	"encoding/json"

	"github.com/google/shlex"
)

type Payload struct {
	RunId  string   `json:"run_id"`
	Script string   `json:"script"`
	Argv   []string `json:"argv"`
}

func NewPayload(runId string, script string, argvString string) *Payload {
	argv, _ := shlex.Split(argvString)
	p := &Payload{
		Script: script,
		Argv:   argv,
		RunId:  runId,
	}
	return p
}

func (p Payload) String() string {
	marshalled, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(marshalled)
}

func UnmarshalPayload(b []byte) (*Payload, error) {
	payload := &Payload{}
	err := json.Unmarshal(b, payload)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

type Output struct {
	// Argv contains the args passed to the command
	Argv []string `json:"argv"`
	// RunId contains a *client-assigned* identifier for the runs (as in, it's an ID set by mass-exec run to group executions)
	RunId string `json:"run_id"`
	// Bucket contains the name of the bucket containing output
	Bucket string `json:"bucket"`
	// StdoutKey contains the the key of the object containing stderr
	ObjectKey string `json:"key"`
	// Status contains the exit status of the command
	Status int `json:"status"`
	// Errors contains errors from the runtime
	Errors []string `json:"errors"`
}

func NewOutputFromPayload(payload *Payload, bucket string) *Output {
	return &Output{
		Argv:   payload.Argv,
		RunId:  payload.RunId,
		Errors: []string{},
		Bucket: bucket,
	}
}

func (o *Output) AddError(e error) {
	o.Errors = append(o.Errors, e.Error())
}

func (o *Output) SetObjectKey(k string) {
	o.ObjectKey = k
}

func (o *Output) String() string {
	marshalled, _ := json.Marshal(o)
	return string(marshalled)
}

func UnmarshalOutput(b []byte) (*Output, error) {
	output := &Output{}
	err := json.Unmarshal(b, output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// TODO refactor
type ProcessOutput struct {
	Argv   []string `json:"argv"`
	Stdout []byte   `json:"stdout"`
	Stderr []byte   `json:"stderr"`
	Status int      `json:"status"`
}
