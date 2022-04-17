package rendering

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var netTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}

var netClient = &http.Client{
	Transport: netTransport,
}

func (rs *RenderingService) renderViaHTTP(ctx context.Context, renderKey string, opts Opts) (*RenderResult, error) {
	filePath, err := rs.getNewFilePath(RenderPNG)
	if err != nil {
		return nil, err
	}

	rendererURL, err := url.Parse(rs.Cfg.RendererUrl)
	if err != nil {
		return nil, err
	}

	queryParams := rendererURL.Query()
	queryParams.Add("url", rs.getURL(opts.Path))
	queryParams.Add("renderKey", renderKey)
	queryParams.Add("width", strconv.Itoa(opts.Width))
	queryParams.Add("height", strconv.Itoa(opts.Height))
	queryParams.Add("domain", rs.domain)
	queryParams.Add("timezone", isoTimeOffsetToPosixTz(opts.Timezone))
	queryParams.Add("encoding", opts.Encoding)
	queryParams.Add("timeout", strconv.Itoa(int(opts.Timeout.Seconds())))
	queryParams.Add("deviceScaleFactor", fmt.Sprintf("%f", opts.DeviceScaleFactor))

	rendererURL.RawQuery = queryParams.Encode()

	// gives service some additional time to timeout and return possible errors.
	reqContext, cancel := context.WithTimeout(ctx, opts.Timeout+time.Second*2)
	defer cancel()

	resp, err := rs.doRequest(reqContext, rendererURL, opts.Headers)
	if err != nil {
		return nil, err
	}

	// save response to file
	defer func() {
		if err := resp.Body.Close(); err != nil {
			rs.log.Warn("Failed to close response body", "err", err)
		}
	}()

	err = rs.readFileResponse(reqContext, resp, filePath)
	if err != nil {
		return nil, err
	}

	return &RenderResult{FilePath: filePath}, nil
}

func (rs *RenderingService) renderCSVViaHTTP(ctx context.Context, renderKey string, opts CSVOpts) (*RenderCSVResult, error) {
	filePath, err := rs.getNewFilePath(RenderCSV)
	if err != nil {
		return nil, err
	}

	rendererURL, err := url.Parse(rs.Cfg.RendererUrl + "/csv")
	if err != nil {
		return nil, err
	}

	queryParams := rendererURL.Query()
	queryParams.Add("url", rs.getURL(opts.Path))
	queryParams.Add("renderKey", renderKey)
	queryParams.Add("domain", rs.domain)
	queryParams.Add("timezone", isoTimeOffsetToPosixTz(opts.Timezone))
	queryParams.Add("encoding", opts.Encoding)
	queryParams.Add("timeout", strconv.Itoa(int(opts.Timeout.Seconds())))

	rendererURL.RawQuery = queryParams.Encode()

	// gives service some additional time to timeout and return possible errors.
	reqContext, cancel := context.WithTimeout(ctx, opts.Timeout+time.Second*2)
	defer cancel()

	resp, err := rs.doRequest(reqContext, rendererURL, opts.Headers)
	if err != nil {
		return nil, err
	}

	// save response to file
	defer func() {
		if err := resp.Body.Close(); err != nil {
			rs.log.Warn("Failed to close response body", "err", err)
		}
	}()

	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	if err != nil {
		return nil, err
	}
	downloadFileName := params["filename"]

	err = rs.readFileResponse(reqContext, resp, filePath)
	if err != nil {
		return nil, err
	}

	return &RenderCSVResult{FilePath: filePath, FileName: downloadFileName}, nil
}

func (rs *RenderingService) doRequest(ctx context.Context, url *url.URL, headers map[string][]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("Grafana/%s", rs.Cfg.BuildVersion))
	for k, v := range headers {
		req.Header[k] = v
	}

	rs.log.Debug("calling remote rendering service", "url", url)

	// make request to renderer server
	resp, err := netClient.Do(req)
	if err != nil {
		rs.log.Error("Failed to send request to remote rendering service", "error", err)
		return nil, fmt.Errorf("failed to send request to remote rendering service: %w", err)
	}

	return resp, nil
}

func (rs *RenderingService) readFileResponse(ctx context.Context, resp *http.Response, filePath string) error {
	// check for timeout first
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		rs.log.Info("Rendering timed out")
		return ErrTimeout
	}

	// if we didn't get a 200 response, something went wrong.
	if resp.StatusCode != http.StatusOK {
		rs.log.Error("Remote rendering request failed", "error", resp.Status)
		return fmt.Errorf("remote rendering request failed, status code: %d, status: %s", resp.StatusCode,
			resp.Status)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer func() {
		if err := out.Close(); err != nil && !errors.Is(err, fs.ErrClosed) {
			// We already close the file explicitly in the non-error path, so shouldn't be a problem
			rs.log.Warn("Failed to close file", "path", filePath, "err", err)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		// check that we didn't timeout while receiving the response.
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			rs.log.Info("Rendering timed out")
			return ErrTimeout
		}

		rs.log.Error("Remote rendering request failed", "error", err)
		return fmt.Errorf("remote rendering request failed: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("failed to write to %q: %w", filePath, err)
	}

	return nil
}

func (rs *RenderingService) getRemotePluginVersion() (string, error) {
	rendererURL, err := url.Parse(rs.Cfg.RendererUrl + "/version")
	if err != nil {
		return "", err
	}

	headers := make(map[string][]string)
	resp, err := rs.doRequest(context.Background(), rendererURL, headers)
	if err != nil {
		return "", err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			rs.log.Warn("Failed to close response body", "err", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("remote rendering request to get version failed, status code: %d, status: %s", resp.StatusCode,
			resp.Status)
	}

	var info struct {
		Version string
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	return info.Version, nil
}
