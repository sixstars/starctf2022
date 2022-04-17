package sqlstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/securejsondata"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/util"
)

func init() {
	bus.AddHandler("sql", GetAlertNotifications)
	bus.AddHandler("sql", CreateAlertNotificationCommand)
	bus.AddHandler("sql", UpdateAlertNotification)
	bus.AddHandler("sql", DeleteAlertNotification)
	bus.AddHandler("sql", GetAllAlertNotifications)
	bus.AddHandlerCtx("sql", GetOrCreateAlertNotificationState)
	bus.AddHandlerCtx("sql", SetAlertNotificationStateToCompleteCommand)
	bus.AddHandlerCtx("sql", SetAlertNotificationStateToPendingCommand)

	bus.AddHandler("sql", GetAlertNotificationsWithUid)
	bus.AddHandler("sql", UpdateAlertNotificationWithUid)
	bus.AddHandler("sql", DeleteAlertNotificationWithUid)
	bus.AddHandler("sql", GetAlertNotificationsWithUidToSend)
}

func DeleteAlertNotification(cmd *models.DeleteAlertNotificationCommand) error {
	return inTransaction(func(sess *DBSession) error {
		sql := "DELETE FROM alert_notification WHERE alert_notification.org_id = ? AND alert_notification.id = ?"
		res, err := sess.Exec(sql, cmd.OrgId, cmd.Id)
		if err != nil {
			return err
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return models.ErrAlertNotificationNotFound
		}

		if _, err := sess.Exec("DELETE FROM alert_notification_state WHERE alert_notification_state.org_id = ? AND alert_notification_state.notifier_id = ?", cmd.OrgId, cmd.Id); err != nil {
			return err
		}

		return nil
	})
}

func DeleteAlertNotificationWithUid(cmd *models.DeleteAlertNotificationWithUidCommand) error {
	existingNotification := &models.GetAlertNotificationsWithUidQuery{OrgId: cmd.OrgId, Uid: cmd.Uid}
	if err := getAlertNotificationWithUidInternal(existingNotification, newSession(context.Background())); err != nil {
		return err
	}

	if existingNotification.Result == nil {
		return models.ErrAlertNotificationNotFound
	}

	cmd.DeletedAlertNotificationId = existingNotification.Result.Id
	deleteCommand := &models.DeleteAlertNotificationCommand{
		Id:    existingNotification.Result.Id,
		OrgId: existingNotification.Result.OrgId,
	}
	if err := bus.Dispatch(deleteCommand); err != nil {
		return err
	}

	return nil
}

func GetAlertNotifications(query *models.GetAlertNotificationsQuery) error {
	return getAlertNotificationInternal(query, newSession(context.Background()))
}

func (ss *SQLStore) addAlertNotificationUidByIdHandler() {
	bus.AddHandler("sql", ss.GetAlertNotificationUidWithId)
}

func (ss *SQLStore) GetAlertNotificationUidWithId(query *models.GetAlertNotificationUidQuery) error {
	cacheKey := newAlertNotificationUidCacheKey(query.OrgId, query.Id)

	if cached, found := ss.CacheService.Get(cacheKey); found {
		query.Result = cached.(string)
		return nil
	}

	err := getAlertNotificationUidInternal(query, newSession(context.Background()))
	if err != nil {
		return err
	}

	ss.CacheService.Set(cacheKey, query.Result, -1) // Infinite, never changes

	return nil
}

func newAlertNotificationUidCacheKey(orgID, notificationId int64) string {
	return fmt.Sprintf("notification-uid-by-org-%d-and-id-%d", orgID, notificationId)
}

func GetAlertNotificationsWithUid(query *models.GetAlertNotificationsWithUidQuery) error {
	return getAlertNotificationWithUidInternal(query, newSession(context.Background()))
}

func GetAllAlertNotifications(query *models.GetAllAlertNotificationsQuery) error {
	results := make([]*models.AlertNotification, 0)
	if err := x.Where("org_id = ?", query.OrgId).Asc("name").Find(&results); err != nil {
		return err
	}

	query.Result = results
	return nil
}

