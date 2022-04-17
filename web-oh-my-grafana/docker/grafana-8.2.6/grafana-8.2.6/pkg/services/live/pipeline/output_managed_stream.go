package pipeline

import (
	"context"

	"github.com/grafana/grafana/pkg/services/live/managedstream"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type ManagedStreamOutput struct {
	managedStream *managedstream.Runner
}

func NewManagedStreamOutput(managedStream *managedstream.Runner) *ManagedStreamOutput {
	return &ManagedStreamOutput{managedStream: managedStream}
}

func (l *ManagedStreamOutput) Output(_ context.Context, vars OutputVars, frame *data.Frame) ([]*ChannelFrame, error) {
	stream, err := l.managedStream.GetOrCreateStream(vars.OrgID, vars.Scope, vars.Namespace)
	if err != nil {
		logger.Error("Error getting stream", "error", err)
		return nil, err
	}
	return nil, stream.Push(vars.Path, frame)
}
