// Copyright 2013 Martini Authors
// Copyright 2014 The Macaron Authors
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
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"

	"gopkg.in/macaron.v1"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/contexthandler"
	"github.com/grafana/grafana/pkg/setting"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// stack returns a nicely formatted stack frame, skipping skip frames
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			// We can ignore the gosec G304 warning on this one because `file`
			// comes from the runtime.Caller() function.
			// nolint:gosec
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.ReplaceAll(name, centerDot, dot)
	return name
}

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
// While Martini is in development mode, Recovery will also output the panic as HTML.
func Recovery(cfg *setting.Cfg) macaron.Handler {
	return func(c *macaron.Context) {
		defer func() {
			if r := recover(); r != nil {
				panicLogger := log.Root
				// try to get request logger
				ctx := contexthandler.FromContext(c.Req.Context())
				if ctx != nil {
					panicLogger = ctx.Logger
				}

				if err, ok := r.(error); ok {
					// http.ErrAbortHandler is suppressed by default in the http package
					// and used as a signal for aborting requests. Suppresses stacktrace
					// since it doesn't add any important information.
					if errors.Is(err, http.ErrAbortHandler) {
						panicLogger.Error("Request error", "error", err)
						return
					}
				}

				stack := stack(3)
				panicLogger.Error("Request error", "error", r, "stack", string(stack))

				// if response has already been written, skip.
				if c.Resp.Written() {
					return
				}

				data := struct {
					Title     string
					AppTitle  string
					AppSubUrl string
					Theme     string
					ErrorMsg  string
				}{"Server Error", "Grafana", cfg.AppSubURL, cfg.DefaultTheme, ""}

				if setting.Env == setting.Dev {
					if err, ok := r.(error); ok {
						data.Title = err.Error()
					}

					data.ErrorMsg = string(stack)
				}

				if ctx != nil && ctx.IsApiRequest() {
					resp := make(map[string]interface{})
					resp["message"] = "Internal Server Error - Check the Grafana server logs for the detailed error message."

					if data.ErrorMsg != "" {
						resp["error"] = fmt.Sprintf("%v - %v", data.Title, data.ErrorMsg)
					} else {
						resp["error"] = data.Title
					}

					c.JSON(500, resp)
				} else {
					c.HTML(500, cfg.ErrTemplateName, data)
				}
			}
		}()

		c.Next()
	}
}