func GetAlertNotificationsWithUidToSend(query *models.GetAlertNotificationsWithUidToSendQuery) error {
	var sql bytes.Buffer
	params := make([]interface{}, 0)

	sql.WriteString(`SELECT
										alert_notification.id,
										alert_notification.uid,
										alert_notification.org_id,
										alert_notification.name,
										alert_notification.type,
										alert_notification.created,
										alert_notification.updated,
										alert_notification.settings,
										alert_notification.secure_settings,
										alert_notification.is_default,
										alert_notification.disable_resolve_message,
										alert_notification.send_reminder,
										alert_notification.frequency
										FROM alert_notification
	  							`)

	sql.WriteString(` WHERE alert_notification.org_id = ?`)
	params = append(params, query.OrgId)

	sql.WriteString(` AND ((alert_notification.is_default = ?)`)
	params = append(params, dialect.BooleanStr(true))

	if len(query.Uids) > 0 {
		sql.WriteString(` OR alert_notification.uid IN (?` + strings.Repeat(",?", len(query.Uids)-1) + ")")
		for _, v := range query.Uids {
			params = append(params, v)
		}
	}
	sql.WriteString(`)`)

	results := make([]*models.AlertNotification, 0)
	if err := x.SQL(sql.String(), params...).Find(&results); err != nil {
		return err
	}

	query.Result = results
	return nil
}

func getAlertNotificationUidInternal(query *models.GetAlertNotificationUidQuery, sess *DBSession) error {
	var sql bytes.Buffer
	params := make([]interface{}, 0)

	sql.WriteString(`SELECT
										alert_notification.uid
										FROM alert_notification
	  							`)

	sql.WriteString(` WHERE alert_notification.org_id = ?`)
	params = append(params, query.OrgId)

	sql.WriteString(` AND alert_notification.id = ?`)
	params = append(params, query.Id)

	results := make([]string, 0)
	if err := sess.SQL(sql.String(), params...).Find(&results); err != nil {
		return err
	}

	if len(results) == 0 {
		return models.ErrAlertNotificationFailedTranslateUniqueID
	}

	query.Result = results[0]

	return nil
}

func getAlertNotificationInternal(query *models.GetAlertNotificationsQuery, sess *DBSession) error {
	var sql bytes.Buffer
	params := make([]interface{}, 0)

	sql.WriteString(`SELECT
										alert_notification.id,
										alert_notification.uid,
										alert_notification.org_id,
										alert_notification.name,
										alert_notification.type,
										alert_notification.created,
										alert_notification.updated,
										alert_notification.settings,
										alert_notification.secure_settings,
										alert_notification.is_default,
										alert_notification.disable_resolve_message,
										alert_notification.send_reminder,
										alert_notification.frequency
										FROM alert_notification
	  							`)

	sql.WriteString(` WHERE alert_notification.org_id = ?`)
	params = append(params, query.OrgId)

	if query.Name != "" || query.Id != 0 {
		if query.Name != "" {
			sql.WriteString(` AND alert_notification.name = ?`)
			params = append(params, query.Name)
		}

		if query.Id != 0 {
			sql.WriteString(` AND alert_notification.id = ?`)
			params = append(params, query.Id)
		}
	}

	results := make([]*models.AlertNotification, 0)
	if err := sess.SQL(sql.String(), params...).Find(&results); err != nil {
		return err
	}

	if len(results) == 0 {
		query.Result = nil
	} else {
		query.Result = results[0]
	}

	return nil
}

func getAlertNotificationWithUidInternal(query *models.GetAlertNotificationsWithUidQuery, sess *DBSession) error {
	var sql bytes.Buffer
	params := make([]interface{}, 0)

	sql.WriteString(`SELECT
										alert_notification.id,
										alert_notification.uid,
										alert_notification.org_id,
										alert_notification.name,
										alert_notification.type,
										alert_notification.created,
										alert_notification.updated,
										alert_notification.settings,
										alert_notification.secure_settings,
										alert_notification.is_default,
										alert_notification.disable_resolve_message,
										alert_notification.send_reminder,
										alert_notification.frequency
										FROM alert_notification
	  							`)

	sql.WriteString(` WHERE alert_notification.org_id = ? AND alert_notification.uid = ?`)
	params = append(params, query.OrgId, query.Uid)

	results := make([]*models.AlertNotification, 0)
	if err := sess.SQL(sql.String(), params...).Find(&results); err != nil {
		return err
	}

	if len(results) == 0 {
		query.Result = nil
	} else {
		query.Result = results[0]
	}

	return nil
}

