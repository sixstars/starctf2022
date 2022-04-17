package auth

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/setting"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUserAuthToken(t *testing.T) {
	Convey("Test user auth token", t, func() {
		ctx := createTestContext(t)
		userAuthTokenService := ctx.tokenService
		user := &models.User{Id: int64(10)}
		userID := user.Id

		t := time.Date(2018, 12, 13, 13, 45, 0, 0, time.UTC)
		getTime = func() time.Time {
			return t
		}

		Convey("When creating token", func() {
			userToken, err := userAuthTokenService.CreateToken(context.Background(), user,
				net.ParseIP("192.168.10.11"), "some user agent")
			So(err, ShouldBeNil)
			So(userToken, ShouldNotBeNil)
			So(userToken.AuthTokenSeen, ShouldBeFalse)

			Convey("Can count active tokens", func() {
				count, err := userAuthTokenService.ActiveTokenCount(context.Background())
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1)
			})

			Convey("When lookup unhashed token should return user auth token", func() {
				userToken, err := userAuthTokenService.LookupToken(context.Background(), userToken.UnhashedToken)
				So(err, ShouldBeNil)
				So(userToken, ShouldNotBeNil)
				So(userToken.UserId, ShouldEqual, userID)
				So(userToken.AuthTokenSeen, ShouldBeTrue)

				storedAuthToken, err := ctx.getAuthTokenByID(userToken.Id)
				So(err, ShouldBeNil)
				So(storedAuthToken, ShouldNotBeNil)
				So(storedAuthToken.AuthTokenSeen, ShouldBeTrue)
			})

			Convey("When lookup hashed token should return user auth token not found error", func() {
				userToken, err := userAuthTokenService.LookupToken(context.Background(), userToken.AuthToken)
				So(err, ShouldEqual, models.ErrUserTokenNotFound)
				So(userToken, ShouldBeNil)
			})

			Convey("soft revoking existing token should not delete it", func() {
				err = userAuthTokenService.RevokeToken(context.Background(), userToken, true)
				So(err, ShouldBeNil)

				model, err := ctx.getAuthTokenByID(userToken.Id)
				So(err, ShouldBeNil)
				So(model, ShouldNotBeNil)
				So(model.RevokedAt, ShouldBeGreaterThan, 0)
			})

			Convey("revoking existing token should delete it", func() {
				err = userAuthTokenService.RevokeToken(context.Background(), userToken, false)
				So(err, ShouldBeNil)

				model, err := ctx.getAuthTokenByID(userToken.Id)
				So(err, ShouldBeNil)
				So(model, ShouldBeNil)
			})

			Convey("revoking nil token should return error", func() {
				err = userAuthTokenService.RevokeToken(context.Background(), nil, false)
				So(err, ShouldEqual, models.ErrUserTokenNotFound)
			})

			Convey("revoking non-existing token should return error", func() {
				userToken.Id = 1000
				err = userAuthTokenService.RevokeToken(context.Background(), userToken, false)
				So(err, ShouldEqual, models.ErrUserTokenNotFound)
			})

			Convey("When creating an additional token", func() {
				userToken2, err := userAuthTokenService.CreateToken(context.Background(), user,
					net.ParseIP("192.168.10.11"), "some user agent")
				So(err, ShouldBeNil)
				So(userToken2, ShouldNotBeNil)

				Convey("Can get first user token", func() {
					token, err := userAuthTokenService.GetUserToken(context.Background(), userID, userToken.Id)
					So(err, ShouldBeNil)
					So(token, ShouldNotBeNil)
					So(token.Id, ShouldEqual, userToken.Id)
				})

				Convey("Can get second user token", func() {
					token, err := userAuthTokenService.GetUserToken(context.Background(), userID, userToken2.Id)
					So(err, ShouldBeNil)
					So(token, ShouldNotBeNil)
					So(token.Id, ShouldEqual, userToken2.Id)
				})

				Convey("Can get user tokens", func() {
					tokens, err := userAuthTokenService.GetUserTokens(context.Background(), userID)
					So(err, ShouldBeNil)
					So(tokens, ShouldHaveLength, 2)
					So(tokens[0].Id, ShouldEqual, userToken.Id)
					So(tokens[1].Id, ShouldEqual, userToken2.Id)
				})

				Convey("Can revoke all user tokens", func() {
					err := userAuthTokenService.RevokeAllUserTokens(context.Background(), userID)
					So(err, ShouldBeNil)

					model, err := ctx.getAuthTokenByID(userToken.Id)
					So(err, ShouldBeNil)
					So(model, ShouldBeNil)

					model2, err := ctx.getAuthTokenByID(userToken2.Id)
					So(err, ShouldBeNil)
					So(model2, ShouldBeNil)
				})
			})

			Convey("When revoking users tokens in a batch", func() {
				Convey("Can revoke all users tokens", func() {
					userIds := []int64{}
					for i := 0; i < 3; i++ {
						userId := userID + int64(i+1)
						userIds = append(userIds, userId)
						_, err := userAuthTokenService.CreateToken(context.Background(), user,
							net.ParseIP("192.168.10.11"), "some user agent")
						So(err, ShouldBeNil)
					}

					err := userAuthTokenService.BatchRevokeAllUserTokens(context.Background(), userIds)
					So(err, ShouldBeNil)

					for _, v := range userIds {
						tokens, err := userAuthTokenService.GetUserTokens(context.Background(), v)
						So(err, ShouldBeNil)
						So(len(tokens), ShouldEqual, 0)
					}
				})
			})
		})

		Convey("expires correctly", func() {
			userToken, err := userAuthTokenService.CreateToken(context.Background(), user,
				net.ParseIP("192.168.10.11"), "some user agent")
			So(err, ShouldBeNil)

			userToken, err = userAuthTokenService.LookupToken(context.Background(), userToken.UnhashedToken)
			So(err, ShouldBeNil)

			getTime = func() time.Time {
				return t.Add(time.Hour)
			}

			rotated, err := userAuthTokenService.TryRotateToken(context.Background(), userToken,
				net.ParseIP("192.168.10.11"), "some user agent")
			So(err, ShouldBeNil)
			So(rotated, ShouldBeTrue)

			userToken, err = userAuthTokenService.LookupToken(context.Background(), userToken.UnhashedToken)
			So(err, ShouldBeNil)

			stillGood, err := userAuthTokenService.LookupToken(context.Background(), userToken.UnhashedToken)
			So(err, ShouldBeNil)
			So(stillGood, ShouldNotBeNil)

			model, err := ctx.getAuthTokenByID(userToken.Id)
			So(err, ShouldBeNil)

			Convey("when rotated_at is 6:59:59 ago should find token", func() {
				getTime = func() time.Time {
					return time.Unix(model.RotatedAt, 0).Add(24 * 7 * time.Hour).Add(-time.Second)
				}

				stillGood, err = userAuthTokenService.LookupToken(context.Background(), stillGood.UnhashedToken)
				So(err, ShouldBeNil)
				So(stillGood, ShouldNotBeNil)
			})

			Convey("when rotated_at is 7:00:00 ago should return token expired error", func() {
				getTime = func() time.Time {
					return time.Unix(model.RotatedAt, 0).Add(24 * 7 * time.Hour)
				}

				notGood, err := userAuthTokenService.LookupToken(context.Background(), userToken.UnhashedToken)
				So(err, ShouldHaveSameTypeAs, &models.TokenExpiredError{})
				So(notGood, ShouldBeNil)

				Convey("should not find active token when expired", func() {
					count, err := userAuthTokenService.ActiveTokenCount(context.Background())
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 0)
				})
			})

			Convey("when rotated_at is 5 days ago and created_at is 29 days and 23:59:59 ago should not find token", func() {
				updated, err := ctx.updateRotatedAt(model.Id, time.Unix(model.CreatedAt, 0).Add(24*25*time.Hour).Unix())
				So(err, ShouldBeNil)
				So(updated, ShouldBeTrue)

				getTime = func() time.Time {
					return time.Unix(model.CreatedAt, 0).Add(24 * 30 * time.Hour).Add(-time.Second)
				}

				stillGood, err = userAuthTokenService.LookupToken(context.Background(), stillGood.UnhashedToken)
				So(err, ShouldBeNil)
				So(stillGood, ShouldNotBeNil)
			})

			Convey("when rotated_at is 5 days ago and created_at is 30 days ago should return token expired error", func() {
				updated, err := ctx.updateRotatedAt(model.Id, time.Unix(model.CreatedAt, 0).Add(24*25*time.Hour).Unix())
				So(err, ShouldBeNil)
				So(updated, ShouldBeTrue)

				getTime = func() time.Time {
					return time.Unix(model.CreatedAt, 0).Add(24 * 30 * time.Hour)
				}

				notGood, err := userAuthTokenService.LookupToken(context.Background(), userToken.UnhashedToken)
				So(err, ShouldHaveSameTypeAs, &models.TokenExpiredError{})
				So(notGood, ShouldBeNil)
			})
		})

		Convey("can properly rotate tokens", func() {
			userToken, err := userAuthTokenService.CreateToken(context.Background(), user,
				net.ParseIP("192.168.10.11"), "some user agent")
			So(err, ShouldBeNil)

			prevToken := userToken.AuthToken
			unhashedPrev := userToken.UnhashedToken

			rotated, err := userAuthTokenService.TryRotateToken(context.Background(), userToken,
				net.ParseIP("192.168.10.12"), "a new user agent")
			So(err, ShouldBeNil)
			So(rotated, ShouldBeFalse)

			updated, err := ctx.markAuthTokenAsSeen(userToken.Id)
			So(err, ShouldBeNil)
			So(updated, ShouldBeTrue)

			model, err := ctx.getAuthTokenByID(userToken.Id)
			So(err, ShouldBeNil)

			var tok models.UserToken
			err = model.toUserToken(&tok)
			So(err, ShouldBeNil)

			getTime = func() time.Time {
				return t.Add(time.Hour)
			}

			rotated, err = userAuthTokenService.TryRotateToken(context.Background(), &tok,
				net.ParseIP("192.168.10.12"), "a new user agent")
			So(err, ShouldBeNil)
			So(rotated, ShouldBeTrue)

			unhashedToken := tok.UnhashedToken

			model, err = ctx.getAuthTokenByID(tok.Id)
			So(err, ShouldBeNil)
			model.UnhashedToken = unhashedToken

			So(model.RotatedAt, ShouldEqual, getTime().Unix())
			So(model.ClientIp, ShouldEqual, "192.168.10.12")
			So(model.UserAgent, ShouldEqual, "a new user agent")
			So(model.AuthTokenSeen, ShouldBeFalse)
			So(model.SeenAt, ShouldEqual, 0)
			So(model.PrevAuthToken, ShouldEqual, prevToken)

			// ability to auth using an old token

			lookedUpUserToken, err := userAuthTokenService.LookupToken(context.Background(), model.UnhashedToken)
			So(err, ShouldBeNil)
			So(lookedUpUserToken, ShouldNotBeNil)
			So(lookedUpUserToken.AuthTokenSeen, ShouldBeTrue)
			So(lookedUpUserToken.SeenAt, ShouldEqual, getTime().Unix())

			lookedUpUserToken, err = userAuthTokenService.LookupToken(context.Background(), unhashedPrev)
			So(err, ShouldBeNil)
			So(lookedUpUserToken, ShouldNotBeNil)
			So(lookedUpUserToken.Id, ShouldEqual, model.Id)
			So(lookedUpUserToken.AuthTokenSeen, ShouldBeTrue)

			getTime = func() time.Time {
				return t.Add(time.Hour + (2 * time.Minute))
			}

			lookedUpUserToken, err = userAuthTokenService.LookupToken(context.Background(), unhashedPrev)
			So(err, ShouldBeNil)
			So(lookedUpUserToken, ShouldNotBeNil)
			So(lookedUpUserToken.AuthTokenSeen, ShouldBeTrue)

			lookedUpModel, err := ctx.getAuthTokenByID(lookedUpUserToken.Id)
			So(err, ShouldBeNil)
			So(lookedUpModel, ShouldNotBeNil)
			So(lookedUpModel.AuthTokenSeen, ShouldBeFalse)

			rotated, err = userAuthTokenService.TryRotateToken(context.Background(), userToken,
				net.ParseIP("192.168.10.12"), "a new user agent")
			So(err, ShouldBeNil)
			So(rotated, ShouldBeTrue)

			model, err = ctx.getAuthTokenByID(userToken.Id)
			So(err, ShouldBeNil)
			So(model, ShouldNotBeNil)
			So(model.SeenAt, ShouldEqual, 0)
		})

		Convey("keeps prev token valid for 1 minute after it is confirmed", func() {
			userToken, err := userAuthTokenService.CreateToken(context.Background(), user,
				net.ParseIP("192.168.10.11"), "some user agent")
			So(err, ShouldBeNil)
			So(userToken, ShouldNotBeNil)

			lookedUpUserToken, err := userAuthTokenService.LookupToken(context.Background(), userToken.UnhashedToken)
			So(err, ShouldBeNil)
			So(lookedUpUserToken, ShouldNotBeNil)

			getTime = func() time.Time {
				return t.Add(10 * time.Minute)
			}

			prevToken := userToken.UnhashedToken
			rotated, err := userAuthTokenService.TryRotateToken(context.Background(), userToken,
				net.ParseIP("1.1.1.1"), "firefox")
			So(err, ShouldBeNil)
			So(rotated, ShouldBeTrue)

			getTime = func() time.Time {
				return t.Add(20 * time.Minute)
			}

			currentUserToken, err := userAuthTokenService.LookupToken(context.Background(), userToken.UnhashedToken)
			So(err, ShouldBeNil)
			So(currentUserToken, ShouldNotBeNil)

			prevUserToken, err := userAuthTokenService.LookupToken(context.Background(), prevToken)
			So(err, ShouldBeNil)
			So(prevUserToken, ShouldNotBeNil)
		})

		Convey("will not mark token unseen when prev and current are the same", func() {
			userToken, err := userAuthTokenService.CreateToken(context.Background(), user,
				net.ParseIP("192.168.10.11"), "some user agent")
			So(err, ShouldBeNil)
			So(userToken, ShouldNotBeNil)

			lookedUpUserToken, err := userAuthTokenService.LookupToken(context.Background(), userToken.UnhashedToken)
			So(err, ShouldBeNil)
			So(lookedUpUserToken, ShouldNotBeNil)

			lookedUpUserToken, err = userAuthTokenService.LookupToken(context.Background(), userToken.UnhashedToken)
			So(err, ShouldBeNil)
			So(lookedUpUserToken, ShouldNotBeNil)

			lookedUpModel, err := ctx.getAuthTokenByID(lookedUpUserToken.Id)
			So(err, ShouldBeNil)
			So(lookedUpModel, ShouldNotBeNil)
			So(lookedUpModel.AuthTokenSeen, ShouldBeTrue)
		})

		Convey("Rotate token", func() {
			userToken, err := userAuthTokenService.CreateToken(context.Background(), user,
				net.ParseIP("192.168.10.11"), "some user agent")
			So(err, ShouldBeNil)
			So(userToken, ShouldNotBeNil)

			prevToken := userToken.AuthToken

			Convey("Should rotate current token and previous token when auth token seen", func() {
				updated, err := ctx.markAuthTokenAsSeen(userToken.Id)
				So(err, ShouldBeNil)
				So(updated, ShouldBeTrue)

				getTime = func() time.Time {
					return t.Add(10 * time.Minute)
				}

				rotated, err := userAuthTokenService.TryRotateToken(context.Background(), userToken,
					net.ParseIP("1.1.1.1"), "firefox")
				So(err, ShouldBeNil)
				So(rotated, ShouldBeTrue)

				storedToken, err := ctx.getAuthTokenByID(userToken.Id)
				So(err, ShouldBeNil)
				So(storedToken, ShouldNotBeNil)
				So(storedToken.AuthTokenSeen, ShouldBeFalse)
				So(storedToken.PrevAuthToken, ShouldEqual, prevToken)
				So(storedToken.AuthToken, ShouldNotEqual, prevToken)

				prevToken = storedToken.AuthToken

				updated, err = ctx.markAuthTokenAsSeen(userToken.Id)
				So(err, ShouldBeNil)
				So(updated, ShouldBeTrue)

				getTime = func() time.Time {
					return t.Add(20 * time.Minute)
				}

				rotated, err = userAuthTokenService.TryRotateToken(context.Background(), userToken,
					net.ParseIP("1.1.1.1"), "firefox")
				So(err, ShouldBeNil)
				So(rotated, ShouldBeTrue)

				storedToken, err = ctx.getAuthTokenByID(userToken.Id)
				So(err, ShouldBeNil)
				So(storedToken, ShouldNotBeNil)
				So(storedToken.AuthTokenSeen, ShouldBeFalse)
				So(storedToken.PrevAuthToken, ShouldEqual, prevToken)
				So(storedToken.AuthToken, ShouldNotEqual, prevToken)
			})

			Convey("Should rotate current token, but keep previous token when auth token not seen", func() {
				userToken.RotatedAt = getTime().Add(-2 * time.Minute).Unix()

				getTime = func() time.Time {
					return t.Add(2 * time.Minute)
				}

				rotated, err := userAuthTokenService.TryRotateToken(context.Background(), userToken,
					net.ParseIP("1.1.1.1"), "firefox")
				So(err, ShouldBeNil)
				So(rotated, ShouldBeTrue)

				storedToken, err := ctx.getAuthTokenByID(userToken.Id)
				So(err, ShouldBeNil)
				So(storedToken, ShouldNotBeNil)
				So(storedToken.AuthTokenSeen, ShouldBeFalse)
				So(storedToken.PrevAuthToken, ShouldEqual, prevToken)
				So(storedToken.AuthToken, ShouldNotEqual, prevToken)
			})
		})

		Convey("When populating userAuthToken from UserToken should copy all properties", func() {
			ut := models.UserToken{
				Id:            1,
				UserId:        2,
				AuthToken:     "a",
				PrevAuthToken: "b",
				UserAgent:     "c",
				ClientIp:      "d",
				AuthTokenSeen: true,
				SeenAt:        3,
				RotatedAt:     4,
				CreatedAt:     5,
				UpdatedAt:     6,
				UnhashedToken: "e",
			}
			utBytes, err := json.Marshal(ut)
			So(err, ShouldBeNil)
			utJSON, err := simplejson.NewJson(utBytes)
			So(err, ShouldBeNil)
			utMap := utJSON.MustMap()

			var uat userAuthToken
			err = uat.fromUserToken(&ut)
			So(err, ShouldBeNil)
			uatBytes, err := json.Marshal(uat)
			So(err, ShouldBeNil)
			uatJSON, err := simplejson.NewJson(uatBytes)
			So(err, ShouldBeNil)
			uatMap := uatJSON.MustMap()

			So(uatMap, ShouldResemble, utMap)
		})

		Convey("When populating userToken from userAuthToken should copy all properties", func() {
			uat := userAuthToken{
				Id:            1,
				UserId:        2,
				AuthToken:     "a",
				PrevAuthToken: "b",
				UserAgent:     "c",
				ClientIp:      "d",
				AuthTokenSeen: true,
				SeenAt:        3,
				RotatedAt:     4,
				CreatedAt:     5,
				UpdatedAt:     6,
				UnhashedToken: "e",
			}
			uatBytes, err := json.Marshal(uat)
			So(err, ShouldBeNil)
			uatJSON, err := simplejson.NewJson(uatBytes)
			So(err, ShouldBeNil)
			uatMap := uatJSON.MustMap()

			var ut models.UserToken
			err = uat.toUserToken(&ut)
			So(err, ShouldBeNil)
			utBytes, err := json.Marshal(ut)
			So(err, ShouldBeNil)
			utJSON, err := simplejson.NewJson(utBytes)
			So(err, ShouldBeNil)
			utMap := utJSON.MustMap()

			So(utMap, ShouldResemble, uatMap)
		})

		Reset(func() {
			getTime = time.Now
		})
	})
}

