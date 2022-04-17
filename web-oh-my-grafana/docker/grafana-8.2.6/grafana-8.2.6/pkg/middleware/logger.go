// Copyright 2013 Martini Authors
// Copyright 2014 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package middleware

import (
	"net/http"
	"time"

	"github.com/grafana/grafana/pkg/services/contexthandler"
	"github.com/grafana/grafana/pkg/setting"
	cw "github.com/weaveworks/common/middleware"
	"gopkg.in/macaron.v1"
)

func Logger(cfg *setting.Cfg) macaron.Handler {
	return func(res http.ResponseWriter, req *http.Request, c *macaron.Context) {
		start := time.Now()

		rw := res.(macaron.ResponseWriter)
		c.Next()

		timeTaken := time.Since(start) / time.Millisecond

		ctx := contexthandler.FromContext(c.Req.Context())
		if ctx != nil && ctx.PerfmonTimer != nil {
			ctx.PerfmonTimer.Observe(float64(timeTaken))
		}

		status := rw.Status()
		if status == 200 || status == 304 {
			if !cfg.RouterLogging {
				return
			}
		}

		if ctx != nil {
			logParams := []interface{}{
				"method", req.Method,
				"path", req.URL.Path,
				"status", status,
				"remote_addr", c.RemoteAddr(),
				"time_ms", int64(timeTaken),
				"size", rw.Size(),
				"referer", req.Referer(),
			}

			traceID, exist := cw.ExtractTraceID(ctx.Req.Context())
			if exist {
				logParams = append(logParams, "traceID", traceID)
			}

			if status >= 500 {
				ctx.Logger.Error("Request Completed", logParams...)
			} else {
				ctx.Logger.Info("Request Completed", logParams...)
			}
		}
	}
}
