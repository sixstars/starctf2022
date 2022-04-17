package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/spyzhov/ajson"
)

type ExactJsonConverterConfig struct {
	Fields []Field `json:"fields"`
}

// ExactJsonConverter can convert JSON to a single data.Frame according to
// user-defined field configuration and value extraction rules.
type ExactJsonConverter struct {
	config      ExactJsonConverterConfig
	nowTimeFunc func() time.Time
}

func NewExactJsonConverter(c ExactJsonConverterConfig) *ExactJsonConverter {
	return &ExactJsonConverter{config: c}
}

func (c *ExactJsonConverter) Convert(_ context.Context, vars Vars, body []byte) ([]*ChannelFrame, error) {
	//obj, err := oj.Parse(body)
	//if err != nil {
	//	return nil, err
	//}

	var fields []*data.Field

	var initGojaOnce sync.Once
	var gojaRuntime *gojaRuntime

	for _, f := range c.config.Fields {
		field := data.NewFieldFromFieldType(f.Type, 1)
		field.Name = f.Name
		field.Config = f.Config

		if strings.HasPrefix(f.Value, "$") {
			// JSON path.
			nodes, err := ajson.JSONPath(body, f.Value)
			if err != nil {
				return nil, err
			}
			if len(nodes) == 0 {
				field.Set(0, nil)
			} else if len(nodes) == 1 {
				val, err := nodes[0].Value()
				if err != nil {
					return nil, err
				}
				switch f.Type {
				case data.FieldTypeNullableFloat64:
					if val == nil {
						field.Set(0, nil)
					} else {
						switch v := val.(type) {
						case float64:
							field.SetConcrete(0, v)
						case int64:
							field.SetConcrete(0, float64(v))
						default:
							return nil, errors.New("malformed float64 type for: " + f.Name)
						}
					}
				case data.FieldTypeNullableString:
					v, ok := val.(string)
					if !ok {
						return nil, errors.New("malformed string type")
					}
					field.SetConcrete(0, v)
				default:
					return nil, fmt.Errorf("unsupported field type: %s (%s)", f.Type, f.Name)
				}
			} else {
				return nil, errors.New("too many values")
			}
			//x, err := jp.ParseString(f.Value[1:])
			//if err != nil {
			//	return nil, err
			//}
			//value := x.Get(obj)
			//if len(value) == 0 {
			//	field.Set(0, nil)
			//} else if len(value) == 1 {
			//	val := value[0]
			//	switch f.Type {
			//	case data.FieldTypeNullableFloat64:
			//		if val == nil {
			//			field.Set(0, nil)
			//		} else {
			//			switch v := val.(type) {
			//			case float64:
			//				field.SetConcrete(0, v)
			//			case int64:
			//				field.SetConcrete(0, float64(v))
			//			default:
			//				return nil, errors.New("malformed float64 type for: " + f.Name)
			//			}
			//		}
			//	case data.FieldTypeNullableString:
			//		v, ok := val.(string)
			//		if !ok {
			//			return nil, errors.New("malformed string type")
			//		}
			//		field.SetConcrete(0, v)
			//	default:
			//		return nil, fmt.Errorf("unsupported field type: %s (%s)", f.Type, f.Name)
			//	}
			//} else {
			//	return nil, errors.New("too many values")
			//}
		} else if strings.HasPrefix(f.Value, "{") {
			// Goja script.
			script := strings.Trim(f.Value, "{}")
			var err error
			initGojaOnce.Do(func() {
				gojaRuntime, err = getRuntime(body)
			})
			if err != nil {
				return nil, err
			}
			switch f.Type {
			case data.FieldTypeNullableBool:
				v, err := gojaRuntime.getBool(script)
				if err != nil {
					return nil, err
				}
				field.SetConcrete(0, v)
			case data.FieldTypeNullableFloat64:
				v, err := gojaRuntime.getFloat64(script)
				if err != nil {
					return nil, err
				}
				field.SetConcrete(0, v)
			default:
				return nil, fmt.Errorf("unsupported field type: %s (%s)", f.Type, f.Name)
			}
		} else if f.Value == "#{now}" {
			// Variable.
			// TODO: make consistent with Grafana variables?
			nowTimeFunc := c.nowTimeFunc
			if nowTimeFunc == nil {
				nowTimeFunc = time.Now
			}
			field.SetConcrete(0, nowTimeFunc())
		}

		labels := map[string]string{}
		for _, label := range f.Labels {
			if strings.HasPrefix(label.Value, "$") {
				nodes, err := ajson.JSONPath(body, label.Value)
				if err != nil {
					return nil, err
				}
				if len(nodes) == 0 {
					labels[label.Name] = ""
				} else if len(nodes) == 1 {
					value, err := nodes[0].Value()
					if err != nil {
						return nil, err
					}
					labels[label.Name] = fmt.Sprintf("%v", value)
				} else {
					return nil, errors.New("too many values for a label")
				}
				//x, err := jp.ParseString(label.Value[1:])
				//if err != nil {
				//	return nil, err
				//}
				//value := x.Get(obj)
				//if len(value) == 0 {
				//	labels[label.Name] = ""
				//} else if len(value) == 1 {
				//	labels[label.Name] = fmt.Sprintf("%v", value[0])
				//} else {
				//	return nil, errors.New("too many values for a label")
				//}
			} else if strings.HasPrefix(label.Value, "{") {
				script := strings.Trim(label.Value, "{}")
				var err error
				initGojaOnce.Do(func() {
					gojaRuntime, err = getRuntime(body)
				})
				if err != nil {
					return nil, err
				}
				v, err := gojaRuntime.getString(script)
				if err != nil {
					return nil, err
				}
				labels[label.Name] = v
			} else {
				labels[label.Name] = label.Value
			}
		}
		field.Labels = labels
		fields = append(fields, field)
	}

	frame := data.NewFrame(vars.Path, fields...)
	return []*ChannelFrame{
		{Channel: "", Frame: frame},
	}, nil
}