func createTestContext(t *testing.T) *testContext {
	t.Helper()
	maxInactiveDurationVal, _ := time.ParseDuration("168h")
	maxLifetimeDurationVal, _ := time.ParseDuration("720h")
	sqlstore := sqlstore.InitTestDB(t)
	tokenService := &UserAuthTokenService{
		SQLStore: sqlstore,
		Cfg: &setting.Cfg{
			LoginMaxInactiveLifetime:     maxInactiveDurationVal,
			LoginMaxLifetime:             maxLifetimeDurationVal,
			TokenRotationIntervalMinutes: 10,
		},
		log: log.New("test-logger"),
	}

	return &testContext{
		sqlstore:     sqlstore,
		tokenService: tokenService,
	}
}

type testContext struct {
	sqlstore     *sqlstore.SQLStore
	tokenService *UserAuthTokenService
}

func (c *testContext) getAuthTokenByID(id int64) (*userAuthToken, error) {
	sess := c.sqlstore.NewSession(context.Background())
	var t userAuthToken
	found, err := sess.ID(id).Get(&t)
	if err != nil || !found {
		return nil, err
	}

	return &t, nil
}

func (c *testContext) markAuthTokenAsSeen(id int64) (bool, error) {
	sess := c.sqlstore.NewSession(context.Background())
	res, err := sess.Exec("UPDATE user_auth_token SET auth_token_seen = ? WHERE id = ?", c.sqlstore.Dialect.BooleanStr(true), id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected == 1, nil
}

func (c *testContext) updateRotatedAt(id, rotatedAt int64) (bool, error) {
	sess := c.sqlstore.NewSession(context.Background())
	res, err := sess.Exec("UPDATE user_auth_token SET rotated_at = ? WHERE id = ?", rotatedAt, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected == 1, nil
}
