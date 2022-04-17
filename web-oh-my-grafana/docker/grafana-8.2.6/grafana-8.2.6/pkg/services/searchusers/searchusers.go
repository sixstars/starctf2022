package searchusers

import (
	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
)

type Service interface {
	SearchUsers(c *models.ReqContext) response.Response
	SearchUsersWithPaging(c *models.ReqContext) response.Response
}

type OSSService struct {
	bus bus.Bus
}

func ProvideUsersService(bus bus.Bus) *OSSService {
	return &OSSService{bus: bus}
}

func (s *OSSService) SearchUsers(c *models.ReqContext) response.Response {
	query, err := s.SearchUser(c)
	if err != nil {
		return response.Error(500, "Failed to fetch users", err)
	}

	return response.JSON(200, query.Result.Users)
}

func (s *OSSService) SearchUsersWithPaging(c *models.ReqContext) response.Response {
	query, err := s.SearchUser(c)
	if err != nil {
		return response.Error(500, "Failed to fetch users", err)
	}

	return response.JSON(200, query.Result)
}

func (s *OSSService) SearchUser(c *models.ReqContext) (*models.SearchUsersQuery, error) {
	perPage := c.QueryInt("perpage")
	if perPage <= 0 {
		perPage = 1000
	}
	page := c.QueryInt("page")

	if page < 1 {
		page = 1
	}

	searchQuery := c.Query("query")
	filter := c.Query("filter")

	query := &models.SearchUsersQuery{Query: searchQuery, Filter: models.SearchUsersFilter(filter), Page: page, Limit: perPage}
	if err := s.bus.Dispatch(query); err != nil {
		return nil, err
	}

	for _, user := range query.Result.Users {
		user.AvatarUrl = dtos.GetGravatarUrl(user.Email)
		user.AuthLabels = make([]string, 0)
		if user.AuthModule != nil && len(user.AuthModule) > 0 {
			for _, authModule := range user.AuthModule {
				user.AuthLabels = append(user.AuthLabels, GetAuthProviderLabel(authModule))
			}
		}
	}

	query.Result.Page = page
	query.Result.PerPage = perPage

	return query, nil
}

func GetAuthProviderLabel(authModule string) string {
	switch authModule {
	case "oauth_github":
		return "GitHub"
	case "oauth_google":
		return "Google"
	case "oauth_azuread":
		return "AzureAD"
	case "oauth_gitlab":
		return "GitLab"
	case "oauth_grafana_com", "oauth_grafananet":
		return "grafana.com"
	case "auth.saml":
		return "SAML"
	case "ldap", "":
		return "LDAP"
	default:
		return "OAuth"
	}
}
