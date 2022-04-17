package sqlstore

import (
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
)

const (
	alertRuleTarget = "alert_rule"
	dashboardTarget = "dashboard"
)

func init() {
	bus.AddHandler("sql", GetOrgQuotaByTarget)
	bus.AddHandler("sql", GetOrgQuotas)
	bus.AddHandler("sql", UpdateOrgQuota)
	bus.AddHandler("sql", GetUserQuotaByTarget)
	bus.AddHandler("sql", GetUserQuotas)
	bus.AddHandler("sql", UpdateUserQuota)
	bus.AddHandler("sql", GetGlobalQuotaByTarget)
}

type targetCount struct {
	Count int64
}

func GetOrgQuotaByTarget(query *models.GetOrgQuotaByTargetQuery) error {
	quota := models.Quota{
		Target: query.Target,
		OrgId:  query.OrgId,
	}
	has, err := x.Get(&quota)
	if err != nil {
		return err
	} else if !has {
		quota.Limit = query.Default
	}

	var used int64
	if query.Target != alertRuleTarget || query.UnifiedAlertingEnabled {
		// get quota used.
		rawSQL := fmt.Sprintf("SELECT COUNT(*) AS count FROM %s WHERE org_id=?",
			dialect.Quote(query.Target))

		if query.Target == dashboardTarget {
			rawSQL += fmt.Sprintf(" AND is_folder=%s", dialect.BooleanStr(false))
		}

		resp := make([]*targetCount, 0)
		if err := x.SQL(rawSQL, query.OrgId).Find(&resp); err != nil {
			return err
		}
		used = resp[0].Count
	}

	query.Result = &models.OrgQuotaDTO{
		Target: query.Target,
		Limit:  quota.Limit,
		OrgId:  query.OrgId,
		Used:   used,
	}

	return nil
}

func GetOrgQuotas(query *models.GetOrgQuotasQuery) error {
	quotas := make([]*models.Quota, 0)
	sess := x.Table("quota")
	if err := sess.Where("org_id=? AND user_id=0", query.OrgId).Find(&quotas); err != nil {
		return err
	}

	defaultQuotas := setting.Quota.Org.ToMap()

	seenTargets := make(map[string]bool)
	for _, q := range quotas {
		seenTargets[q.Target] = true
	}

	for t, v := range defaultQuotas {
		if _, ok := seenTargets[t]; !ok {
			quotas = append(quotas, &models.Quota{
				OrgId:  query.OrgId,
				Target: t,
				Limit:  v,
			})
		}
	}

	result := make([]*models.OrgQuotaDTO, len(quotas))
	for i, q := range quotas {
		var used int64
		if q.Target != alertRuleTarget || query.UnifiedAlertingEnabled {
			// get quota used.
			rawSQL := fmt.Sprintf("SELECT COUNT(*) as count from %s where org_id=?", dialect.Quote(q.Target))
			resp := make([]*targetCount, 0)
			if err := x.SQL(rawSQL, q.OrgId).Find(&resp); err != nil {
				return err
			}
			used = resp[0].Count
		}
		result[i] = &models.OrgQuotaDTO{
			Target: q.Target,
			Limit:  q.Limit,
			OrgId:  q.OrgId,
			Used:   used,
		}
	}
	query.Result = result
	return nil
}