func CreateAlertNotificationCommand(cmd *models.CreateAlertNotificationCommand) error {
	return inTransaction(func(sess *DBSession) error {
		if cmd.Uid == "" {
			uid, uidGenerationErr := generateNewAlertNotificationUid(sess, cmd.OrgId)
			if uidGenerationErr != nil {
				return uidGenerationErr
			}

			cmd.Uid = uid
		}
		existingQuery := &models.GetAlertNotificationsWithUidQuery{OrgId: cmd.OrgId, Uid: cmd.Uid}
		err := getAlertNotificationWithUidInternal(existingQuery, sess)

		if err != nil {
			return err
		}

		if existingQuery.Result != nil {
			return models.ErrAlertNotificationWithSameUIDExists
		}

		// check if name exists
		sameNameQuery := &models.GetAlertNotificationsQuery{OrgId: cmd.OrgId, Name: cmd.Name}
		if err := getAlertNotificationInternal(sameNameQuery, sess); err != nil {
			return err
		}

		if sameNameQuery.Result != nil {
			return models.ErrAlertNotificationWithSameNameExists
		}

		var frequency time.Duration
		if cmd.SendReminder {
			if cmd.Frequency == "" {
				return models.ErrNotificationFrequencyNotFound
			}

			frequency, err = time.ParseDuration(cmd.Frequency)
			if err != nil {
				return err
			}
		}

		// delete empty keys
		for k, v := range cmd.SecureSettings {
			if v == "" {
				delete(cmd.SecureSettings, k)
			}
		}

		alertNotification := &models.AlertNotification{
			Uid:                   cmd.Uid,
			OrgId:                 cmd.OrgId,
			Name:                  cmd.Name,
			Type:                  cmd.Type,
			Settings:              cmd.Settings,
			SecureSettings:        securejsondata.GetEncryptedJsonData(cmd.SecureSettings),
			SendReminder:          cmd.SendReminder,
			DisableResolveMessage: cmd.DisableResolveMessage,
			Frequency:             frequency,
			Created:               time.Now(),
			Updated:               time.Now(),
			IsDefault:             cmd.IsDefault,
		}

		if _, err = sess.MustCols("send_reminder").Insert(alertNotification); err != nil {
			return err
		}

		cmd.Result = alertNotification
		return nil
	})
}

func generateNewAlertNotificationUid(sess *DBSession, orgId int64) (string, error) {
	for i := 0; i < 3; i++ {
		uid := util.GenerateShortUID()
		exists, err := sess.Where("org_id=? AND uid=?", orgId, uid).Get(&models.AlertNotification{})
		if err != nil {
			return "", err
		}

		if !exists {
			return uid, nil
		}
	}

	return "", models.ErrAlertNotificationFailedGenerateUniqueUid
}

func UpdateAlertNotification(cmd *models.UpdateAlertNotificationCommand) error {
	return inTransaction(func(sess *DBSession) (err error) {
		current := models.AlertNotification{}

		if _, err = sess.ID(cmd.Id).Get(&current); err != nil {
			return err
		}

		if current.Id == 0 {
			return models.ErrAlertNotificationNotFound
		}

		// check if name exists
		sameNameQuery := &models.GetAlertNotificationsQuery{OrgId: cmd.OrgId, Name: cmd.Name}
		if err := getAlertNotificationInternal(sameNameQuery, sess); err != nil {
			return err
		}

		if sameNameQuery.Result != nil && sameNameQuery.Result.Id != current.Id {
			return fmt.Errorf("alert notification name %q already exists", cmd.Name)
		}

		// delete empty keys
		for k, v := range cmd.SecureSettings {
			if v == "" {
				delete(cmd.SecureSettings, k)
			}
		}

		current.Updated = time.Now()
		current.Settings = cmd.Settings
		current.SecureSettings = securejsondata.GetEncryptedJsonData(cmd.SecureSettings)
		current.Name = cmd.Name
		current.Type = cmd.Type
		current.IsDefault = cmd.IsDefault
		current.SendReminder = cmd.SendReminder
		current.DisableResolveMessage = cmd.DisableResolveMessage

		if cmd.Uid != "" {
			current.Uid = cmd.Uid
		}

		if current.SendReminder {
			if cmd.Frequency == "" {
				return models.ErrNotificationFrequencyNotFound
			}

			frequency, err := time.ParseDuration(cmd.Frequency)
			if err != nil {
				return err
			}

			current.Frequency = frequency
		}

		sess.UseBool("is_default", "send_reminder", "disable_resolve_message")

		if affected, err := sess.ID(cmd.Id).Update(current); err != nil {
			return err
		} else if affected == 0 {
			return fmt.Errorf("could not update alert notification")
		}

		cmd.Result = &current
		return nil
	})
}

