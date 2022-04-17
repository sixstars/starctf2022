//go:build integration
// +build integration

package sqlstore

import (
	"testing"

	"github.com/grafana/grafana/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestPlaylistDataAccess(t *testing.T) {
	InitTestDB(t)

	t.Run("Can create playlist", func(t *testing.T) {
		items := []models.PlaylistItemDTO{
			{Title: "graphite", Value: "graphite", Type: "dashboard_by_tag"},
			{Title: "Backend response times", Value: "3", Type: "dashboard_by_id"},
		}
		cmd := models.CreatePlaylistCommand{Name: "NYC office", Interval: "10m", OrgId: 1, Items: items}
		err := CreatePlaylist(&cmd)
		require.NoError(t, err)

		t.Run("Can update playlist", func(t *testing.T) {
			items := []models.PlaylistItemDTO{
				{Title: "influxdb", Value: "influxdb", Type: "dashboard_by_tag"},
				{Title: "Backend response times", Value: "2", Type: "dashboard_by_id"},
			}
			query := models.UpdatePlaylistCommand{Name: "NYC office ", OrgId: 1, Id: 1, Interval: "10s", Items: items}
			err = UpdatePlaylist(&query)
			require.NoError(t, err)
		})

		t.Run("Can remove playlist", func(t *testing.T) {
			deleteQuery := models.DeletePlaylistCommand{Id: 1, OrgId: 1}
			err = DeletePlaylist(&deleteQuery)
			require.NoError(t, err)

			getQuery := models.GetPlaylistByIdQuery{Id: 1}
			err = GetPlaylist(&getQuery)
			require.NoError(t, err)
			require.Equal(t, int64(0), getQuery.Result.Id, "playlist should've been removed")
		})
	})

	t.Run("Delete playlist that doesn't exist", func(t *testing.T) {
		deleteQuery := models.DeletePlaylistCommand{Id: 1, OrgId: 1}
		err := DeletePlaylist(&deleteQuery)
		require.NoError(t, err)
	})

	t.Run("Delete playlist with invalid command yields error", func(t *testing.T) {
		testCases := []struct {
			desc string
			cmd  models.DeletePlaylistCommand
		}{
			{desc: "none", cmd: models.DeletePlaylistCommand{}},
			{desc: "no OrgId", cmd: models.DeletePlaylistCommand{Id: 1}},
			{desc: "no Id", cmd: models.DeletePlaylistCommand{OrgId: 1}},
		}

		for _, tc := range testCases {
			t.Run(tc.desc, func(t *testing.T) {
				err := DeletePlaylist(&tc.cmd)
				require.EqualError(t, err, models.ErrCommandValidationFailed.Error())
			})
		}
	})
}
