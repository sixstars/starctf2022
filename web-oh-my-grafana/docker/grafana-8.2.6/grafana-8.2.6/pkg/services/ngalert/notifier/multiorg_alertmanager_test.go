package notifier

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/ngalert/metrics"
	"github.com/grafana/grafana/pkg/services/ngalert/models"
	"github.com/grafana/grafana/pkg/setting"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func TestMultiOrgAlertmanager_SyncAlertmanagersForOrgs(t *testing.T) {
	configStore := &FakeConfigStore{
		configs: map[int64]*models.AlertConfiguration{},
	}
	orgStore := &FakeOrgStore{
		orgs: []int64{1, 2, 3},
	}

	tmpDir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	kvStore := newFakeKVStore(t)
	reg := prometheus.NewPedanticRegistry()
	m := metrics.NewNGAlert(reg)
	cfg := &setting.Cfg{
		DataPath: tmpDir,
		UnifiedAlerting: setting.UnifiedAlertingSettings{
			AlertmanagerConfigPollInterval: 3 * time.Minute,
			DefaultConfiguration:           setting.GetAlertmanagerDefaultConfiguration(),
			DisabledOrgs:                   map[int64]struct{}{5: {}},
		}, // do not poll in tests.
	}
	mam, err := NewMultiOrgAlertmanager(cfg, configStore, orgStore, kvStore, m.GetMultiOrgAlertmanagerMetrics(), log.New("testlogger"))
	require.NoError(t, err)
	ctx := context.Background()

	t.Cleanup(cleanOrgDirectories(tmpDir, t))

	// Ensure that one Alertmanager is created per org.
	{
		require.NoError(t, mam.LoadAndSyncAlertmanagersForOrgs(ctx))
		require.Len(t, mam.alertmanagers, 3)
		require.NoError(t, testutil.GatherAndCompare(reg, bytes.NewBufferString(`
# HELP grafana_alerting_active_configurations The number of active Alertmanager configurations.
# TYPE grafana_alerting_active_configurations gauge
grafana_alerting_active_configurations 3
# HELP grafana_alerting_discovered_configurations The number of organizations we've discovered that require an Alertmanager configuration.
# TYPE grafana_alerting_discovered_configurations gauge
grafana_alerting_discovered_configurations 3
`), "grafana_alerting_discovered_configurations", "grafana_alerting_active_configurations"))
	}
	// When an org is removed, it should detect it.
	{
		orgStore.orgs = []int64{1, 3}
		require.NoError(t, mam.LoadAndSyncAlertmanagersForOrgs(ctx))
		require.Len(t, mam.alertmanagers, 2)
		require.NoError(t, testutil.GatherAndCompare(reg, bytes.NewBufferString(`
# HELP grafana_alerting_active_configurations The number of active Alertmanager configurations.
# TYPE grafana_alerting_active_configurations gauge
grafana_alerting_active_configurations 2
# HELP grafana_alerting_discovered_configurations The number of organizations we've discovered that require an Alertmanager configuration.
# TYPE grafana_alerting_discovered_configurations gauge
grafana_alerting_discovered_configurations 2
`), "grafana_alerting_discovered_configurations", "grafana_alerting_active_configurations"))
	}
	// if the org comes back, it should detect it.
	{
		orgStore.orgs = []int64{1, 2, 3, 4}
		require.NoError(t, mam.LoadAndSyncAlertmanagersForOrgs(ctx))
		require.Len(t, mam.alertmanagers, 4)
		require.NoError(t, testutil.GatherAndCompare(reg, bytes.NewBufferString(`
# HELP grafana_alerting_active_configurations The number of active Alertmanager configurations.
# TYPE grafana_alerting_active_configurations gauge
grafana_alerting_active_configurations 4
# HELP grafana_alerting_discovered_configurations The number of organizations we've discovered that require an Alertmanager configuration.
# TYPE grafana_alerting_discovered_configurations gauge
grafana_alerting_discovered_configurations 4
`), "grafana_alerting_discovered_configurations", "grafana_alerting_active_configurations"))
	}
	// if the disabled org comes back, it should not detect it.
	{
		orgStore.orgs = []int64{1, 2, 3, 4, 5}
		require.NoError(t, mam.LoadAndSyncAlertmanagersForOrgs(ctx))
		require.Len(t, mam.alertmanagers, 4)
	}

	// Orphaned state should be removed.
	{
		orgID := int64(6)
		// First we create a directory and two files for an ograniztation that
		// is not existing in the current state.
		orphanDir := filepath.Join(tmpDir, "alerting", "6")
		err := os.Mkdir(orphanDir, 0750)
		require.NoError(t, err)

		silencesPath := filepath.Join(orphanDir, silencesFilename)
		err = os.WriteFile(silencesPath, []byte("file_1"), 0644)
		require.NoError(t, err)

		notificationPath := filepath.Join(orphanDir, notificationLogFilename)
		err = os.WriteFile(notificationPath, []byte("file_2"), 0644)
		require.NoError(t, err)

		// We make sure that both files are on disk.
		info, err := os.Stat(silencesPath)
		require.NoError(t, err)
		require.Equal(t, info.Name(), silencesFilename)
		info, err = os.Stat(notificationPath)
		require.NoError(t, err)
		require.Equal(t, info.Name(), notificationLogFilename)

		// We also populate the kvstore with orphaned records.
		err = kvStore.Set(ctx, orgID, KVNamespace, silencesFilename, "file_1")
		require.NoError(t, err)

		err = kvStore.Set(ctx, orgID, KVNamespace, notificationLogFilename, "file_1")
		require.NoError(t, err)

		// Now re run the sync job once.
		require.NoError(t, mam.LoadAndSyncAlertmanagersForOrgs(ctx))

		// The organization directory should be gone by now.
		_, err = os.Stat(orphanDir)
		require.True(t, errors.Is(err, fs.ErrNotExist))

		// The organization kvstore records should be gone by now.
		_, exists, _ := kvStore.Get(ctx, orgID, KVNamespace, silencesFilename)
		require.False(t, exists)

		_, exists, _ = kvStore.Get(ctx, orgID, KVNamespace, notificationLogFilename)
		require.False(t, exists)
	}
}

