//go:build integration
// +build integration

package sqlstore

import (
	"context"
	"fmt"
	"testing"

	"github.com/grafana/grafana/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsDataAccess(t *testing.T) {
	sqlStore := InitTestDB(t)
	populateDB(t, sqlStore)

	t.Run("Get system stats should not results in error", func(t *testing.T) {
		query := models.GetSystemStatsQuery{}
		err := GetSystemStats(&query)
		require.NoError(t, err)
		assert.Equal(t, int64(3), query.Result.Users)
		assert.Equal(t, int64(0), query.Result.Editors)
		assert.Equal(t, int64(0), query.Result.Viewers)
		assert.Equal(t, int64(3), query.Result.Admins)
		assert.Equal(t, int64(0), query.Result.LibraryPanels)
		assert.Equal(t, int64(0), query.Result.LibraryVariables)
	})

	t.Run("Get system user count stats should not results in error", func(t *testing.T) {
		query := models.GetSystemUserCountStatsQuery{}
		err := GetSystemUserCountStats(context.Background(), &query)
		assert.NoError(t, err)
	})

	t.Run("Get datasource stats should not results in error", func(t *testing.T) {
		query := models.GetDataSourceStatsQuery{}
		err := GetDataSourceStats(&query)
		assert.NoError(t, err)
	})

	t.Run("Get datasource access stats should not results in error", func(t *testing.T) {
		query := models.GetDataSourceAccessStatsQuery{}
		err := GetDataSourceAccessStats(&query)
		assert.NoError(t, err)
	})

	t.Run("Get alert notifier stats should not results in error", func(t *testing.T) {
		query := models.GetAlertNotifierUsageStatsQuery{}
		err := GetAlertNotifiersUsageStats(context.Background(), &query)
		assert.NoError(t, err)
	})

	t.Run("Get admin stats should not result in error", func(t *testing.T) {
		query := models.GetAdminStatsQuery{}
		err := GetAdminStats(&query)
		assert.NoError(t, err)
	})
}

func populateDB(t *testing.T, sqlStore *SQLStore) {
	t.Helper()

	users := make([]models.User, 3)
	for i := range users {
		cmd := models.CreateUserCommand{
			Email:   fmt.Sprintf("usertest%v@test.com", i),
			Name:    fmt.Sprintf("user name %v", i),
			Login:   fmt.Sprintf("user_test_%v_login", i),
			OrgName: fmt.Sprintf("Org #%v", i),
		}
		user, err := sqlStore.CreateUser(context.Background(), cmd)
		require.NoError(t, err)
		users[i] = *user
	}

	// get 1st user's organisation
	getOrgByIdQuery := &models.GetOrgByIdQuery{Id: users[0].OrgId}
	err := GetOrgById(getOrgByIdQuery)
	require.NoError(t, err)
	org := getOrgByIdQuery.Result

	// add 2nd user as editor
	cmd := &models.AddOrgUserCommand{
		OrgId:  org.Id,
		UserId: users[1].Id,
		Role:   models.ROLE_EDITOR,
	}
	err = AddOrgUser(cmd)
	require.NoError(t, err)

	// add 3rd user as viewer
	cmd = &models.AddOrgUserCommand{
		OrgId:  org.Id,
		UserId: users[2].Id,
		Role:   models.ROLE_VIEWER,
	}
	err = AddOrgUser(cmd)
	require.NoError(t, err)

	// get 2nd user's organisation
	getOrgByIdQuery = &models.GetOrgByIdQuery{Id: users[1].OrgId}
	err = GetOrgById(getOrgByIdQuery)
	require.NoError(t, err)
	org = getOrgByIdQuery.Result

	// add 1st user as admin
	cmd = &models.AddOrgUserCommand{
		OrgId:  org.Id,
		UserId: users[0].Id,
		Role:   models.ROLE_ADMIN,
	}
	err = AddOrgUser(cmd)
	require.NoError(t, err)

	// update 1st user last seen at
	updateUserLastSeenAtCmd := &models.UpdateUserLastSeenAtCommand{
		UserId: users[0].Id,
	}
	err = UpdateUserLastSeenAt(updateUserLastSeenAtCmd)
	require.NoError(t, err)

	// force renewal of user stats
	err = updateUserRoleCountsIfNecessary(context.Background(), true)
	require.NoError(t, err)
}
