package setting

import (
	"reflect"
)

type OrgQuota struct {
	User       int64 `target:"org_user"`
	DataSource int64 `target:"data_source"`
	Dashboard  int64 `target:"dashboard"`
	ApiKey     int64 `target:"api_key"`
	AlertRule  int64 `target:"alert_rule"`
}

type UserQuota struct {
	Org int64 `target:"org_user"`
}

type GlobalQuota struct {
	Org        int64 `target:"org"`
	User       int64 `target:"user"`
	DataSource int64 `target:"data_source"`
	Dashboard  int64 `target:"dashboard"`
	ApiKey     int64 `target:"api_key"`
	Session    int64 `target:"-"`
	AlertRule  int64 `target:"alert_rule"`
}

func (q *OrgQuota) ToMap() map[string]int64 {
	return quotaToMap(*q)
}

func (q *UserQuota) ToMap() map[string]int64 {
	return quotaToMap(*q)
}

func quotaToMap(q interface{}) map[string]int64 {
	qMap := make(map[string]int64)
	typ := reflect.TypeOf(q)
	val := reflect.ValueOf(q)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name := field.Tag.Get("target")
		if name == "" {
			name = field.Name
		}
		if name == "-" {
			continue
		}
		value := val.Field(i)
		qMap[name] = value.Int()
	}
	return qMap
}

type QuotaSettings struct {
	Enabled bool
	Org     *OrgQuota
	User    *UserQuota
	Global  *GlobalQuota
}

func (cfg *Cfg) readQuotaSettings() {
	// set global defaults.
	quota := cfg.Raw.Section("quota")
	Quota.Enabled = quota.Key("enabled").MustBool(false)

	var alertOrgQuota int64
	var alertGlobalQuota int64
	if cfg.UnifiedAlerting.Enabled {
		alertOrgQuota = quota.Key("org_alert_rule").MustInt64(100)
		alertGlobalQuota = quota.Key("global_alert_rule").MustInt64(-1)
	}
	// per ORG Limits
	Quota.Org = &OrgQuota{
		User:       quota.Key("org_user").MustInt64(10),
		DataSource: quota.Key("org_data_source").MustInt64(10),
		Dashboard:  quota.Key("org_dashboard").MustInt64(10),
		ApiKey:     quota.Key("org_api_key").MustInt64(10),
		AlertRule:  alertOrgQuota,
	}

	// per User limits
	Quota.User = &UserQuota{
		Org: quota.Key("user_org").MustInt64(10),
	}

	// Global Limits
	Quota.Global = &GlobalQuota{
		User:       quota.Key("global_user").MustInt64(-1),
		Org:        quota.Key("global_org").MustInt64(-1),
		DataSource: quota.Key("global_data_source").MustInt64(-1),
		Dashboard:  quota.Key("global_dashboard").MustInt64(-1),
		ApiKey:     quota.Key("global_api_key").MustInt64(-1),
		Session:    quota.Key("global_session").MustInt64(-1),
		AlertRule:  alertGlobalQuota,
	}

	cfg.Quota = Quota
}