func TestMultiOrgAlertmanager_SyncAlertmanagersForOrgsWithFailures(t *testing.T) {
	// Include a broken configuration for organization 2.
	configStore := &FakeConfigStore{
		configs: map[int64]*models.AlertConfiguration{
			2: {AlertmanagerConfiguration: brokenConfig, OrgID: 2},
		},
	}
	orgStore := &FakeOrgStore{
		orgs: []int64{1, 2, 3},
	}

	tmpDir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	kvStore := newFakeKVStore(t)
	reg := prometheus.NewPedanticRegistry()
	m := metrics.NewNGAlert(reg)
	cfg := &setting.Cfg{
		DataPath: tmpDir,
		UnifiedAlerting: setting.UnifiedAlertingSettings{
			AlertmanagerConfigPollInterval: 10 * time.Minute,
			DefaultConfiguration:           setting.GetAlertmanagerDefaultConfiguration(),
		}, // do not poll in tests.
	}
	mam, err := NewMultiOrgAlertmanager(cfg, configStore, orgStore, kvStore, m.GetMultiOrgAlertmanagerMetrics(), log.New("testlogger"))
	require.NoError(t, err)
	ctx := context.Background()

	// When you sync the first time, the alertmanager is created but is doesn't become ready until you have a configuration applied.
	{
		require.NoError(t, mam.LoadAndSyncAlertmanagersForOrgs(ctx))
		require.Len(t, mam.alertmanagers, 3)
		require.True(t, mam.alertmanagers[1].ready())
		require.False(t, mam.alertmanagers[2].ready())
		require.True(t, mam.alertmanagers[3].ready())
	}

	// On the next sync, it never panics and alertmanager is still not ready.
	{
		require.NoError(t, mam.LoadAndSyncAlertmanagersForOrgs(ctx))
		require.Len(t, mam.alertmanagers, 3)
		require.True(t, mam.alertmanagers[1].ready())
		require.False(t, mam.alertmanagers[2].ready())
		require.True(t, mam.alertmanagers[3].ready())
	}

	// If we fix the configuration, it becomes ready.
	{
		configStore.configs = map[int64]*models.AlertConfiguration{} // It'll apply the default config.
		require.NoError(t, mam.LoadAndSyncAlertmanagersForOrgs(ctx))
		require.Len(t, mam.alertmanagers, 3)
		require.True(t, mam.alertmanagers[1].ready())
		require.True(t, mam.alertmanagers[2].ready())
		require.True(t, mam.alertmanagers[3].ready())
	}
}

