package tracing

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/setting"

	opentracing "github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
)

const (
	envJaegerAgentHost = "JAEGER_AGENT_HOST"
	envJaegerAgentPort = "JAEGER_AGENT_PORT"
)

func ProvideService(cfg *setting.Cfg) (*TracingService, error) {
	ts := &TracingService{
		Cfg: cfg,
		log: log.New("tracing"),
	}
	if err := ts.parseSettings(); err != nil {
		return nil, err
	}

	if ts.enabled {
		return ts, ts.initGlobalTracer()
	}

	return ts, nil
}

type TracingService struct {
	enabled                  bool
	address                  string
	customTags               map[string]string
	samplerType              string
	samplerParam             float64
	samplingServerURL        string
	log                      log.Logger
	closer                   io.Closer
	zipkinPropagation        bool
	disableSharedZipkinSpans bool

	Cfg *setting.Cfg
}

func (ts *TracingService) parseSettings() error {
	var section, err = ts.Cfg.Raw.GetSection("tracing.jaeger")
	if err != nil {
		return err
	}

	ts.address = section.Key("address").MustString("")
	if ts.address == "" {
		host := os.Getenv(envJaegerAgentHost)
		port := os.Getenv(envJaegerAgentPort)
		if host != "" || port != "" {
			ts.address = fmt.Sprintf("%s:%s", host, port)
		}
	}
	if ts.address != "" {
		ts.enabled = true
	}

	ts.customTags = splitTagSettings(section.Key("always_included_tag").MustString(""))
	ts.samplerType = section.Key("sampler_type").MustString("")
	ts.samplerParam = section.Key("sampler_param").MustFloat64(1)
	ts.zipkinPropagation = section.Key("zipkin_propagation").MustBool(false)
	ts.disableSharedZipkinSpans = section.Key("disable_shared_zipkin_spans").MustBool(false)
	ts.samplingServerURL = section.Key("sampling_server_url").MustString("")
	return nil
}

func (ts *TracingService) initJaegerCfg() (jaegercfg.Configuration, error) {
	cfg := jaegercfg.Configuration{
		ServiceName: "grafana",
		Disabled:    !ts.enabled,
		Sampler: &jaegercfg.SamplerConfig{
			Type:              ts.samplerType,
			Param:             ts.samplerParam,
			SamplingServerURL: ts.samplingServerURL,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           false,
			LocalAgentHostPort: ts.address,
		},
	}

	_, err := cfg.FromEnv()
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (ts *TracingService) initGlobalTracer() error {
	cfg, err := ts.initJaegerCfg()
	if err != nil {
		return err
	}

	jLogger := &jaegerLogWrapper{logger: log.New("jaeger")}

	options := []jaegercfg.Option{}
	options = append(options, jaegercfg.Logger(jLogger))

	for tag, value := range ts.customTags {
		options = append(options, jaegercfg.Tag(tag, value))
	}

	if ts.zipkinPropagation {
		zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
		options = append(options,
			jaegercfg.Injector(opentracing.HTTPHeaders, zipkinPropagator),
			jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		)

		if !ts.disableSharedZipkinSpans {
			options = append(options, jaegercfg.ZipkinSharedRPCSpan(true))
		}
	}

	tracer, closer, err := cfg.NewTracer(options...)
	if err != nil {
		return err
	}

	opentracing.SetGlobalTracer(tracer)

	ts.closer = closer

	return nil
}

func (ts *TracingService) Run(ctx context.Context) error {
	<-ctx.Done()

	if ts.closer != nil {
		ts.log.Info("Closing tracing")
		return ts.closer.Close()
	}

	return nil
}

func splitTagSettings(input string) map[string]string {
	res := map[string]string{}

	tags := strings.Split(input, ",")
	for _, v := range tags {
		kv := strings.Split(v, ":")
		if len(kv) > 1 {
			res[kv[0]] = kv[1]
		}
	}

	return res
}

type jaegerLogWrapper struct {
	logger log.Logger
}

func (jlw *jaegerLogWrapper) Error(msg string) {
	jlw.logger.Error(msg)
}

func (jlw *jaegerLogWrapper) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	jlw.logger.Info(msg)
}
