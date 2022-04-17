package tempo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/translator/conventions"
	tracetranslator "go.opentelemetry.io/collector/translator/trace"
)

type KeyValue struct {
	Value interface{} `json:"value"`
	Key   string      `json:"key"`
}

type TraceLog struct {
	// Millisecond epoch time
	Timestamp float64     `json:"timestamp"`
	Fields    []*KeyValue `json:"fields"`
}

func TraceToFrame(td pdata.Traces) (*data.Frame, error) {
	// In open telemetry format the spans are grouped first by resource/service they originated in and inside that
	// resource they are grouped by the instrumentation library which created them.

	resourceSpans := td.ResourceSpans()

	if resourceSpans.Len() == 0 {
		return nil, nil
	}

	frame := &data.Frame{
		Name: "Trace",
		Fields: []*data.Field{
			data.NewField("traceID", nil, []string{}),
			data.NewField("spanID", nil, []string{}),
			data.NewField("parentSpanID", nil, []string{}),
			data.NewField("operationName", nil, []string{}),
			data.NewField("serviceName", nil, []string{}),
			data.NewField("serviceTags", nil, []string{}),
			data.NewField("startTime", nil, []float64{}),
			data.NewField("duration", nil, []float64{}),
			data.NewField("logs", nil, []string{}),
			data.NewField("tags", nil, []string{}),
		},
		Meta: &data.FrameMeta{
			// TODO: use constant once available in the SDK
			PreferredVisualization: "trace",
		},
	}

	for i := 0; i < resourceSpans.Len(); i++ {
		rs := resourceSpans.At(i)
		rows, err := resourceSpansToRows(rs)
		if err != nil {
			return nil, err
		}

		for _, row := range rows {
			frame.AppendRow(row...)
		}
	}

	return frame, nil
}

// resourceSpansToRows processes all the spans for a particular resource/service
func resourceSpansToRows(rs pdata.ResourceSpans) ([][]interface{}, error) {
	resource := rs.Resource()
	ilss := rs.InstrumentationLibrarySpans()

	if resource.Attributes().Len() == 0 || ilss.Len() == 0 {
		return [][]interface{}{}, nil
	}

	// Approximate the number of the spans as the number of the spans in the first
	// instrumentation library info.
	rows := make([][]interface{}, 0, ilss.At(0).Spans().Len())

	for i := 0; i < ilss.Len(); i++ {
		ils := ilss.At(i)

		// These are finally the actual spans
		spans := ils.Spans()

		for j := 0; j < spans.Len(); j++ {
			span := spans.At(j)
			row, err := spanToSpanRow(span, ils.InstrumentationLibrary(), resource)
			if err != nil {
				return nil, err
			}
			if row != nil {
				rows = append(rows, row)
			}
		}
	}

	return rows, nil
}

func spanToSpanRow(span pdata.Span, libraryTags pdata.InstrumentationLibrary, resource pdata.Resource) ([]interface{}, error) {
	// If the id representation changed from hexstring to something else we need to change the transformBase64IDToHexString in the frontend code
	traceID := span.TraceID().HexString()
	traceID = strings.TrimLeft(traceID, "0")

	spanID := span.SpanID().HexString()

	parentSpanID := span.ParentSpanID().HexString()
	startTime := float64(span.StartTimestamp()) / 1_000_000
	serviceName, serviceTags := resourceToProcess(resource)

	serviceTagsJson, err := json.Marshal(serviceTags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal service tags: %w", err)
	}

	spanTags, err := json.Marshal(getSpanTags(span, libraryTags))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal span tags: %w", err)
	}

	logs, err := json.Marshal(spanEventsToLogs(span.Events()))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal span logs: %w", err)
	}

	return []interface{}{
		traceID,
		spanID,
		parentSpanID,
		span.Name(),
		serviceName,
		toJSONString(serviceTagsJson),
		startTime,
		float64(span.EndTimestamp()-span.StartTimestamp()) / 1_000_000,
		toJSONString(logs),
		toJSONString(spanTags),
	}, nil
}

func toJSONString(json []byte) string {
	s := string(json)
	if s == "null" {
		return ""
	}
	return s
}

func resourceToProcess(resource pdata.Resource) (string, []*KeyValue) {
	attrs := resource.Attributes()
	serviceName := tracetranslator.ResourceNoServiceName
	if attrs.Len() == 0 {
		return serviceName, nil
	}

	tags := make([]*KeyValue, 0, attrs.Len()-1)
	attrs.Range(func(key string, attr pdata.AttributeValue) bool {
		if key == conventions.AttributeServiceName {
			serviceName = attr.StringVal()
		}
		tags = append(tags, &KeyValue{Key: key, Value: getAttributeVal(attr)})
		return true
	})

	return serviceName, tags
}

