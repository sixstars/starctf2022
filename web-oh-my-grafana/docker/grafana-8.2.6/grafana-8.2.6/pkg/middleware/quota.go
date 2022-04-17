package middleware

import (
	"fmt"

	"gopkg.in/macaron.v1"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/quota"
)

// Quota returns a function that returns a function used to call quotaservice based on target name
func Quota(quotaService *quota.QuotaService) func(string) macaron.Handler {
	if quotaService == nil {
		panic("quotaService is nil")
	}
	//https://open.spotify.com/track/7bZSoBEAEEUsGEuLOf94Jm?si=T1Tdju5qRSmmR0zph_6RBw fuuuuunky
	return func(target string) macaron.Handler {
		return func(c *models.ReqContext) {
			limitReached, err := quotaService.QuotaReached(c, target)
			if err != nil {
				c.JsonApiErr(500, "Failed to get quota", err)
				return
			}
			if limitReached {
				c.JsonApiErr(403, fmt.Sprintf("%s Quota reached", target), nil)
				return
			}
		}
	}
}
