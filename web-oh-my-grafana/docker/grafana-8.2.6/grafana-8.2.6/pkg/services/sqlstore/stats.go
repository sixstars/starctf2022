package sqlstore

import (
	"context"
	"strconv"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
)

func init() {
	bus.AddHandler("sql", GetSystemStats)
	bus.AddHandler("sql", GetDataSourceStats)
	bus.AddHandler("sql", GetDataSourceAccessStats)
	bus.AddHandler("sql", GetAdminStats)
	bus.AddHandlerCtx("sql", GetAlertNotifiersUsageStats)
	bus.AddHandlerCtx("sql", GetSystemUserCountStats)
}

const activeUserTimeLimit = time.Hour * 24 * 30
const dailyActiveUserTimeLimit = time.Hour * 24

func GetAlertNotifiersUsageStats(ctx context.Context, query *models.GetAlertNotifierUsageStatsQuery) error {
	var rawSQL = `SELECT COUNT(*) AS count, type FROM ` + dialect.Quote("alert_notification") + ` GROUP BY type`
	query.Result = make([]*models.NotifierUsageStats, 0)
	err := x.SQL(rawSQL).Find(&query.Result)
	return err
}

func GetDataSourceStats(query *models.GetDataSourceStatsQuery) error {
	var rawSQL = `SELECT COUNT(*) AS count, type FROM ` + dialect.Quote("data_source") + ` GROUP BY type`
	query.Result = make([]*models.DataSourceStats, 0)
	err := x.SQL(rawSQL).Find(&query.Result)
	return err
}

func GetDataSourceAccessStats(query *models.GetDataSourceAccessStatsQuery) error {
	var rawSQL = `SELECT COUNT(*) AS count, type, access FROM ` + dialect.Quote("data_source") + ` GROUP BY type, access`
	query.Result = make([]*models.DataSourceAccessStats, 0)
	err := x.SQL(rawSQL).Find(&query.Result)
	return err
}

func GetSystemStats(query *models.GetSystemStatsQuery) error {
	sb := &SQLBuilder{}
	sb.Write("SELECT ")
	sb.Write(`(SELECT COUNT(*) FROM ` + dialect.Quote("user") + `) AS users,`)
	sb.Write(`(SELECT COUNT(*) FROM ` + dialect.Quote("org") + `) AS orgs,`)
	sb.Write(`(SELECT COUNT(*) FROM ` + dialect.Quote("data_source") + `) AS datasources,`)
	sb.Write(`(SELECT COUNT(*) FROM ` + dialect.Quote("star") + `) AS stars,`)
	sb.Write(`(SELECT COUNT(*) FROM ` + dialect.Quote("playlist") + `) AS playlists,`)
	sb.Write(`(SELECT COUNT(*) FROM ` + dialect.Quote("alert") + `) AS alerts,`)

	activeUserDeadlineDate := time.Now().Add(-activeUserTimeLimit)
	sb.Write(`(SELECT COUNT(*) FROM `+dialect.Quote("user")+` WHERE last_seen_at > ?) AS active_users,`, activeUserDeadlineDate)

	dailyActiveUserDeadlineDate := time.Now().Add(-dailyActiveUserTimeLimit)
	sb.Write(`(SELECT COUNT(*) FROM `+dialect.Quote("user")+` WHERE last_seen_at > ?) AS daily_active_users,`, dailyActiveUserDeadlineDate)

	sb.Write(`(SELECT COUNT(id) FROM `+dialect.Quote("dashboard")+` WHERE is_folder = ?) AS dashboards,`, dialect.BooleanStr(false))
	sb.Write(`(SELECT COUNT(id) FROM `+dialect.Quote("dashboard")+` WHERE is_folder = ?) AS folders,`, dialect.BooleanStr(true))

	sb.Write(`(
		SELECT COUNT(acl.id)
		FROM `+dialect.Quote("dashboard_acl")+` AS acl
			INNER JOIN `+dialect.Quote("dashboard")+` AS d
			ON d.id = acl.dashboard_id
		WHERE d.is_folder = ?
	) AS dashboard_permissions,`, dialect.BooleanStr(false))

	sb.Write(`(
		SELECT COUNT(acl.id)
		FROM `+dialect.Quote("dashboard_acl")+` AS acl
			INNER JOIN `+dialect.Quote("dashboard")+` AS d
			ON d.id = acl.dashboard_id
		WHERE d.is_folder = ?
	) AS folder_permissions,`, dialect.BooleanStr(true))

	sb.Write(viewersPermissionsCounterSQL("dashboards_viewers_can_edit", false, models.PERMISSION_EDIT))
	sb.Write(viewersPermissionsCounterSQL("dashboards_viewers_can_admin", false, models.PERMISSION_ADMIN))
	sb.Write(viewersPermissionsCounterSQL("folders_viewers_can_edit", true, models.PERMISSION_EDIT))
	sb.Write(viewersPermissionsCounterSQL("folders_viewers_can_admin", true, models.PERMISSION_ADMIN))

	sb.Write(`(SELECT COUNT(id) FROM ` + dialect.Quote("dashboard_provisioning") + `) AS provisioned_dashboards,`)
	sb.Write(`(SELECT COUNT(id) FROM ` + dialect.Quote("dashboard_snapshot") + `) AS snapshots,`)
	sb.Write(`(SELECT COUNT(id) FROM ` + dialect.Quote("dashboard_version") + `) AS dashboard_versions,`)
	sb.Write(`(SELECT COUNT(id) FROM ` + dialect.Quote("annotation") + `) AS annotations,`)
	sb.Write(`(SELECT COUNT(id) FROM ` + dialect.Quote("team") + `) AS teams,`)
	sb.Write(`(SELECT COUNT(id) FROM ` + dialect.Quote("user_auth_token") + `) AS auth_tokens,`)
	sb.Write(`(SELECT COUNT(id) FROM ` + dialect.Quote("alert_rule") + `) AS alert_rules,`)
	sb.Write(`(SELECT COUNT(id) FROM `+dialect.Quote("library_element")+` WHERE kind = ?) AS library_panels,`, models.PanelElement)
	sb.Write(`(SELECT COUNT(id) FROM `+dialect.Quote("library_element")+` WHERE kind = ?) AS library_variables,`, models.VariableElement)

	sb.Write(roleCounterSQL())

	var stats models.SystemStats
	_, err := x.SQL(sb.GetSQLString(), sb.params...).Get(&stats)
	if err != nil {
		return err
	}

	query.Result = &stats

	return nil
}