func getAttributeVal(attr pdata.AttributeValue) interface{} {
	switch attr.Type() {
	case pdata.AttributeValueTypeString:
		return attr.StringVal()
	case pdata.AttributeValueTypeInt:
		return attr.IntVal()
	case pdata.AttributeValueTypeBool:
		return attr.BoolVal()
	case pdata.AttributeValueTypeDouble:
		return attr.DoubleVal()
	case pdata.AttributeValueTypeMap, pdata.AttributeValueTypeArray:
		return tracetranslator.AttributeValueToString(attr)
	default:
		return nil
	}
}

func getSpanTags(span pdata.Span, instrumentationLibrary pdata.InstrumentationLibrary) []*KeyValue {
	var tags []*KeyValue

	libraryTags := getTagsFromInstrumentationLibrary(instrumentationLibrary)
	if libraryTags != nil {
		tags = append(tags, libraryTags...)
	}
	span.Attributes().Range(func(key string, attr pdata.AttributeValue) bool {
		tags = append(tags, &KeyValue{Key: key, Value: getAttributeVal(attr)})
		return true
	})

	status := span.Status()
	possibleNilTags := []*KeyValue{
		getTagFromSpanKind(span.Kind()),
		getTagFromStatusCode(status.Code()),
		getErrorTagFromStatusCode(status.Code()),
		getTagFromStatusMsg(status.Message()),
		getTagFromTraceState(span.TraceState()),
	}

	for _, tag := range possibleNilTags {
		if tag != nil {
			tags = append(tags, tag)
		}
	}
	return tags
}

func getTagsFromInstrumentationLibrary(il pdata.InstrumentationLibrary) []*KeyValue {
	var keyValues []*KeyValue
	if ilName := il.Name(); ilName != "" {
		kv := &KeyValue{
			Key:   conventions.InstrumentationLibraryName,
			Value: ilName,
		}
		keyValues = append(keyValues, kv)
	}
	if ilVersion := il.Version(); ilVersion != "" {
		kv := &KeyValue{
			Key:   conventions.InstrumentationLibraryVersion,
			Value: ilVersion,
		}
		keyValues = append(keyValues, kv)
	}

	return keyValues
}

func getTagFromSpanKind(spanKind pdata.SpanKind) *KeyValue {
	var tagStr string
	switch spanKind {
	case pdata.SpanKindClient:
		tagStr = string(tracetranslator.OpenTracingSpanKindClient)
	case pdata.SpanKindServer:
		tagStr = string(tracetranslator.OpenTracingSpanKindServer)
	case pdata.SpanKindProducer:
		tagStr = string(tracetranslator.OpenTracingSpanKindProducer)
	case pdata.SpanKindConsumer:
		tagStr = string(tracetranslator.OpenTracingSpanKindConsumer)
	case pdata.SpanKindInternal:
		tagStr = string(tracetranslator.OpenTracingSpanKindInternal)
	default:
		return nil
	}

	return &KeyValue{
		Key:   tracetranslator.TagSpanKind,
		Value: tagStr,
	}
}

func getTagFromStatusCode(statusCode pdata.StatusCode) *KeyValue {
	return &KeyValue{
		Key:   tracetranslator.TagStatusCode,
		Value: int64(statusCode),
	}
}

func getErrorTagFromStatusCode(statusCode pdata.StatusCode) *KeyValue {
	if statusCode == pdata.StatusCodeError {
		return &KeyValue{
			Key:   tracetranslator.TagError,
			Value: true,
		}
	}
	return nil
}

func getTagFromStatusMsg(statusMsg string) *KeyValue {
	if statusMsg == "" {
		return nil
	}
	return &KeyValue{
		Key:   tracetranslator.TagStatusMsg,
		Value: statusMsg,
	}
}

func getTagFromTraceState(traceState pdata.TraceState) *KeyValue {
	if traceState != pdata.TraceStateEmpty {
		return &KeyValue{
			Key:   tracetranslator.TagW3CTraceState,
			Value: string(traceState),
		}
	}
	return nil
}

func spanEventsToLogs(events pdata.SpanEventSlice) []*TraceLog {
	if events.Len() == 0 {
		return nil
	}

	logs := make([]*TraceLog, 0, events.Len())
	for i := 0; i < events.Len(); i++ {
		event := events.At(i)
		fields := make([]*KeyValue, 0, event.Attributes().Len()+1)
		if event.Name() != "" {
			fields = append(fields, &KeyValue{
				Key:   tracetranslator.TagMessage,
				Value: event.Name(),
			})
		}
		event.Attributes().Range(func(key string, attr pdata.AttributeValue) bool {
			fields = append(fields, &KeyValue{Key: key, Value: getAttributeVal(attr)})
			return true
		})
		logs = append(logs, &TraceLog{
			Timestamp: float64(event.Timestamp()) / 1_000_000,
			Fields:    fields,
		})
	}

	return logs
}
