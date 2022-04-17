//go:build integration
// +build integration

package sqlstore

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAlertNotificationSQLAccess(t *testing.T) {
	Convey("Testing Alert notification sql access", t, func() {
		InitTestDB(t)

		Convey("Alert notification state", func() {
			var alertID int64 = 7
			var orgID int64 = 5
			var notifierID int64 = 10
			oldTimeNow := timeNow
			now := time.Date(2018, 9, 30, 0, 0, 0, 0, time.UTC)
			timeNow = func() time.Time { return now }

			Convey("Get no existing state should create a new state", func() {
				query := &models.GetOrCreateNotificationStateQuery{AlertId: alertID, OrgId: orgID, NotifierId: notifierID}
				err := GetOrCreateAlertNotificationState(context.Background(), query)
				So(err, ShouldBeNil)
				So(query.Result, ShouldNotBeNil)
				So(query.Result.State, ShouldEqual, "unknown")
				So(query.Result.Version, ShouldEqual, 0)
				So(query.Result.UpdatedAt, ShouldEqual, now.Unix())

				Convey("Get existing state should not create a new state", func() {
					query2 := &models.GetOrCreateNotificationStateQuery{AlertId: alertID, OrgId: orgID, NotifierId: notifierID}
					err := GetOrCreateAlertNotificationState(context.Background(), query2)
					So(err, ShouldBeNil)
					So(query2.Result, ShouldNotBeNil)
					So(query2.Result.Id, ShouldEqual, query.Result.Id)
					So(query2.Result.UpdatedAt, ShouldEqual, now.Unix())
				})

				Convey("Update existing state to pending with correct version should update database", func() {
					s := *query.Result

					cmd := models.SetAlertNotificationStateToPendingCommand{
						Id:                           s.Id,
						Version:                      s.Version,
						AlertRuleStateUpdatedVersion: s.AlertRuleStateUpdatedVersion,
					}

					err := SetAlertNotificationStateToPendingCommand(context.Background(), &cmd)
					So(err, ShouldBeNil)
					So(cmd.ResultVersion, ShouldEqual, 1)

					query2 := &models.GetOrCreateNotificationStateQuery{AlertId: alertID, OrgId: orgID, NotifierId: notifierID}
					err = GetOrCreateAlertNotificationState(context.Background(), query2)
					So(err, ShouldBeNil)
					So(query2.Result.Version, ShouldEqual, 1)
					So(query2.Result.State, ShouldEqual, models.AlertNotificationStatePending)
					So(query2.Result.UpdatedAt, ShouldEqual, now.Unix())

					Convey("Update existing state to completed should update database", func() {
						s := *query.Result
						setStateCmd := models.SetAlertNotificationStateToCompleteCommand{
							Id:      s.Id,
							Version: cmd.ResultVersion,
						}
						err := SetAlertNotificationStateToCompleteCommand(context.Background(), &setStateCmd)
						So(err, ShouldBeNil)

						query3 := &models.GetOrCreateNotificationStateQuery{AlertId: alertID, OrgId: orgID, NotifierId: notifierID}
						err = GetOrCreateAlertNotificationState(context.Background(), query3)
						So(err, ShouldBeNil)
						So(query3.Result.Version, ShouldEqual, 2)
						So(query3.Result.State, ShouldEqual, models.AlertNotificationStateCompleted)
						So(query3.Result.UpdatedAt, ShouldEqual, now.Unix())
					})

					Convey("Update existing state to completed should update database. regardless of version", func() {
						s := *query.Result
						unknownVersion := int64(1000)
						cmd := models.SetAlertNotificationStateToCompleteCommand{
							Id:      s.Id,
							Version: unknownVersion,
						}
						err := SetAlertNotificationStateToCompleteCommand(context.Background(), &cmd)
						So(err, ShouldBeNil)

						query3 := &models.GetOrCreateNotificationStateQuery{AlertId: alertID, OrgId: orgID, NotifierId: notifierID}
						err = GetOrCreateAlertNotificationState(context.Background(), query3)
						So(err, ShouldBeNil)
						So(query3.Result.Version, ShouldEqual, unknownVersion+1)
						So(query3.Result.State, ShouldEqual, models.AlertNotificationStateCompleted)
						So(query3.Result.UpdatedAt, ShouldEqual, now.Unix())
					})
				})

				Convey("Update existing state to pending with incorrect version should return version mismatch error", func() {
					s := *query.Result
					s.Version = 1000
					cmd := models.SetAlertNotificationStateToPendingCommand{
						Id:                           s.NotifierId,
						Version:                      s.Version,
						AlertRuleStateUpdatedVersion: s.AlertRuleStateUpdatedVersion,
					}
					err := SetAlertNotificationStateToPendingCommand(context.Background(), &cmd)
					So(err, ShouldEqual, models.ErrAlertNotificationStateVersionConflict)
				})

				Convey("Updating existing state to pending with incorrect version since alert rule state update version is higher", func() {
					s := *query.Result
					cmd := models.SetAlertNotificationStateToPendingCommand{
						Id:                           s.Id,
						Version:                      s.Version,
						AlertRuleStateUpdatedVersion: 1000,
					}
					err := SetAlertNotificationStateToPendingCommand(context.Background(), &cmd)
					So(err, ShouldBeNil)

					So(cmd.ResultVersion, ShouldEqual, 1)
				})

				Convey("different version and same alert state change version should return error", func() {
					s := *query.Result
					s.Version = 1000
					cmd := models.SetAlertNotificationStateToPendingCommand{
						Id:                           s.Id,
						Version:                      s.Version,
						AlertRuleStateUpdatedVersion: s.AlertRuleStateUpdatedVersion,
					}
					err := SetAlertNotificationStateToPendingCommand(context.Background(), &cmd)
					So(err, ShouldNotBeNil)
				})
			})

			Reset(func() {
				timeNow = oldTimeNow
			})
		})

		Convey("Alert notifications should be empty", func() {
			cmd := &models.GetAlertNotificationsQuery{
				OrgId: 2,
				Name:  "email",
			}

			err := GetAlertNotifications(cmd)
			So(err, ShouldBeNil)
			So(cmd.Result, ShouldBeNil)
		})

		Convey("Cannot save alert notifier with send reminder = true", func() {
			cmd := &models.CreateAlertNotificationCommand{
				Name:         "ops",
				Type:         "email",
				OrgId:        1,
				SendReminder: true,
				Settings:     simplejson.New(),
			}

			Convey("and missing frequency", func() {
				err := CreateAlertNotificationCommand(cmd)
				So(err, ShouldEqual, models.ErrNotificationFrequencyNotFound)
			})

			Convey("invalid frequency", func() {
				cmd.Frequency = "invalid duration"

				err := CreateAlertNotificationCommand(cmd)
				So(regexp.MustCompile(`^time: invalid duration "?invalid duration"?$`).MatchString(
					err.Error()), ShouldBeTrue)
			})
		})

		Convey("Cannot update alert notifier with send reminder = false", func() {
			cmd := &models.CreateAlertNotificationCommand{
				Name:         "ops update",
				Type:         "email",
				OrgId:        1,
				SendReminder: false,
				Settings:     simplejson.New(),
			}

			err := CreateAlertNotificationCommand(cmd)
			So(err, ShouldBeNil)

			updateCmd := &models.UpdateAlertNotificationCommand{
				Id:           cmd.Result.Id,
				SendReminder: true,
			}

			Convey("and missing frequency", func() {
				err := UpdateAlertNotification(updateCmd)
				So(err, ShouldEqual, models.ErrNotificationFrequencyNotFound)
			})

			Convey("invalid frequency", func() {
				updateCmd.Frequency = "invalid duration"

				err := UpdateAlertNotification(updateCmd)
				So(err, ShouldNotBeNil)
				So(regexp.MustCompile(`^time: invalid duration "?invalid duration"?$`).MatchString(
					err.Error()), ShouldBeTrue)
			})
		})

		Convey("Can save Alert Notification", func() {
			cmd := &models.CreateAlertNotificationCommand{
				Name:         "ops",
				Type:         "email",
				OrgId:        1,
				SendReminder: true,
				Frequency:    "10s",
				Settings:     simplejson.New(),
			}

			err := CreateAlertNotificationCommand(cmd)
			So(err, ShouldBeNil)
			So(cmd.Result.Id, ShouldNotEqual, 0)
			So(cmd.Result.OrgId, ShouldNotEqual, 0)
			So(cmd.Result.Type, ShouldEqual, "email")
			So(cmd.Result.Frequency, ShouldEqual, 10*time.Second)
			So(cmd.Result.DisableResolveMessage, ShouldBeFalse)
			So(cmd.Result.Uid, ShouldNotBeEmpty)

			Convey("Cannot save Alert Notification with the same name", func() {
				err = CreateAlertNotificationCommand(cmd)
				So(err, ShouldNotBeNil)
			})
			Convey("Cannot save Alert Notification with the same name and another uid", func() {
				anotherUidCmd := &models.CreateAlertNotificationCommand{
					Name:         cmd.Name,
					Type:         cmd.Type,
					OrgId:        1,
					SendReminder: cmd.SendReminder,
					Frequency:    cmd.Frequency,
					Settings:     cmd.Settings,
					Uid:          "notifier1",
				}
				err = CreateAlertNotificationCommand(anotherUidCmd)
				So(err, ShouldNotBeNil)
			})
			Convey("Can save Alert Notification with another name and another uid", func() {
				anotherUidCmd := &models.CreateAlertNotificationCommand{
					Name:         "another ops",
					Type:         cmd.Type,
					OrgId:        1,
					SendReminder: cmd.SendReminder,
					Frequency:    cmd.Frequency,
					Settings:     cmd.Settings,
					Uid:          "notifier2",
				}
				err = CreateAlertNotificationCommand(anotherUidCmd)
				So(err, ShouldBeNil)
			})

			Convey("Can update alert notification", func() {
				newCmd := &models.UpdateAlertNotificationCommand{
					Name:                  "NewName",
					Type:                  "webhook",
					OrgId:                 cmd.Result.OrgId,
					SendReminder:          true,
					DisableResolveMessage: true,
					Frequency:             "60s",
					Settings:              simplejson.New(),
					Id:                    cmd.Result.Id,
				}
				err := UpdateAlertNotification(newCmd)
				So(err, ShouldBeNil)
				So(newCmd.Result.Name, ShouldEqual, "NewName")
				So(newCmd.Result.Frequency, ShouldEqual, 60*time.Second)
				So(newCmd.Result.DisableResolveMessage, ShouldBeTrue)
			})

			Convey("Can update alert notification to disable sending of reminders", func() {
				newCmd := &models.UpdateAlertNotificationCommand{
					Name:         "NewName",
					Type:         "webhook",
					OrgId:        cmd.Result.OrgId,
					SendReminder: false,
					Settings:     simplejson.New(),
					Id:           cmd.Result.Id,
				}
				err := UpdateAlertNotification(newCmd)
				So(err, ShouldBeNil)
				So(newCmd.Result.SendReminder, ShouldBeFalse)
			})
		})

		Convey("Can search using an array of ids", func() {
			cmd1 := models.CreateAlertNotificationCommand{Name: "nagios", Type: "webhook", OrgId: 1, SendReminder: true, Frequency: "10s", Settings: simplejson.New()}
			cmd2 := models.CreateAlertNotificationCommand{Name: "slack", Type: "webhook", OrgId: 1, SendReminder: true, Frequency: "10s", Settings: simplejson.New()}
			cmd3 := models.CreateAlertNotificationCommand{Name: "ops2", Type: "email", OrgId: 1, SendReminder: true, Frequency: "10s", Settings: simplejson.New()}
			cmd4 := models.CreateAlertNotificationCommand{IsDefault: true, Name: "default", Type: "email", OrgId: 1, SendReminder: true, Frequency: "10s", Settings: simplejson.New()}

			otherOrg := models.CreateAlertNotificationCommand{Name: "default", Type: "email", OrgId: 2, SendReminder: true, Frequency: "10s", Settings: simplejson.New()}

			So(CreateAlertNotificationCommand(&cmd1), ShouldBeNil)
			So(CreateAlertNotificationCommand(&cmd2), ShouldBeNil)
			So(CreateAlertNotificationCommand(&cmd3), ShouldBeNil)
			So(CreateAlertNotificationCommand(&cmd4), ShouldBeNil)
			So(CreateAlertNotificationCommand(&otherOrg), ShouldBeNil)

			Convey("search", func() {
				query := &models.GetAlertNotificationsWithUidToSendQuery{
					Uids:  []string{cmd1.Result.Uid, cmd2.Result.Uid, "112341231"},
					OrgId: 1,
				}

				err := GetAlertNotificationsWithUidToSend(query)
				So(err, ShouldBeNil)
				So(len(query.Result), ShouldEqual, 3)
			})

			Convey("all", func() {
				query := &models.GetAllAlertNotificationsQuery{
					OrgId: 1,
				}

				err := GetAllAlertNotifications(query)
				So(err, ShouldBeNil)
				So(len(query.Result), ShouldEqual, 4)
				So(query.Result[0].Name, ShouldEqual, cmd4.Name)
				So(query.Result[1].Name, ShouldEqual, cmd1.Name)
				So(query.Result[2].Name, ShouldEqual, cmd3.Name)
				So(query.Result[3].Name, ShouldEqual, cmd2.Name)
			})
		})

		Convey("Notification Uid by Id Caching", func() {
			ss := InitTestDB(t)

			notification := &models.CreateAlertNotificationCommand{Uid: "aNotificationUid", OrgId: 1, Name: "aNotificationUid"}
			err := CreateAlertNotificationCommand(notification)
			So(err, ShouldBeNil)

			byUidQuery := &models.GetAlertNotificationsWithUidQuery{
				Uid:   notification.Uid,
				OrgId: notification.OrgId,
			}

			notificationByUidErr := GetAlertNotificationsWithUid(byUidQuery)
			So(notificationByUidErr, ShouldBeNil)

			Convey("Can cache notification Uid", func() {
				byIdQuery := &models.GetAlertNotificationUidQuery{
					Id:    byUidQuery.Result.Id,
					OrgId: byUidQuery.Result.OrgId,
				}

				cacheKey := newAlertNotificationUidCacheKey(byIdQuery.OrgId, byIdQuery.Id)

				resultBeforeCaching, foundBeforeCaching := ss.CacheService.Get(cacheKey)
				So(foundBeforeCaching, ShouldBeFalse)
				So(resultBeforeCaching, ShouldBeNil)

				notificationByIdErr := ss.GetAlertNotificationUidWithId(byIdQuery)
				So(notificationByIdErr, ShouldBeNil)

				resultAfterCaching, foundAfterCaching := ss.CacheService.Get(cacheKey)
				So(foundAfterCaching, ShouldBeTrue)
				So(resultAfterCaching, ShouldEqual, notification.Uid)
			})

			Convey("Retrieves from cache when exists", func() {
				query := &models.GetAlertNotificationUidQuery{
					Id:    999,
					OrgId: 100,
				}
				cacheKey := newAlertNotificationUidCacheKey(query.OrgId, query.Id)
				ss.CacheService.Set(cacheKey, "a-cached-uid", -1)

				err := ss.GetAlertNotificationUidWithId(query)
				So(err, ShouldBeNil)
				So(query.Result, ShouldEqual, "a-cached-uid")
			})

			Convey("Returns an error without populating cache when the notification doesn't exist in the database", func() {
				query := &models.GetAlertNotificationUidQuery{
					Id:    -1,
					OrgId: 100,
				}

				err := ss.GetAlertNotificationUidWithId(query)
				So(query.Result, ShouldEqual, "")
				So(err, ShouldNotBeNil)
				So(errors.Is(err, models.ErrAlertNotificationFailedTranslateUniqueID), ShouldBeTrue)

				cacheKey := newAlertNotificationUidCacheKey(query.OrgId, query.Id)
				result, found := ss.CacheService.Get(cacheKey)
				So(found, ShouldBeFalse)
				So(result, ShouldBeNil)
			})
		})

		Convey("Cannot update non-existing Alert Notification", func() {
			updateCmd := &models.UpdateAlertNotificationCommand{
				Name:                  "NewName",
				Type:                  "webhook",
				OrgId:                 1,
				SendReminder:          true,
				DisableResolveMessage: true,
				Frequency:             "60s",
				Settings:              simplejson.New(),
				Id:                    1,
			}
			err := UpdateAlertNotification(updateCmd)
			So(err, ShouldEqual, models.ErrAlertNotificationNotFound)

			Convey("using UID", func() {
				updateWithUidCmd := &models.UpdateAlertNotificationWithUidCommand{
					Name:                  "NewName",
					Type:                  "webhook",
					OrgId:                 1,
					SendReminder:          true,
					DisableResolveMessage: true,
					Frequency:             "60s",
					Settings:              simplejson.New(),
					Uid:                   "uid",
					NewUid:                "newUid",
				}
				err := UpdateAlertNotificationWithUid(updateWithUidCmd)
				So(err, ShouldEqual, models.ErrAlertNotificationNotFound)
			})
		})

		Convey("Can delete Alert Notification", func() {
			cmd := &models.CreateAlertNotificationCommand{
				Name:         "ops update",
				Type:         "email",
				OrgId:        1,
				SendReminder: false,
				Settings:     simplejson.New(),
			}

			err := CreateAlertNotificationCommand(cmd)
			So(err, ShouldBeNil)

			deleteCmd := &models.DeleteAlertNotificationCommand{
				Id:    cmd.Result.Id,
				OrgId: 1,
			}
			err = DeleteAlertNotification(deleteCmd)
			So(err, ShouldBeNil)

			Convey("using UID", func() {
				err := CreateAlertNotificationCommand(cmd)
				So(err, ShouldBeNil)

				deleteWithUidCmd := &models.DeleteAlertNotificationWithUidCommand{
					Uid:   cmd.Result.Uid,
					OrgId: 1,
				}
				err = DeleteAlertNotificationWithUid(deleteWithUidCmd)
				So(err, ShouldBeNil)
				So(deleteWithUidCmd.DeletedAlertNotificationId, ShouldEqual, cmd.Result.Id)
			})
		})

		Convey("Cannot delete non-existing Alert Notification", func() {
			deleteCmd := &models.DeleteAlertNotificationCommand{
				Id:    1,
				OrgId: 1,
			}
			err := DeleteAlertNotification(deleteCmd)
			So(err, ShouldEqual, models.ErrAlertNotificationNotFound)

			Convey("using UID", func() {
				deleteWithUidCmd := &models.DeleteAlertNotificationWithUidCommand{
					Uid:   "uid",
					OrgId: 1,
				}
				err = DeleteAlertNotificationWithUid(deleteWithUidCmd)
				So(err, ShouldEqual, models.ErrAlertNotificationNotFound)
			})
		})
	})
}
