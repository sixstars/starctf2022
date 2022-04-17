package pipeline

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/stretchr/testify/require"
)

type testRuleGetter struct {
	mu    sync.Mutex
	rules map[string]*LiveChannelRule
}

func (t *testRuleGetter) Get(orgID int64, channel string) (*LiveChannelRule, bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	rule, ok := t.rules[channel]
	return rule, ok, nil
}

func TestPipeline_New(t *testing.T) {
	p, err := New(&testRuleGetter{})
	require.NoError(t, err)
	require.NotNil(t, p)
}

func TestPipelineNoConverter(t *testing.T) {
	p, err := New(&testRuleGetter{
		rules: map[string]*LiveChannelRule{
			"test": {
				Converter: nil,
			},
		},
	})
	require.NoError(t, err)
	ok, err := p.ProcessInput(context.Background(), 1, "test", []byte(`{}`))
	require.NoError(t, err)
	require.False(t, ok)
}

type testConverter struct {
	channel string
	frame   *data.Frame
}

func (t *testConverter) Convert(_ context.Context, _ Vars, _ []byte) ([]*ChannelFrame, error) {
	return []*ChannelFrame{{Channel: t.channel, Frame: t.frame}}, nil
}

type testProcessor struct{}

func (t *testProcessor) Process(_ context.Context, _ ProcessorVars, frame *data.Frame) (*data.Frame, error) {
	return frame, nil
}

type testOutputter struct {
	err   error
	frame *data.Frame
}

func (t *testOutputter) Output(_ context.Context, _ OutputVars, frame *data.Frame) ([]*ChannelFrame, error) {
	if t.err != nil {
		return nil, t.err
	}
	t.frame = frame
	return nil, nil
}

func TestPipeline(t *testing.T) {
	outputter := &testOutputter{}
	p, err := New(&testRuleGetter{
		rules: map[string]*LiveChannelRule{
			"stream/test/xxx": {
				Converter: &testConverter{"", data.NewFrame("test")},
				Processor: &testProcessor{},
				Outputter: outputter,
			},
		},
	})
	require.NoError(t, err)
	ok, err := p.ProcessInput(context.Background(), 1, "stream/test/xxx", []byte(`{}`))
	require.NoError(t, err)
	require.True(t, ok)
	require.NotNil(t, outputter.frame)
}

func TestPipeline_OutputError(t *testing.T) {
	boomErr := errors.New("boom")
	outputter := &testOutputter{err: boomErr}
	p, err := New(&testRuleGetter{
		rules: map[string]*LiveChannelRule{
			"stream/test/xxx": {
				Converter: &testConverter{"", data.NewFrame("test")},
				Processor: &testProcessor{},
				Outputter: outputter,
			},
		},
	})
	require.NoError(t, err)
	_, err = p.ProcessInput(context.Background(), 1, "stream/test/xxx", []byte(`{}`))
	require.ErrorIs(t, err, boomErr)
}
