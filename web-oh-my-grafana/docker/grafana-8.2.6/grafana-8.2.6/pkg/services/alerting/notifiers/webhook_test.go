package notifiers

import (
	"testing"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookNotifier_parsingFromSettings(t *testing.T) {
	t.Run("Empty settings should cause error", func(t *testing.T) {
		const json = `{}`

		settingsJSON, err := simplejson.NewJson([]byte(json))
		require.NoError(t, err)
		model := &models.AlertNotification{
			Name:     "ops",
			Type:     "webhook",
			Settings: settingsJSON,
		}

		_, err = NewWebHookNotifier(model)
		require.Error(t, err)
	})

	t.Run("Valid settings should result in a valid notifier", func(t *testing.T) {
		const json = `{"url": "http://google.com"}`

		settingsJSON, err := simplejson.NewJson([]byte(json))
		require.NoError(t, err)
		model := &models.AlertNotification{
			Name:     "ops",
			Type:     "webhook",
			Settings: settingsJSON,
		}

		not, err := NewWebHookNotifier(model)
		require.NoError(t, err)
		webhookNotifier := not.(*WebhookNotifier)

		assert.Equal(t, "ops", webhookNotifier.Name)
		assert.Equal(t, "webhook", webhookNotifier.Type)
		assert.Equal(t, "http://google.com", webhookNotifier.URL)
	})
}
