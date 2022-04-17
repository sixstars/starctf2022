package api

import (
	"errors"
	"time"

	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/apikeygen"
	"github.com/grafana/grafana/pkg/models"
)

func GetAPIKeys(c *models.ReqContext) response.Response {
	query := models.GetApiKeysQuery{OrgId: c.OrgId, IncludeExpired: c.QueryBool("includeExpired")}

	if err := bus.Dispatch(&query); err != nil {
		return response.Error(500, "Failed to list api keys", err)
	}

	result := make([]*models.ApiKeyDTO, len(query.Result))
	for i, t := range query.Result {
		var expiration *time.Time = nil
		if t.Expires != nil {
			v := time.Unix(*t.Expires, 0)
			expiration = &v
		}
		result[i] = &models.ApiKeyDTO{
			Id:         t.Id,
			Name:       t.Name,
			Role:       t.Role,
			Expiration: expiration,
		}
	}

	return response.JSON(200, result)
}

func DeleteAPIKey(c *models.ReqContext) response.Response {
	id := c.ParamsInt64(":id")

	cmd := &models.DeleteApiKeyCommand{Id: id, OrgId: c.OrgId}

	err := bus.Dispatch(cmd)
	if err != nil {
		var status int
		if errors.Is(err, models.ErrApiKeyNotFound) {
			status = 404
		} else {
			status = 500
		}
		return response.Error(status, "Failed to delete API key", err)
	}

	return response.Success("API key deleted")
}

func (hs *HTTPServer) AddAPIKey(c *models.ReqContext, cmd models.AddApiKeyCommand) response.Response {
	if !cmd.Role.IsValid() {
		return response.Error(400, "Invalid role specified", nil)
	}

	if hs.Cfg.ApiKeyMaxSecondsToLive != -1 {
		if cmd.SecondsToLive == 0 {
			return response.Error(400, "Number of seconds before expiration should be set", nil)
		}
		if cmd.SecondsToLive > hs.Cfg.ApiKeyMaxSecondsToLive {
			return response.Error(400, "Number of seconds before expiration is greater than the global limit", nil)
		}
	}
	cmd.OrgId = c.OrgId

	newKeyInfo, err := apikeygen.New(cmd.OrgId, cmd.Name)
	if err != nil {
		return response.Error(500, "Generating API key failed", err)
	}

	cmd.Key = newKeyInfo.HashedKey

	if err := bus.Dispatch(&cmd); err != nil {
		if errors.Is(err, models.ErrInvalidApiKeyExpiration) {
			return response.Error(400, err.Error(), nil)
		}
		if errors.Is(err, models.ErrDuplicateApiKey) {
			return response.Error(409, err.Error(), nil)
		}
		return response.Error(500, "Failed to add API Key", err)
	}

	result := &dtos.NewApiKeyResult{
		ID:   cmd.Result.Id,
		Name: cmd.Result.Name,
		Key:  newKeyInfo.ClientSecret,
	}

	return response.JSON(200, result)
}