func UpdateOrgQuota(cmd *models.UpdateOrgQuotaCmd) error {
	return inTransaction(func(sess *DBSession) error {
		// Check if quota is already defined in the DB
		quota := models.Quota{
			Target: cmd.Target,
			OrgId:  cmd.OrgId,
		}
		has, err := sess.Get(&quota)
		if err != nil {
			return err
		}
		quota.Updated = time.Now()
		quota.Limit = cmd.Limit
		if !has {
			quota.Created = time.Now()
			// No quota in the DB for this target, so create a new one.
			if _, err := sess.Insert(&quota); err != nil {
				return err
			}
		} else {
			// update existing quota entry in the DB.
			_, err := sess.ID(quota.Id).Update(&quota)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func GetUserQuotaByTarget(query *models.GetUserQuotaByTargetQuery) error {
	quota := models.Quota{
		Target: query.Target,
		UserId: query.UserId,
	}
	has, err := x.Get(&quota)
	if err != nil {
		return err
	} else if !has {
		quota.Limit = query.Default
	}

	var used int64
	if query.Target != alertRuleTarget || query.UnifiedAlertingEnabled {
		// get quota used.
		rawSQL := fmt.Sprintf("SELECT COUNT(*) as count from %s where user_id=?", dialect.Quote(query.Target))
		resp := make([]*targetCount, 0)
		if err := x.SQL(rawSQL, query.UserId).Find(&resp); err != nil {
			return err
		}
		used = resp[0].Count
	}

	query.Result = &models.UserQuotaDTO{
		Target: query.Target,
		Limit:  quota.Limit,
		UserId: query.UserId,
		Used:   used,
	}

	return nil
}

func GetUserQuotas(query *models.GetUserQuotasQuery) error {
	quotas := make([]*models.Quota, 0)
	sess := x.Table("quota")
	if err := sess.Where("user_id=? AND org_id=0", query.UserId).Find(&quotas); err != nil {
		return err
	}

	defaultQuotas := setting.Quota.User.ToMap()

	seenTargets := make(map[string]bool)
	for _, q := range quotas {
		seenTargets[q.Target] = true
	}

	for t, v := range defaultQuotas {
		if _, ok := seenTargets[t]; !ok {
			quotas = append(quotas, &models.Quota{
				UserId: query.UserId,
				Target: t,
				Limit:  v,
			})
		}
	}

	result := make([]*models.UserQuotaDTO, len(quotas))
	for i, q := range quotas {
		var used int64
		if q.Target != alertRuleTarget || query.UnifiedAlertingEnabled {
			// get quota used.
			rawSQL := fmt.Sprintf("SELECT COUNT(*) as count from %s where user_id=?", dialect.Quote(q.Target))
			resp := make([]*targetCount, 0)
			if err := x.SQL(rawSQL, q.UserId).Find(&resp); err != nil {
				return err
			}
			used = resp[0].Count
		}
		result[i] = &models.UserQuotaDTO{
			Target: q.Target,
			Limit:  q.Limit,
			UserId: q.UserId,
			Used:   used,
		}
	}
	query.Result = result
	return nil
}

func UpdateUserQuota(cmd *models.UpdateUserQuotaCmd) error {
	return inTransaction(func(sess *DBSession) error {
		// Check if quota is already defined in the DB
		quota := models.Quota{
			Target: cmd.Target,
			UserId: cmd.UserId,
		}
		has, err := sess.Get(&quota)
		if err != nil {
			return err
		}
		quota.Updated = time.Now()
		quota.Limit = cmd.Limit
		if !has {
			quota.Created = time.Now()
			// No quota in the DB for this target, so create a new one.
			if _, err := sess.Insert(&quota); err != nil {
				return err
			}
		} else {
			// update existing quota entry in the DB.
			_, err := sess.ID(quota.Id).Update(&quota)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func GetGlobalQuotaByTarget(query *models.GetGlobalQuotaByTargetQuery) error {
	var used int64
	if query.Target != alertRuleTarget || query.UnifiedAlertingEnabled {
		// get quota used.
		rawSQL := fmt.Sprintf("SELECT COUNT(*) AS count FROM %s",
			dialect.Quote(query.Target))

		if query.Target == dashboardTarget {
			rawSQL += fmt.Sprintf(" WHERE is_folder=%s", dialect.BooleanStr(false))
		}

		resp := make([]*targetCount, 0)
		if err := x.SQL(rawSQL).Find(&resp); err != nil {
			return err
		}
		used = resp[0].Count
	}

	query.Result = &models.GlobalQuotaDTO{
		Target: query.Target,
		Limit:  query.Default,
		Used:   used,
	}

	return nil
}
