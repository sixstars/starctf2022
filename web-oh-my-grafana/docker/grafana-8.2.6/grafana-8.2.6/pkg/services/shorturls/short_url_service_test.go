package shorturls

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/stretchr/testify/require"
)

func TestShortURLService(t *testing.T) {
	user := &models.SignedInUser{UserId: 1}
	sqlStore := sqlstore.InitTestDB(t)

	t.Run("User can create and read short URLs", func(t *testing.T) {
		const refPath = "mock/path?test=true"

		service := ShortURLService{SQLStore: sqlStore}

		newShortURL, err := service.CreateShortURL(context.Background(), user, refPath)
		require.NoError(t, err)
		require.NotNil(t, newShortURL)
		require.NotEmpty(t, newShortURL.Uid)

		existingShortURL, err := service.GetShortURLByUID(context.Background(), user, newShortURL.Uid)
		require.NoError(t, err)
		require.NotNil(t, existingShortURL)
		require.Equal(t, refPath, existingShortURL.Path)

		t.Run("and update last seen at", func(t *testing.T) {
			origGetTime := getTime
			t.Cleanup(func() {
				getTime = origGetTime
			})

			expectedTime := time.Date(2020, time.November, 27, 6, 5, 1, 0, time.UTC)
			getTime = func() time.Time {
				return expectedTime
			}

			err := service.UpdateLastSeenAt(context.Background(), existingShortURL)
			require.NoError(t, err)

			updatedShortURL, err := service.GetShortURLByUID(context.Background(), user, existingShortURL.Uid)
			require.NoError(t, err)
			require.Equal(t, expectedTime.Unix(), updatedShortURL.LastSeenAt)
		})

		t.Run("and stale short urls can be deleted", func(t *testing.T) {
			staleShortURL, err := service.CreateShortURL(context.Background(), user, refPath)
			require.NoError(t, err)
			require.NotNil(t, staleShortURL)
			require.NotEmpty(t, staleShortURL.Uid)
			require.Equal(t, int64(0), staleShortURL.LastSeenAt)

			cmd := models.DeleteShortUrlCommand{OlderThan: time.Unix(staleShortURL.CreatedAt, 0)}
			err = service.DeleteStaleShortURLs(context.Background(), &cmd)
			require.NoError(t, err)
			require.Equal(t, int64(1), cmd.NumDeleted)

			t.Run("and previously accessed short urls will still exist", func(t *testing.T) {
				updatedShortURL, err := service.GetShortURLByUID(context.Background(), user, existingShortURL.Uid)
				require.NoError(t, err)
				require.NotNil(t, updatedShortURL)
			})

			t.Run("and no action when no stale short urls exist", func(t *testing.T) {
				cmd := models.DeleteShortUrlCommand{OlderThan: time.Unix(existingShortURL.CreatedAt, 0)}
				require.NoError(t, err)
				require.Equal(t, int64(0), cmd.NumDeleted)
			})
		})
	})

	t.Run("User cannot look up nonexistent short URLs", func(t *testing.T) {
		service := ShortURLService{SQLStore: sqlStore}

		shortURL, err := service.GetShortURLByUID(context.Background(), user, "testnotfounduid")
		require.Error(t, err)
		require.Equal(t, models.ErrShortURLNotFound, err)
		require.Nil(t, shortURL)
	})
}
