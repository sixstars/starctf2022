package azuremonitor

import (
	"testing"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/tsdb/azuremonitor/azcredentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentials_getAuthType(t *testing.T) {
	cfg := &setting.Cfg{}

	t.Run("when managed identities enabled", func(t *testing.T) {
		cfg.Azure.ManagedIdentityEnabled = true

		t.Run("should be client secret if auth type is set to client secret", func(t *testing.T) {
			jsonData := simplejson.NewFromAny(map[string]interface{}{
				"azureAuthType": azcredentials.AzureAuthClientSecret,
			})

			authType := getAuthType(cfg, jsonData)

			assert.Equal(t, azcredentials.AzureAuthClientSecret, authType)
		})

		t.Run("should be managed identity if datasource not configured", func(t *testing.T) {
			jsonData := simplejson.NewFromAny(map[string]interface{}{
				"azureAuthType": "",
			})

			authType := getAuthType(cfg, jsonData)

			assert.Equal(t, azcredentials.AzureAuthManagedIdentity, authType)
		})

		t.Run("should be client secret if auth type not specified but credentials configured", func(t *testing.T) {
			jsonData := simplejson.NewFromAny(map[string]interface{}{
				"azureAuthType": "",
				"tenantId":      "9b9d90ee-a5cc-49c2-b97e-0d1b0f086b5c",
				"clientId":      "849ccbb0-92eb-4226-b228-ef391abd8fe6",
			})

			authType := getAuthType(cfg, jsonData)

			assert.Equal(t, azcredentials.AzureAuthClientSecret, authType)
		})
	})

	t.Run("when managed identities disabled", func(t *testing.T) {
		cfg.Azure.ManagedIdentityEnabled = false

		t.Run("should be managed identity if auth type is set to managed identity", func(t *testing.T) {
			jsonData := simplejson.NewFromAny(map[string]interface{}{
				"azureAuthType": azcredentials.AzureAuthManagedIdentity,
			})

			authType := getAuthType(cfg, jsonData)

			assert.Equal(t, azcredentials.AzureAuthManagedIdentity, authType)
		})

		t.Run("should be client secret if datasource not configured", func(t *testing.T) {
			jsonData := simplejson.NewFromAny(map[string]interface{}{
				"azureAuthType": "",
			})

			authType := getAuthType(cfg, jsonData)

			assert.Equal(t, azcredentials.AzureAuthClientSecret, authType)
		})
	})
}

func TestCredentials_getAzureCloud(t *testing.T) {
	cfg := &setting.Cfg{
		Azure: setting.AzureSettings{
			Cloud: setting.AzureChina,
		},
	}

	t.Run("when auth type is managed identity", func(t *testing.T) {
		jsonData := simplejson.NewFromAny(map[string]interface{}{
			"azureAuthType": azcredentials.AzureAuthManagedIdentity,
			"cloudName":     azureMonitorGermany,
		})

		t.Run("should be from server configuration regardless of datasource value", func(t *testing.T) {
			cloud, err := getAzureCloud(cfg, jsonData)
			require.NoError(t, err)

			assert.Equal(t, setting.AzureChina, cloud)
		})

		t.Run("should be public if not set in server configuration", func(t *testing.T) {
			cfg := &setting.Cfg{
				Azure: setting.AzureSettings{
					Cloud: "",
				},
			}

			cloud, err := getAzureCloud(cfg, jsonData)
			require.NoError(t, err)

			assert.Equal(t, setting.AzurePublic, cloud)
		})
	})

	t.Run("when auth type is client secret", func(t *testing.T) {
		t.Run("should be from datasource value normalized to known cloud name", func(t *testing.T) {
			jsonData := simplejson.NewFromAny(map[string]interface{}{
				"azureAuthType": azcredentials.AzureAuthClientSecret,
				"cloudName":     azureMonitorGermany,
			})

			cloud, err := getAzureCloud(cfg, jsonData)
			require.NoError(t, err)

			assert.Equal(t, setting.AzureGermany, cloud)
		})

		t.Run("should be from server configuration if not set in datasource", func(t *testing.T) {
			jsonData := simplejson.NewFromAny(map[string]interface{}{
				"azureAuthType": azcredentials.AzureAuthClientSecret,
				"cloudName":     "",
			})

			cloud, err := getAzureCloud(cfg, jsonData)
			require.NoError(t, err)

			assert.Equal(t, setting.AzureChina, cloud)
		})
	})
}

func TestCredentials_getAzureCredentials(t *testing.T) {
	cfg := &setting.Cfg{
		Azure: setting.AzureSettings{
			Cloud: setting.AzureChina,
		},
	}

	secureJsonData := map[string]string{
		"clientSecret": "59e3498f-eb12-4943-b8f0-a5aa42640058",
	}

	t.Run("when auth type is managed identity", func(t *testing.T) {
		jsonData := simplejson.NewFromAny(map[string]interface{}{
			"azureAuthType": azcredentials.AzureAuthManagedIdentity,
			"cloudName":     azureMonitorGermany,
			"tenantId":      "9b9d90ee-a5cc-49c2-b97e-0d1b0f086b5c",
			"clientId":      "849ccbb0-92eb-4226-b228-ef391abd8fe6",
		})

		t.Run("should return managed identity credentials", func(t *testing.T) {
			credentials, err := getAzureCredentials(cfg, jsonData, secureJsonData)
			require.NoError(t, err)
			require.IsType(t, &azcredentials.AzureManagedIdentityCredentials{}, credentials)
			msiCredentials := credentials.(*azcredentials.AzureManagedIdentityCredentials)

			// Azure Monitor datasource doesn't support user-assigned managed identities (ClientId is always empty)
			assert.Equal(t, "", msiCredentials.ClientId)
		})
	})

	t.Run("when auth type is client secret", func(t *testing.T) {
		jsonData := simplejson.NewFromAny(map[string]interface{}{
			"azureAuthType": azcredentials.AzureAuthClientSecret,
			"cloudName":     azureMonitorGermany,
			"tenantId":      "9b9d90ee-a5cc-49c2-b97e-0d1b0f086b5c",
			"clientId":      "849ccbb0-92eb-4226-b228-ef391abd8fe6",
		})

		t.Run("should return client secret credentials", func(t *testing.T) {
			cfg := &setting.Cfg{
				Azure: setting.AzureSettings{
					Cloud: setting.AzureChina,
				},
			}

			credentials, err := getAzureCredentials(cfg, jsonData, secureJsonData)
			require.NoError(t, err)
			require.IsType(t, &azcredentials.AzureClientSecretCredentials{}, credentials)
			clientSecretCredentials := credentials.(*azcredentials.AzureClientSecretCredentials)

			assert.Equal(t, setting.AzureGermany, clientSecretCredentials.AzureCloud)
			assert.Equal(t, "9b9d90ee-a5cc-49c2-b97e-0d1b0f086b5c", clientSecretCredentials.TenantId)
			assert.Equal(t, "849ccbb0-92eb-4226-b228-ef391abd8fe6", clientSecretCredentials.ClientId)
			assert.Equal(t, "59e3498f-eb12-4943-b8f0-a5aa42640058", clientSecretCredentials.ClientSecret)

			// Azure Monitor datasource doesn't support custom IdP authorities (Authority is always empty)
			assert.Equal(t, "", clientSecretCredentials.Authority)
		})
	})
}
