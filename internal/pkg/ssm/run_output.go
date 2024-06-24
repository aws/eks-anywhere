package ssm

import "github.com/aws/aws-sdk-go/service/ssm"

type RunOutput struct {
	commandOut     *ssm.GetCommandInvocationOutput
	CommandId      string
	StdOut, StdErr []byte
}

func buildRunOutput(commandOut *ssm.GetCommandInvocationOutput) *RunOutput {
	return &RunOutput{
		commandOut: commandOut,
		CommandId:  *commandOut.CommandId,
		StdOut:     []byte(*commandOut.StandardOutputContent),
		StdErr:     []byte(*commandOut.StandardErrorContent),
	}
}

func (r *RunOutput) Successful() bool {
	return *r.commandOut.Status == ssm.CommandInvocationStatusSuccess
}

// StatusDetails returns the status details of the ssm command.
func (r *RunOutput) StatusDetails() string {
	// handle nil pointer
	if r.commandOut == nil {
		return "NoCommandOutput"
	}
	return *r.commandOut.StatusDetails
}
