package sqlstore

import (
	"context"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
)

func init() {
	bus.AddHandler("sql", StarDashboard)
	bus.AddHandler("sql", UnstarDashboard)
	bus.AddHandler("sql", GetUserStars)
	bus.AddHandlerCtx("sql", IsStarredByUserCtx)
}

func IsStarredByUserCtx(ctx context.Context, query *models.IsStarredByUserQuery) error {
	return withDbSession(ctx, x, func(dbSession *DBSession) error {
		rawSQL := "SELECT 1 from star where user_id=? and dashboard_id=?"
		results, err := dbSession.Query(rawSQL, query.UserId, query.DashboardId)

		if err != nil {
			return err
		}

		if len(results) == 0 {
			return nil
		}

		query.Result = true

		return nil
	})
}

func StarDashboard(cmd *models.StarDashboardCommand) error {
	if cmd.DashboardId == 0 || cmd.UserId == 0 {
		return models.ErrCommandValidationFailed
	}

	return inTransaction(func(sess *DBSession) error {
		entity := models.Star{
			UserId:      cmd.UserId,
			DashboardId: cmd.DashboardId,
		}

		_, err := sess.Insert(&entity)
		return err
	})
}

func UnstarDashboard(cmd *models.UnstarDashboardCommand) error {
	if cmd.DashboardId == 0 || cmd.UserId == 0 {
		return models.ErrCommandValidationFailed
	}

	return inTransaction(func(sess *DBSession) error {
		var rawSQL = "DELETE FROM star WHERE user_id=? and dashboard_id=?"
		_, err := sess.Exec(rawSQL, cmd.UserId, cmd.DashboardId)
		return err
	})
}

func GetUserStars(query *models.GetUserStarsQuery) error {
	var stars = make([]models.Star, 0)
	err := x.Where("user_id=?", query.UserId).Find(&stars)

	query.Result = make(map[int64]bool)
	for _, star := range stars {
		query.Result[star.DashboardId] = true
	}

	return err
}
