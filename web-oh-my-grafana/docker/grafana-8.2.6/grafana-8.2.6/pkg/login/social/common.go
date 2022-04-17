package social

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/util/errutil"
	"github.com/jmespath/go-jmespath"
)

var (
	errMissingGroupMembership = Error{"user not a member of one of the required groups"}
)

type httpGetResponse struct {
	Body    []byte
	Headers http.Header
}

func (s *SocialBase) IsEmailAllowed(email string) bool {
	return isEmailAllowed(email, s.allowedDomains)
}

func (s *SocialBase) IsSignupAllowed() bool {
	return s.allowSignup
}

func isEmailAllowed(email string, allowedDomains []string) bool {
	if len(allowedDomains) == 0 {
		return true
	}

	valid := false
	for _, domain := range allowedDomains {
		emailSuffix := fmt.Sprintf("@%s", domain)
		valid = valid || strings.HasSuffix(email, emailSuffix)
	}

	return valid
}

func (s *SocialBase) httpGet(client *http.Client, url string) (response httpGetResponse, err error) {
	r, err := client.Get(url)
	if err != nil {
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			s.log.Warn("Failed to close response body", "err", err)
		}
	}()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	response = httpGetResponse{body, r.Header}

	if r.StatusCode >= 300 {
		err = fmt.Errorf(string(response.Body))
		return
	}

	log.Tracef("HTTP GET %s: %s %s", url, r.Status, string(response.Body))

	err = nil
	return
}

func (s *SocialBase) searchJSONForAttr(attributePath string, data []byte) (interface{}, error) {
	if attributePath == "" {
		return "", errors.New("no attribute path specified")
	}

	if len(data) == 0 {
		return "", errors.New("empty user info JSON response provided")
	}

	var buf interface{}
	if err := json.Unmarshal(data, &buf); err != nil {
		return "", errutil.Wrap("failed to unmarshal user info JSON response", err)
	}

	val, err := jmespath.Search(attributePath, buf)
	if err != nil {
		return "", errutil.Wrapf(err, "failed to search user info JSON response with provided path: %q", attributePath)
	}

	return val, nil
}

func (s *SocialBase) searchJSONForStringAttr(attributePath string, data []byte) (string, error) {
	val, err := s.searchJSONForAttr(attributePath, data)
	if err != nil {
		return "", err
	}

	strVal, ok := val.(string)
	if ok {
		return strVal, nil
	}

	return "", nil
}

func (s *SocialBase) searchJSONForStringArrayAttr(attributePath string, data []byte) ([]string, error) {
	val, err := s.searchJSONForAttr(attributePath, data)
	if err != nil {
		return []string{}, err
	}

	ifArr, ok := val.([]interface{})
	if !ok {
		return []string{}, nil
	}

	result := []string{}
	for _, v := range ifArr {
		if strVal, ok := v.(string); ok {
			result = append(result, strVal)
		}
	}

	return result, nil
}