func UpdateAlertNotificationWithUid(cmd *models.UpdateAlertNotificationWithUidCommand) error {
	getAlertNotificationWithUidQuery := &models.GetAlertNotificationsWithUidQuery{OrgId: cmd.OrgId, Uid: cmd.Uid}

	if err := getAlertNotificationWithUidInternal(getAlertNotificationWithUidQuery, newSession(context.Background())); err != nil {
		return err
	}

	current := getAlertNotificationWithUidQuery.Result

	if current == nil {
		return models.ErrAlertNotificationNotFound
	}

	if cmd.NewUid == "" {
		cmd.NewUid = cmd.Uid
	}

	updateNotification := &models.UpdateAlertNotificationCommand{
		Id:                    current.Id,
		Uid:                   cmd.NewUid,
		Name:                  cmd.Name,
		Type:                  cmd.Type,
		SendReminder:          cmd.SendReminder,
		DisableResolveMessage: cmd.DisableResolveMessage,
		Frequency:             cmd.Frequency,
		IsDefault:             cmd.IsDefault,
		Settings:              cmd.Settings,
		SecureSettings:        cmd.SecureSettings,

		OrgId: cmd.OrgId,
	}

	if err := bus.Dispatch(updateNotification); err != nil {
		return err
	}

	cmd.Result = updateNotification.Result

	return nil
}

func SetAlertNotificationStateToCompleteCommand(ctx context.Context, cmd *models.SetAlertNotificationStateToCompleteCommand) error {
	return inTransactionCtx(ctx, func(sess *DBSession) error {
		version := cmd.Version
		var current models.AlertNotificationState
		if _, err := sess.ID(cmd.Id).Get(&current); err != nil {
			return err
		}

		newVersion := cmd.Version + 1
		sql := `UPDATE alert_notification_state SET
			state = ?,
			version = ?,
			updated_at = ?
		WHERE
			id = ?`

		_, err := sess.Exec(sql, models.AlertNotificationStateCompleted, newVersion, timeNow().Unix(), cmd.Id)
		if err != nil {
			return err
		}

		if current.Version != version {
			sqlog.Error("notification state out of sync. the notification is marked as complete but has been modified between set as pending and completion.", "notifierId", current.NotifierId)
		}

		return nil
	})
}

func SetAlertNotificationStateToPendingCommand(ctx context.Context, cmd *models.SetAlertNotificationStateToPendingCommand) error {
	return withDbSession(ctx, x, func(sess *DBSession) error {
		newVersion := cmd.Version + 1
		sql := `UPDATE alert_notification_state SET
			state = ?,
			version = ?,
			updated_at = ?,
			alert_rule_state_updated_version = ?
		WHERE
			id = ? AND
			(version = ? OR alert_rule_state_updated_version < ?)`

		res, err := sess.Exec(sql,
			models.AlertNotificationStatePending,
			newVersion,
			timeNow().Unix(),
			cmd.AlertRuleStateUpdatedVersion,
			cmd.Id,
			cmd.Version,
			cmd.AlertRuleStateUpdatedVersion)

		if err != nil {
			return err
		}

		affected, _ := res.RowsAffected()
		if affected == 0 {
			return models.ErrAlertNotificationStateVersionConflict
		}

		cmd.ResultVersion = newVersion

		return nil
	})
}

func GetOrCreateAlertNotificationState(ctx context.Context, cmd *models.GetOrCreateNotificationStateQuery) error {
	return inTransactionCtx(ctx, func(sess *DBSession) error {
		nj := &models.AlertNotificationState{}

		exist, err := getAlertNotificationState(sess, cmd, nj)

		// if exists, return it, otherwise create it with default values
		if err != nil {
			return err
		}

		if exist {
			cmd.Result = nj
			return nil
		}

		notificationState := &models.AlertNotificationState{
			OrgId:      cmd.OrgId,
			AlertId:    cmd.AlertId,
			NotifierId: cmd.NotifierId,
			State:      models.AlertNotificationStateUnknown,
			UpdatedAt:  timeNow().Unix(),
		}

		if _, err := sess.Insert(notificationState); err != nil {
			if dialect.IsUniqueConstraintViolation(err) {
				exist, err = getAlertNotificationState(sess, cmd, nj)

				if err != nil {
					return err
				}

				if !exist {
					return errors.New("should not happen")
				}

				cmd.Result = nj
				return nil
			}

			return err
		}

		cmd.Result = notificationState
		return nil
	})
}

func getAlertNotificationState(sess *DBSession, cmd *models.GetOrCreateNotificationStateQuery, nj *models.AlertNotificationState) (bool, error) {
	return sess.
		Where("alert_notification_state.org_id = ?", cmd.OrgId).
		Where("alert_notification_state.alert_id = ?", cmd.AlertId).
		Where("alert_notification_state.notifier_id = ?", cmd.NotifierId).
		Get(nj)
}
