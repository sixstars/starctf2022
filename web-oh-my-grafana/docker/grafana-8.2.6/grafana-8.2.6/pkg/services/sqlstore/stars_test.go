//go:build integration
// +build integration

package sqlstore

import (
	"context"
	"testing"

	"github.com/grafana/grafana/pkg/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUserStarsDataAccess(t *testing.T) {
	Convey("Testing User Stars Data Access", t, func() {
		InitTestDB(t)

		Convey("Given saved star", func() {
			cmd := models.StarDashboardCommand{
				DashboardId: 10,
				UserId:      12,
			}

			err := StarDashboard(&cmd)
			So(err, ShouldBeNil)

			Convey("IsStarredByUser should return true when starred", func() {
				query := models.IsStarredByUserQuery{UserId: 12, DashboardId: 10}
				err := IsStarredByUserCtx(context.Background(), &query)
				So(err, ShouldBeNil)

				So(query.Result, ShouldBeTrue)
			})

			Convey("IsStarredByUser should return false when not starred", func() {
				query := models.IsStarredByUserQuery{UserId: 12, DashboardId: 12}
				err := IsStarredByUserCtx(context.Background(), &query)
				So(err, ShouldBeNil)

				So(query.Result, ShouldBeFalse)
			})
		})
	})
}
