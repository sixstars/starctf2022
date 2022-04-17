package server

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/server/backgroundsvcs"
	"github.com/grafana/grafana/pkg/services/accesscontrol/ossaccesscontrol"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/stretchr/testify/require"
)

type testService struct {
	started    chan struct{}
	runErr     error
	isDisabled bool
}

func newTestService(runErr error, disabled bool) *testService {
	return &testService{
		started:    make(chan struct{}),
		runErr:     runErr,
		isDisabled: disabled,
	}
}

func (s *testService) Run(ctx context.Context) error {
	if s.isDisabled {
		return fmt.Errorf("Shouldn't run disabled service")
	}

	if s.runErr != nil {
		return s.runErr
	}
	close(s.started)
	<-ctx.Done()
	return ctx.Err()
}

func (s *testService) IsDisabled() bool {
	return s.isDisabled
}

func testServer(t *testing.T, services ...registry.BackgroundService) *Server {
	t.Helper()
	s, err := newServer(Options{}, setting.NewCfg(), nil, &ossaccesscontrol.OSSAccessControlService{}, nil, backgroundsvcs.NewBackgroundServiceRegistry(services...))
	require.NoError(t, err)
	// Required to skip configuration initialization that causes
	// DI errors in this test.
	s.isInitialized = true
	return s
}

func TestServer_Run_Error(t *testing.T) {
	testErr := errors.New("boom")
	s := testServer(t, newTestService(nil, false), newTestService(testErr, false))
	err := s.Run()
	require.ErrorIs(t, err, testErr)
	require.NotZero(t, s.ExitCode(err))
}

func TestServer_Shutdown(t *testing.T) {
	ctx := context.Background()

	s := testServer(t, newTestService(nil, false), newTestService(nil, true))

	ch := make(chan error)

	go func() {
		defer close(ch)

		// Wait until all services launched.
		for _, svc := range s.backgroundServices {
			if !svc.(*testService).isDisabled {
				<-svc.(*testService).started
			}
		}
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		err := s.Shutdown(ctx, "test interrupt")
		ch <- err
	}()
	err := s.Run()
	require.NoError(t, err)
	require.Zero(t, s.ExitCode(err))

	err = <-ch
	require.NoError(t, err)
}