func TestMultiOrgAlertmanager_AlertmanagerFor(t *testing.T) {
	configStore := &FakeConfigStore{
		configs: map[int64]*models.AlertConfiguration{},
	}
	orgStore := &FakeOrgStore{
		orgs: []int64{1, 2, 3},
	}
	tmpDir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	cfg := &setting.Cfg{
		DataPath:        tmpDir,
		UnifiedAlerting: setting.UnifiedAlertingSettings{AlertmanagerConfigPollInterval: 3 * time.Minute, DefaultConfiguration: setting.GetAlertmanagerDefaultConfiguration()}, // do not poll in tests.
	}
	kvStore := newFakeKVStore(t)
	reg := prometheus.NewPedanticRegistry()
	m := metrics.NewNGAlert(reg)
	mam, err := NewMultiOrgAlertmanager(cfg, configStore, orgStore, kvStore, m.GetMultiOrgAlertmanagerMetrics(), log.New("testlogger"))
	require.NoError(t, err)
	ctx := context.Background()

	t.Cleanup(cleanOrgDirectories(tmpDir, t))

	// Ensure that one Alertmanagers is created per org.
	{
		require.NoError(t, mam.LoadAndSyncAlertmanagersForOrgs(ctx))
		require.Len(t, mam.alertmanagers, 3)
	}

	// First, let's try to request an Alertmanager from an org that doesn't exist.
	{
		_, err := mam.AlertmanagerFor(5)
		require.EqualError(t, err, ErrNoAlertmanagerForOrg.Error())
	}

	// Now, let's try to request an Alertmanager that is not ready.
	{
		// let's delete its "running config" to make it non-ready
		mam.alertmanagers[1].config = nil
		_, err := mam.AlertmanagerFor(1)
		require.EqualError(t, err, ErrAlertmanagerNotReady.Error())
	}

	// With an Alertmanager that exists, it responds correctly.
	{
		am, err := mam.AlertmanagerFor(2)
		require.NoError(t, err)
		require.Equal(t, *am.GetStatus().VersionInfo.Version, "N/A")
		require.Equal(t, am.orgID, int64(2))
		require.NotNil(t, am.config)
	}

	// Let's now remove the previous queried organization.
	orgStore.orgs = []int64{1, 3}
	require.NoError(t, mam.LoadAndSyncAlertmanagersForOrgs(ctx))
	{
		_, err := mam.AlertmanagerFor(2)
		require.EqualError(t, err, ErrNoAlertmanagerForOrg.Error())
	}
}

// nolint:unused
func cleanOrgDirectories(path string, t *testing.T) func() {
	return func() {
		require.NoError(t, os.RemoveAll(path))
	}
}

var brokenConfig = `
	"alertmanager_config": {
		"route": {
			"receiver": "grafana-default-email"
		},
		"receivers": [{
			"name": "grafana-default-email",
			"grafana_managed_receiver_configs": [{
				"uid": "",
				"name": "slack receiver",
				"type": "slack",
				"isDefault": true,
				"settings": {
					"addresses": "<example@email.com>"
					"url": "�r_��q/b�����p@ⱎȏ =��@ӹtd>Rú�H��           �;�@Uf��0�\k2*jh�}Íu�)"2�F6]�}r��R�b�d�J;��S퓧��$��",
					"recipient": "#graphana-metrics",
				}
			}]
		}]
	}
}`