func roleCounterSQL() string {
	const roleCounterTimeout = 20 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), roleCounterTimeout)
	defer cancel()
	_ = updateUserRoleCountsIfNecessary(ctx, false)
	sqlQuery :=
		strconv.FormatInt(userStatsCache.total.Admins, 10) + ` AS admins, ` +
			strconv.FormatInt(userStatsCache.total.Editors, 10) + ` AS editors, ` +
			strconv.FormatInt(userStatsCache.total.Viewers, 10) + ` AS viewers, ` +
			strconv.FormatInt(userStatsCache.active.Admins, 10) + ` AS active_admins, ` +
			strconv.FormatInt(userStatsCache.active.Editors, 10) + ` AS active_editors, ` +
			strconv.FormatInt(userStatsCache.active.Viewers, 10) + ` AS active_viewers, ` +
			strconv.FormatInt(userStatsCache.dailyActive.Admins, 10) + ` AS daily_active_admins, ` +
			strconv.FormatInt(userStatsCache.dailyActive.Editors, 10) + ` AS daily_active_editors, ` +
			strconv.FormatInt(userStatsCache.dailyActive.Viewers, 10) + ` AS daily_active_viewers`

	return sqlQuery
}

func viewersPermissionsCounterSQL(statName string, isFolder bool, permission models.PermissionType) string {
	return `(
		SELECT COUNT(*)
		FROM ` + dialect.Quote("dashboard_acl") + ` AS acl
			INNER JOIN ` + dialect.Quote("dashboard") + ` AS d
			ON d.id = acl.dashboard_id
		WHERE acl.role = '` + string(models.ROLE_VIEWER) + `'
			AND d.is_folder = ` + dialect.BooleanStr(isFolder) + `
			AND acl.permission = ` + strconv.FormatInt(int64(permission), 10) + `
	) AS ` + statName + `, `
}

