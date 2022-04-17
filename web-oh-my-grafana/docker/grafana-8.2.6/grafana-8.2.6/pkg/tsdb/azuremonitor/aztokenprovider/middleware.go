package aztokenprovider

import (
	"fmt"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
)

const authenticationMiddlewareName = "AzureAuthentication"

func AuthMiddleware(tokenProvider AzureTokenProvider, scopes []string) httpclient.Middleware {
	return httpclient.NamedMiddlewareFunc(authenticationMiddlewareName, func(opts httpclient.Options, next http.RoundTripper) http.RoundTripper {
		return ApplyAuth(tokenProvider, scopes, next)
	})
}

func ApplyAuth(tokenProvider AzureTokenProvider, scopes []string, next http.RoundTripper) http.RoundTripper {
	return httpclient.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		token, err := tokenProvider.GetAccessToken(req.Context(), scopes)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve Azure access token: %w", err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		return next.RoundTrip(req)
	})
}
