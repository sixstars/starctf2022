package pipeline

import (
	"context"
	"reflect"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type ChangeLogOutputConfig struct {
	FieldName string `json:"fieldName"`
	Channel   string `json:"channel"`
}

// ChangeLogOutput can monitor value changes of the specified field and output
// special change frame to the configured channel.
type ChangeLogOutput struct {
	frameStorage FrameGetSetter
	config       ChangeLogOutputConfig
}

func NewChangeLogOutput(frameStorage FrameGetSetter, config ChangeLogOutputConfig) *ChangeLogOutput {
	return &ChangeLogOutput{frameStorage: frameStorage, config: config}
}

func (l ChangeLogOutput) Output(_ context.Context, vars OutputVars, frame *data.Frame) ([]*ChannelFrame, error) {
	previousFrame, previousFrameOK, err := l.frameStorage.Get(vars.OrgID, l.config.Channel)
	if err != nil {
		return nil, err
	}

	fieldName := l.config.FieldName

	previousFrameFieldIndex := -1
	if previousFrameOK {
		for i, f := range previousFrame.Fields {
			if f.Name == fieldName {
				previousFrameFieldIndex = i
			}
		}
	}

	currentFrameFieldIndex := -1
	for i, f := range frame.Fields {
		if f.Name == fieldName {
			currentFrameFieldIndex = i
		}
	}

	var previousValue interface{}
	if previousFrameFieldIndex >= 0 {
		// Take last value for the field.
		previousValue = previousFrame.Fields[previousFrameFieldIndex].At(previousFrame.Fields[previousFrameFieldIndex].Len() - 1)
	}

	fTime := data.NewFieldFromFieldType(data.FieldTypeTime, 0)
	fTime.Name = "time"
	f1 := data.NewFieldFromFieldType(frame.Fields[currentFrameFieldIndex].Type(), 0)
	f1.Name = "old"
	f2 := data.NewFieldFromFieldType(frame.Fields[currentFrameFieldIndex].Type(), 0)
	f2.Name = "new"

	if currentFrameFieldIndex >= 0 {
		for i := 0; i < frame.Fields[currentFrameFieldIndex].Len(); i++ {
			currentValue := frame.Fields[currentFrameFieldIndex].At(i)
			if !reflect.DeepEqual(
				previousValue,
				currentValue,
			) {
				fTime.Append(time.Now())
				f1.Append(previousValue)
				f2.Append(currentValue)
				previousValue = currentValue
			}
		}
	}

	if fTime.Len() > 0 {
		changeFrame := data.NewFrame("change", fTime, f1, f2)
		err := l.frameStorage.Set(vars.OrgID, l.config.Channel, frame)
		if err != nil {
			return nil, err
		}
		return []*ChannelFrame{{
			Channel: l.config.Channel,
			Frame:   changeFrame,
		}}, nil
	}

	return nil, l.frameStorage.Set(vars.OrgID, l.config.Channel, frame)
}