func GetAdminStats(query *models.GetAdminStatsQuery) error {
	activeEndDate := time.Now().Add(-activeUserTimeLimit)
	dailyActiveEndDate := time.Now().Add(-dailyActiveUserTimeLimit)

	var rawSQL = `SELECT
		(
			SELECT COUNT(*)
			FROM ` + dialect.Quote("org") + `
		) AS orgs,
		(
			SELECT COUNT(*)
			FROM ` + dialect.Quote("dashboard") + `WHERE is_folder=` + dialect.BooleanStr(false) + `
		) AS dashboards,
		(
			SELECT COUNT(*)
			FROM ` + dialect.Quote("dashboard_snapshot") + `
		) AS snapshots,
		(
			SELECT COUNT( DISTINCT ( ` + dialect.Quote("term") + ` ))
			FROM ` + dialect.Quote("dashboard_tag") + `
		) AS tags,
		(
			SELECT COUNT(*)
			FROM ` + dialect.Quote("data_source") + `
		) AS datasources,
		(
			SELECT COUNT(*)
			FROM ` + dialect.Quote("playlist") + `
		) AS playlists,
		(
			SELECT COUNT(*)
			FROM ` + dialect.Quote("star") + `
		) AS stars,
		(
			SELECT COUNT(*)
			FROM ` + dialect.Quote("alert") + `
		) AS alerts,
		(
			SELECT COUNT(*)
			FROM ` + dialect.Quote("user") + `
		) AS users,
		(
			SELECT COUNT(*)
			FROM ` + dialect.Quote("user") + ` WHERE last_seen_at > ?
		) AS active_users,
		(
			SELECT COUNT(*)
			FROM ` + dialect.Quote("user") + ` WHERE last_seen_at > ?
		) AS daily_active_users,
		` + roleCounterSQL() + `,
		(
			SELECT COUNT(*)
			FROM ` + dialect.Quote("user_auth_token") + ` WHERE rotated_at > ?
		) AS active_sessions,
		(
			SELECT COUNT(*)
			FROM ` + dialect.Quote("user_auth_token") + ` WHERE rotated_at > ?
		) AS daily_active_sessions`

	var stats models.AdminStats
	_, err := x.SQL(rawSQL, activeEndDate, dailyActiveEndDate, activeEndDate.Unix(), dailyActiveEndDate.Unix()).Get(&stats)
	if err != nil {
		return err
	}

	query.Result = &stats
	return nil
}

func GetSystemUserCountStats(ctx context.Context, query *models.GetSystemUserCountStatsQuery) error {
	return withDbSession(ctx, x, func(sess *DBSession) error {
		var rawSQL = `SELECT COUNT(id) AS Count FROM ` + dialect.Quote("user")
		var stats models.SystemUserCountStats
		_, err := sess.SQL(rawSQL).Get(&stats)
		if err != nil {
			return err
		}

		query.Result = &stats

		return nil
	})
}

func updateUserRoleCountsIfNecessary(ctx context.Context, forced bool) error {
	memoizationPeriod := time.Now().Add(-userStatsCacheLimetime)
	if forced || userStatsCache.memoized.Before(memoizationPeriod) {
		err := updateUserRoleCounts(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type memoUserStats struct {
	active      models.UserStats
	dailyActive models.UserStats
	total       models.UserStats

	memoized time.Time
}

var (
	userStatsCache         = memoUserStats{}
	userStatsCacheLimetime = 5 * time.Minute
)

func updateUserRoleCounts(ctx context.Context) error {
	query := `
SELECT role AS bitrole, active, COUNT(role) AS count FROM
  (SELECT last_seen_at>? AS active, last_seen_at>? AS daily_active, SUM(role) AS role
   FROM (SELECT
      u.id,
      CASE org_user.role
        WHEN 'Admin' THEN 4
        WHEN 'Editor' THEN 2
        ELSE 1
      END AS role,
      u.last_seen_at
    FROM ` + dialect.Quote("user") + ` AS u INNER JOIN org_user ON org_user.user_id = u.id
    GROUP BY u.id, u.last_seen_at, org_user.role) AS t2
  GROUP BY id, last_seen_at) AS t1
GROUP BY active, daily_active, role;`

	activeUserDeadline := time.Now().Add(-activeUserTimeLimit)
	dailyActiveUserDeadline := time.Now().Add(-dailyActiveUserTimeLimit)

	type rolebitmap struct {
		Active      bool
		DailyActive bool
		Bitrole     int64
		Count       int64
	}

	bitmap := []rolebitmap{}
	err := x.Context(ctx).SQL(query, activeUserDeadline, dailyActiveUserDeadline).Find(&bitmap)
	if err != nil {
		return err
	}

	memo := memoUserStats{memoized: time.Now()}
	for _, role := range bitmap {
		roletype := models.ROLE_VIEWER
		if role.Bitrole&0b100 != 0 {
			roletype = models.ROLE_ADMIN
		} else if role.Bitrole&0b10 != 0 {
			roletype = models.ROLE_EDITOR
		}

		memo.total = addToStats(memo.total, roletype, role.Count)
		if role.Active {
			memo.active = addToStats(memo.active, roletype, role.Count)
		}
		if role.DailyActive {
			memo.dailyActive = addToStats(memo.dailyActive, roletype, role.Count)
		}
	}

	userStatsCache = memo
	return nil
}

func addToStats(base models.UserStats, role models.RoleType, count int64) models.UserStats {
	base.Users += count

	switch role {
	case models.ROLE_ADMIN:
		base.Admins += count
	case models.ROLE_EDITOR:
		base.Editors += count
	default:
		base.Viewers += count
	}

	return base
}
