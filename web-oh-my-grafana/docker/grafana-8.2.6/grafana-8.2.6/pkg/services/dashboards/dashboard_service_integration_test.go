//go:build integration
// +build integration

package dashboards

import (
	"testing"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/dashboards"
	"github.com/grafana/grafana/pkg/services/guardian"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/models"
)

const testOrgID int64 = 1

func TestIntegratedDashboardService(t *testing.T) {
	t.Run("Given saved folders and dashboards in organization A", func(t *testing.T) {
		origUpdateAlerting := UpdateAlerting
		t.Cleanup(func() {
			UpdateAlerting = origUpdateAlerting
		})
		UpdateAlerting = func(store dashboards.Store, orgID int64, dashboard *models.Dashboard, user *models.SignedInUser) error {
			return nil
		}

		// Basic validation tests

		permissionScenario(t, "When saving a dashboard with non-existing id", true,
			func(t *testing.T, sc *permissionScenarioContext) {
				cmd := models.SaveDashboardCommand{
					OrgId: testOrgID,
					Dashboard: simplejson.NewFromAny(map[string]interface{}{
						"id":    float64(123412321),
						"title": "Expect error",
					}),
				}

				err := callSaveWithError(cmd, sc.sqlStore)
				assert.Equal(t, models.ErrDashboardNotFound, err)
			})

		// Given other organization

		t.Run("Given organization B", func(t *testing.T) {
			const otherOrgId int64 = 2

			permissionScenario(t, "When creating a dashboard with same id as dashboard in organization A",
				true, func(t *testing.T, sc *permissionScenarioContext) {
					cmd := models.SaveDashboardCommand{
						OrgId: otherOrgId,
						Dashboard: simplejson.NewFromAny(map[string]interface{}{
							"id":    sc.savedDashInFolder.Id,
							"title": "Expect error",
						}),
						Overwrite: false,
					}

					err := callSaveWithError(cmd, sc.sqlStore)
					assert.Equal(t, models.ErrDashboardNotFound, err)
				})

			permissionScenario(t, "When creating a dashboard with same uid as dashboard in organization A, it should create a new dashboard in org B",
				true, func(t *testing.T, sc *permissionScenarioContext) {
					const otherOrgId int64 = 2
					cmd := models.SaveDashboardCommand{
						OrgId: otherOrgId,
						Dashboard: simplejson.NewFromAny(map[string]interface{}{
							"uid":   sc.savedDashInFolder.Uid,
							"title": "Dash with existing uid in other org",
						}),
						Overwrite: false,
					}

					res := callSaveWithResult(t, cmd, sc.sqlStore)
					require.NotNil(t, res)

					dash, err := sc.sqlStore.GetDashboard(0, otherOrgId, sc.savedDashInFolder.Uid, "")
					require.NoError(t, err)

					assert.NotEqual(t, sc.savedDashInFolder.Id, dash.Id)
					assert.Equal(t, res.Id, dash.Id)
					assert.Equal(t, otherOrgId, dash.OrgId)
					assert.Equal(t, sc.savedDashInFolder.Uid, dash.Uid)
				})
		})

		t.Run("Given user has no permission to save", func(t *testing.T) {
			const canSave = false

			permissionScenario(t, "When creating a new dashboard in the General folder", canSave,
				func(t *testing.T, sc *permissionScenarioContext) {
					sqlStore := sqlstore.InitTestDB(t)
					cmd := models.SaveDashboardCommand{
						OrgId: testOrgID,
						Dashboard: simplejson.NewFromAny(map[string]interface{}{
							"title": "Dash",
						}),
						UserId:    10000,
						Overwrite: true,
					}

					err := callSaveWithError(cmd, sqlStore)
					assert.Equal(t, models.ErrDashboardUpdateAccessDenied, err)

					assert.Equal(t, int64(0), sc.dashboardGuardianMock.DashId)
					assert.Equal(t, cmd.OrgId, sc.dashboardGuardianMock.OrgId)
					assert.Equal(t, cmd.UserId, sc.dashboardGuardianMock.User.UserId)
				})

			permissionScenario(t, "When creating a new dashboard in other folder, it should create dashboard guardian for other folder with correct arguments and rsult in access denied error",
				canSave, func(t *testing.T, sc *permissionScenarioContext) {
					cmd := models.SaveDashboardCommand{
						OrgId: testOrgID,
						Dashboard: simplejson.NewFromAny(map[string]interface{}{
							"title": "Dash",
						}),
						FolderId:  sc.otherSavedFolder.Id,
						UserId:    10000,
						Overwrite: true,
					}

					err := callSaveWithError(cmd, sc.sqlStore)
					require.Equal(t, models.ErrDashboardUpdateAccessDenied, err)

					assert.Equal(t, sc.otherSavedFolder.Id, sc.dashboardGuardianMock.DashId)
					assert.Equal(t, cmd.OrgId, sc.dashboardGuardianMock.OrgId)
					assert.Equal(t, cmd.UserId, sc.dashboardGuardianMock.User.UserId)
				})

			permissionScenario(t, "When creating a new dashboard by existing title in folder, it should create dashboard guardian for folder with correct arguments and result in access denied error",
				canSave, func(t *testing.T, sc *permissionScenarioContext) {
					cmd := models.SaveDashboardCommand{
						OrgId: testOrgID,
						Dashboard: simplejson.NewFromAny(map[string]interface{}{
							"title": sc.savedDashInFolder.Title,
						}),
						FolderId:  sc.savedFolder.Id,
						UserId:    10000,
						Overwrite: true,
					}

					err := callSaveWithError(cmd, sc.sqlStore)
					require.Equal(t, models.ErrDashboardUpdateAccessDenied, err)

					assert.Equal(t, sc.savedFolder.Id, sc.dashboardGuardianMock.DashId)
					assert.Equal(t, cmd.OrgId, sc.dashboardGuardianMock.OrgId)
					assert.Equal(t, cmd.UserId, sc.dashboardGuardianMock.User.UserId)
				})

			permissionScenario(t, "When creating a new dashboard by existing UID in folder, it should create dashboard guardian for folder with correct arguments and result in access denied error",
				canSave, func(t *testing.T, sc *permissionScenarioContext) {
					cmd := models.SaveDashboardCommand{
						OrgId: testOrgID,
						Dashboard: simplejson.NewFromAny(map[string]interface{}{
							"uid":   sc.savedDashInFolder.Uid,
							"title": "New dash",
						}),
						FolderId:  sc.savedFolder.Id,
						UserId:    10000,
						Overwrite: true,
					}

					err := callSaveWithError(cmd, sc.sqlStore)
					require.Equal(t, models.ErrDashboardUpdateAccessDenied, err)

					assert.Equal(t, sc.savedFolder.Id, sc.dashboardGuardianMock.DashId)
					assert.Equal(t, cmd.OrgId, sc.dashboardGuardianMock.OrgId)
					assert.Equal(t, cmd.UserId, sc.dashboardGuardianMock.User.UserId)
				})

			permissionScenario(t, "When updating a dashboard by existing id in the General folder, it should create dashboard guardian for dashboard with correct arguments and result in access denied error",
				canSave, func(t *testing.T, sc *permissionScenarioContext) {
					cmd := models.SaveDashboardCommand{
						OrgId: testOrgID,
						Dashboard: simplejson.NewFromAny(map[string]interface{}{
							"id":    sc.savedDashInGeneralFolder.Id,
							"title": "Dash",
						}),
						FolderId:  sc.savedDashInGeneralFolder.FolderId,
						UserId:    10000,
						Overwrite: true,
					}

					err := callSaveWithError(cmd, sc.sqlStore)
					assert.Equal(t, models.ErrDashboardUpdateAccessDenied, err)

					assert.Equal(t, sc.savedDashInGeneralFolder.Id, sc.dashboardGuardianMock.DashId)
					assert.Equal(t, cmd.OrgId, sc.dashboardGuardianMock.OrgId)
					assert.Equal(t, cmd.UserId, sc.dashboardGuardianMock.User.UserId)
				})

			permissionScenario(t, "When updating a dashboard by existing id in other folder, it should create dashboard guardian for dashboard with correct arguments and result in access denied error",
				canSave, func(t *testing.T, sc *permissionScenarioContext) {
					cmd := models.SaveDashboardCommand{
						OrgId: testOrgID,
						Dashboard: simplejson.NewFromAny(map[string]interface{}{
							"id":    sc.savedDashInFolder.Id,
							"title": "Dash",
						}),
						FolderId:  sc.savedDashInFolder.FolderId,
						UserId:    10000,
						Overwrite: true,
					}

					err := callSaveWithError(cmd, sc.sqlStore)
					require.Equal(t, models.ErrDashboardUpdateAccessDenied, err)

					assert.Equal(t, sc.savedDashInFolder.Id, sc.dashboardGuardianMock.DashId)
					assert.Equal(t, cmd.OrgId, sc.dashboardGuardianMock.OrgId)
					assert.Equal(t, cmd.UserId, sc.dashboardGuardianMock.User.UserId)
				})

			permissionScenario(t, "When moving a dashboard by existing ID to other folder from General folder, it should create dashboard guardian for other folder with correct arguments and result in access denied error",
				canSave, func(t *testing.T, sc *permissionScenarioContext) {
					cmd := models.SaveDashboardCommand{
						OrgId: testOrgID,
						Dashboard: simplejson.NewFromAny(map[string]interface{}{
							"id":    sc.savedDashInGeneralFolder.Id,
							"title": "Dash",
						}),
						FolderId:  sc.otherSavedFolder.Id,
						UserId:    10000,
						Overwrite: true,
					}

					err := callSaveWithError(cmd, sc.sqlStore)
					require.Equal(t, models.ErrDashboardUpdateAccessDenied, err)

					assert.Equal(t, sc.otherSavedFolder.Id, sc.dashboardGuardianMock.DashId)
					assert.Equal(t, cmd.OrgId, sc.dashboardGuardianMock.OrgId)
					assert.Equal(t, cmd.UserId, sc.dashboardGuardianMock.User.UserId)
				})

			permissionScenario(t, "When moving a dashboard by existing id to the General folder from other folder, it should create dashboard guardian for General folder with correct arguments and result in access denied error",
				canSave, func(t *testing.T, sc *permissionScenarioContext) {
					cmd := models.SaveDashboardCommand{
						OrgId: testOrgID,
						Dashboard: simplejson.NewFromAny(map[string]interface{}{
							"id":    sc.savedDashInFolder.Id,
							"title": "Dash",
						}),
						FolderId:  0,
						UserId:    10000,
						Overwrite: true,
					}

					err := callSaveWithError(cmd, sc.sqlStore)
					assert.Equal(t, models.ErrDashboardUpdateAccessDenied, err)

					assert.Equal(t, int64(0), sc.dashboardGuardianMock.DashId)
					assert.Equal(t, cmd.OrgId, sc.dashboardGuardianMock.OrgId)
					assert.Equal(t, cmd.UserId, sc.dashboardGuardianMock.User.UserId)
				})

			permissionScenario(t, "When moving a dashboard by existing uid to other folder from General folder, it should create dashboard guardian for other folder with correct arguments and result in access denied error",
				canSave, func(t *testing.T, sc *permissionScenarioContext) {
					cmd := models.SaveDashboardCommand{
						OrgId: testOrgID,
						Dashboard: simplejson.NewFromAny(map[string]interface{}{
							"uid":   sc.savedDashInGeneralFolder.Uid,
							"title": "Dash",
						}),
						FolderId:  sc.otherSavedFolder.Id,
						UserId:    10000,
						Overwrite: true,
					}

					err := callSaveWithError(cmd, sc.sqlStore)
					require.Equal(t, models.ErrDashboardUpdateAccessDenied, err)

					assert.Equal(t, sc.otherSavedFolder.Id, sc.dashboardGuardianMock.DashId)
					assert.Equal(t, cmd.OrgId, sc.dashboardGuardianMock.OrgId)
					assert.Equal(t, cmd.UserId, sc.dashboardGuardianMock.User.UserId)
				})

			permissionScenario(t, "When moving a dashboard by existing UID to the General folder from other folder, it should create dashboard guardian for General folder with correct arguments and result in access denied error",
				canSave, func(t *testing.T, sc *permissionScenarioContext) {
					cmd := models.SaveDashboardCommand{
						OrgId: testOrgID,
						Dashboard: simplejson.NewFromAny(map[string]interface{}{
							"uid":   sc.savedDashInFolder.Uid,
							"title": "Dash",
						}),
						FolderId:  0,
						UserId:    10000,
						Overwrite: true,
					}

					err := callSaveWithError(cmd, sc.sqlStore)
					require.Equal(t, models.ErrDashboardUpdateAccessDenied, err)

					assert.Equal(t, int64(0), sc.dashboardGuardianMock.DashId)
					assert.Equal(t, cmd.OrgId, sc.dashboardGuardianMock.OrgId)
					assert.Equal(t, cmd.UserId, sc.dashboardGuardianMock.User.UserId)
				})
		})

		t.Run("Given user has permission to save", func(t *testing.T) {
			const canSave = true

			t.Run("and overwrite flag is set to false", func(t *testing.T) {
				const shouldOverwrite = false

				permissionScenario(t, "When creating a dashboard in General folder with same name as dashboard in other folder",
					canSave, func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: testOrgID,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    nil,
								"title": sc.savedDashInFolder.Title,
							}),
							FolderId:  0,
							Overwrite: shouldOverwrite,
						}

						res := callSaveWithResult(t, cmd, sc.sqlStore)
						require.NotNil(t, res)

						dash, err := sc.sqlStore.GetDashboard(res.Id, cmd.OrgId, "", "")
						require.NoError(t, err)
						assert.Equal(t, res.Id, dash.Id)
						assert.Equal(t, int64(0), dash.FolderId)
					})

				permissionScenario(t, "When creating a dashboard in other folder with same name as dashboard in General folder",
					canSave, func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: testOrgID,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    nil,
								"title": sc.savedDashInGeneralFolder.Title,
							}),
							FolderId:  sc.savedFolder.Id,
							Overwrite: shouldOverwrite,
						}

						res := callSaveWithResult(t, cmd, sc.sqlStore)
						require.NotNil(t, res)

						assert.NotEqual(t, sc.savedDashInGeneralFolder.Id, res.Id)

						dash, err := sc.sqlStore.GetDashboard(res.Id, cmd.OrgId, "", "")
						require.NoError(t, err)
						assert.Equal(t, sc.savedFolder.Id, dash.FolderId)
					})

				permissionScenario(t, "When creating a folder with same name as dashboard in other folder",
					canSave, func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: testOrgID,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    nil,
								"title": sc.savedDashInFolder.Title,
							}),
							IsFolder:  true,
							Overwrite: shouldOverwrite,
						}

						res := callSaveWithResult(t, cmd, sc.sqlStore)
						require.NotNil(t, res)

						assert.NotEqual(t, sc.savedDashInGeneralFolder.Id, res.Id)
						assert.True(t, res.IsFolder)

						dash, err := sc.sqlStore.GetDashboard(res.Id, cmd.OrgId, "", "")
						require.NoError(t, err)
						assert.Equal(t, int64(0), dash.FolderId)
						assert.True(t, dash.IsFolder)
					})

				permissionScenario(t, "When saving a dashboard without id and uid and unique title in folder",
					canSave, func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: testOrgID,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"title": "Dash without id and uid",
							}),
							Overwrite: shouldOverwrite,
						}

						res := callSaveWithResult(t, cmd, sc.sqlStore)
						require.NotNil(t, res)

						assert.Greater(t, res.Id, int64(0))
						assert.NotEmpty(t, res.Uid)
						dash, err := sc.sqlStore.GetDashboard(res.Id, cmd.OrgId, "", "")
						require.NoError(t, err)
						assert.Equal(t, res.Id, dash.Id)
						assert.Equal(t, res.Uid, dash.Uid)
					})

				permissionScenario(t, "When saving a dashboard when dashboard id is zero ", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: testOrgID,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    0,
								"title": "Dash with zero id",
							}),
							Overwrite: shouldOverwrite,
						}

						res := callSaveWithResult(t, cmd, sc.sqlStore)
						require.NotNil(t, res)

						dash, err := sc.sqlStore.GetDashboard(res.Id, cmd.OrgId, "", "")
						require.NoError(t, err)
						assert.Equal(t, res.Id, dash.Id)
					})

				permissionScenario(t, "When saving a dashboard in non-existing folder", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: testOrgID,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"title": "Expect error",
							}),
							FolderId:  123412321,
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardFolderNotFound, err)
					})

				permissionScenario(t, "When updating an existing dashboard by id without current version", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    sc.savedDashInGeneralFolder.Id,
								"title": "test dash 23",
							}),
							FolderId:  sc.savedFolder.Id,
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardVersionMismatch, err)
					})

				permissionScenario(t, "When updating an existing dashboard by id with current version", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":      sc.savedDashInGeneralFolder.Id,
								"title":   "Updated title",
								"version": sc.savedDashInGeneralFolder.Version,
							}),
							FolderId:  sc.savedFolder.Id,
							Overwrite: shouldOverwrite,
						}

						res := callSaveWithResult(t, cmd, sc.sqlStore)
						require.NotNil(t, res)

						dash, err := sc.sqlStore.GetDashboard(sc.savedDashInGeneralFolder.Id, cmd.OrgId, "", "")
						require.NoError(t, err)
						assert.Equal(t, "Updated title", dash.Title)
						assert.Equal(t, sc.savedFolder.Id, dash.FolderId)
						assert.Greater(t, dash.Version, sc.savedDashInGeneralFolder.Version)
					})

				permissionScenario(t, "When updating an existing dashboard by uid without current version", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"uid":   sc.savedDashInFolder.Uid,
								"title": "test dash 23",
							}),
							FolderId:  0,
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardVersionMismatch, err)
					})

				permissionScenario(t, "When updating an existing dashboard by uid with current version", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"uid":     sc.savedDashInFolder.Uid,
								"title":   "Updated title",
								"version": sc.savedDashInFolder.Version,
							}),
							FolderId:  0,
							Overwrite: shouldOverwrite,
						}

						res := callSaveWithResult(t, cmd, sc.sqlStore)
						require.NotNil(t, res)

						dash, err := sc.sqlStore.GetDashboard(sc.savedDashInFolder.Id, cmd.OrgId, "", "")
						require.NoError(t, err)
						assert.Equal(t, "Updated title", dash.Title)
						assert.Equal(t, int64(0), dash.FolderId)
						assert.Greater(t, dash.Version, sc.savedDashInFolder.Version)
					})

				permissionScenario(t, "When creating a dashboard with same name as dashboard in other folder",
					canSave, func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: testOrgID,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    nil,
								"title": sc.savedDashInFolder.Title,
							}),
							FolderId:  sc.savedDashInFolder.FolderId,
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardWithSameNameInFolderExists, err)
					})

				permissionScenario(t, "When creating a dashboard with same name as dashboard in General folder",
					canSave, func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: testOrgID,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    nil,
								"title": sc.savedDashInGeneralFolder.Title,
							}),
							FolderId:  sc.savedDashInGeneralFolder.FolderId,
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardWithSameNameInFolderExists, err)
					})

				permissionScenario(t, "When creating a folder with same name as existing folder", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: testOrgID,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    nil,
								"title": sc.savedFolder.Title,
							}),
							IsFolder:  true,
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardWithSameNameInFolderExists, err)
					})
			})

			t.Run("and overwrite flag is set to true", func(t *testing.T) {
				const shouldOverwrite = true

				permissionScenario(t, "When updating an existing dashboard by id without current version", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    sc.savedDashInGeneralFolder.Id,
								"title": "Updated title",
							}),
							FolderId:  sc.savedFolder.Id,
							Overwrite: shouldOverwrite,
						}

						res := callSaveWithResult(t, cmd, sc.sqlStore)
						require.NotNil(t, res)

						dash, err := sc.sqlStore.GetDashboard(sc.savedDashInGeneralFolder.Id, cmd.OrgId, "", "")
						require.NoError(t, err)
						assert.Equal(t, "Updated title", dash.Title)
						assert.Equal(t, sc.savedFolder.Id, dash.FolderId)
						assert.Greater(t, dash.Version, sc.savedDashInGeneralFolder.Version)
					})

				permissionScenario(t, "When updating an existing dashboard by uid without current version", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"uid":   sc.savedDashInFolder.Uid,
								"title": "Updated title",
							}),
							FolderId:  0,
							Overwrite: shouldOverwrite,
						}

						res := callSaveWithResult(t, cmd, sc.sqlStore)
						require.NotNil(t, res)

						dash, err := sc.sqlStore.GetDashboard(sc.savedDashInFolder.Id, cmd.OrgId, "", "")
						require.NoError(t, err)
						assert.Equal(t, "Updated title", dash.Title)
						assert.Equal(t, int64(0), dash.FolderId)
						assert.Greater(t, dash.Version, sc.savedDashInFolder.Version)
					})

				permissionScenario(t, "When updating uid for existing dashboard using id", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    sc.savedDashInFolder.Id,
								"uid":   "new-uid",
								"title": sc.savedDashInFolder.Title,
							}),
							Overwrite: shouldOverwrite,
						}

						res := callSaveWithResult(t, cmd, sc.sqlStore)
						require.NotNil(t, res)
						assert.Equal(t, sc.savedDashInFolder.Id, res.Id)
						assert.Equal(t, "new-uid", res.Uid)

						dash, err := sc.sqlStore.GetDashboard(sc.savedDashInFolder.Id, cmd.OrgId, "", "")
						require.NoError(t, err)
						assert.Equal(t, "new-uid", dash.Uid)
						assert.Greater(t, dash.Version, sc.savedDashInFolder.Version)
					})

				permissionScenario(t, "When updating uid to an existing uid for existing dashboard using id", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    sc.savedDashInFolder.Id,
								"uid":   sc.savedDashInGeneralFolder.Uid,
								"title": sc.savedDashInFolder.Title,
							}),
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardWithSameUIDExists, err)
					})

				permissionScenario(t, "When creating a dashboard with same name as dashboard in other folder", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: testOrgID,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    nil,
								"title": sc.savedDashInFolder.Title,
							}),
							FolderId:  sc.savedDashInFolder.FolderId,
							Overwrite: shouldOverwrite,
						}

						res := callSaveWithResult(t, cmd, sc.sqlStore)
						require.NotNil(t, res)
						assert.Equal(t, sc.savedDashInFolder.Id, res.Id)
						assert.Equal(t, sc.savedDashInFolder.Uid, res.Uid)

						dash, err := sc.sqlStore.GetDashboard(res.Id, cmd.OrgId, "", "")
						require.NoError(t, err)
						assert.Equal(t, res.Id, dash.Id)
						assert.Equal(t, res.Uid, dash.Uid)
					})

				permissionScenario(t, "When creating a dashboard with same name as dashboard in General folder", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: testOrgID,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    nil,
								"title": sc.savedDashInGeneralFolder.Title,
							}),
							FolderId:  sc.savedDashInGeneralFolder.FolderId,
							Overwrite: shouldOverwrite,
						}

						res := callSaveWithResult(t, cmd, sc.sqlStore)
						require.NotNil(t, res)
						assert.Equal(t, sc.savedDashInGeneralFolder.Id, res.Id)
						assert.Equal(t, sc.savedDashInGeneralFolder.Uid, res.Uid)

						dash, err := sc.sqlStore.GetDashboard(res.Id, cmd.OrgId, "", "")
						require.NoError(t, err)
						assert.Equal(t, res.Id, dash.Id)
						assert.Equal(t, res.Uid, dash.Uid)
					})

				permissionScenario(t, "When updating existing folder to a dashboard using id", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    sc.savedFolder.Id,
								"title": "new title",
							}),
							IsFolder:  false,
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardTypeMismatch, err)
					})

				permissionScenario(t, "When updating existing dashboard to a folder using id", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"id":    sc.savedDashInFolder.Id,
								"title": "new folder title",
							}),
							IsFolder:  true,
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardTypeMismatch, err)
					})

				permissionScenario(t, "When updating existing folder to a dashboard using uid", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"uid":   sc.savedFolder.Uid,
								"title": "new title",
							}),
							IsFolder:  false,
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardTypeMismatch, err)
					})

				permissionScenario(t, "When updating existing dashboard to a folder using uid", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"uid":   sc.savedDashInFolder.Uid,
								"title": "new folder title",
							}),
							IsFolder:  true,
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardTypeMismatch, err)
					})

				permissionScenario(t, "When updating existing folder to a dashboard using title", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"title": sc.savedFolder.Title,
							}),
							IsFolder:  false,
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardWithSameNameAsFolder, err)
					})

				permissionScenario(t, "When updating existing dashboard to a folder using title", canSave,
					func(t *testing.T, sc *permissionScenarioContext) {
						cmd := models.SaveDashboardCommand{
							OrgId: 1,
							Dashboard: simplejson.NewFromAny(map[string]interface{}{
								"title": sc.savedDashInGeneralFolder.Title,
							}),
							IsFolder:  true,
							Overwrite: shouldOverwrite,
						}

						err := callSaveWithError(cmd, sc.sqlStore)
						assert.Equal(t, models.ErrDashboardFolderWithSameNameAsDashboard, err)
					})
			})
		})
	})
}

type permissionScenarioContext struct {
	dashboardGuardianMock    *guardian.FakeDashboardGuardian
	sqlStore                 *sqlstore.SQLStore
	savedFolder              *models.Dashboard
	savedDashInFolder        *models.Dashboard
	otherSavedFolder         *models.Dashboard
	savedDashInGeneralFolder *models.Dashboard
}

type permissionScenarioFunc func(t *testing.T, sc *permissionScenarioContext)

func permissionScenario(t *testing.T, desc string, canSave bool, fn permissionScenarioFunc) {
	t.Helper()

	mock := &guardian.FakeDashboardGuardian{
		CanSaveValue: canSave,
	}

	t.Run(desc, func(t *testing.T) {
		sqlStore := sqlstore.InitTestDB(t)

		savedFolder := saveTestFolder(t, "Saved folder", testOrgID, sqlStore)
		savedDashInFolder := saveTestDashboard(t, "Saved dash in folder", testOrgID, savedFolder.Id, sqlStore)
		saveTestDashboard(t, "Other saved dash in folder", testOrgID, savedFolder.Id, sqlStore)
		savedDashInGeneralFolder := saveTestDashboard(t, "Saved dashboard in general folder", testOrgID, 0, sqlStore)
		otherSavedFolder := saveTestFolder(t, "Other saved folder", testOrgID, sqlStore)

		require.Equal(t, "Saved folder", savedFolder.Title)
		require.Equal(t, "saved-folder", savedFolder.Slug)
		require.NotEqual(t, int64(0), savedFolder.Id)
		require.True(t, savedFolder.IsFolder)
		require.Equal(t, int64(0), savedFolder.FolderId)
		require.NotEmpty(t, savedFolder.Uid)

		require.Equal(t, "Saved dash in folder", savedDashInFolder.Title)
		require.Equal(t, "saved-dash-in-folder", savedDashInFolder.Slug)
		require.NotEqual(t, int64(0), savedDashInFolder.Id)
		require.False(t, savedDashInFolder.IsFolder)
		require.Equal(t, savedFolder.Id, savedDashInFolder.FolderId)
		require.NotEmpty(t, savedDashInFolder.Uid)

		origNewDashboardGuardian := guardian.New
		t.Cleanup(func() {
			guardian.New = origNewDashboardGuardian
		})
		guardian.MockDashboardGuardian(mock)

		sc := &permissionScenarioContext{
			dashboardGuardianMock:    mock,
			sqlStore:                 sqlStore,
			savedDashInFolder:        savedDashInFolder,
			otherSavedFolder:         otherSavedFolder,
			savedDashInGeneralFolder: savedDashInGeneralFolder,
			savedFolder:              savedFolder,
		}

		fn(t, sc)
	})
}

func callSaveWithResult(t *testing.T, cmd models.SaveDashboardCommand, sqlStore *sqlstore.SQLStore) *models.Dashboard {
	t.Helper()

	dto := toSaveDashboardDto(cmd)
	res, err := NewService(sqlStore).SaveDashboard(&dto, false)
	require.NoError(t, err)

	return res
}

func callSaveWithError(cmd models.SaveDashboardCommand, sqlStore *sqlstore.SQLStore) error {
	dto := toSaveDashboardDto(cmd)
	_, err := NewService(sqlStore).SaveDashboard(&dto, false)
	return err
}

func saveTestDashboard(t *testing.T, title string, orgID, folderID int64, sqlStore *sqlstore.SQLStore) *models.Dashboard {
	t.Helper()

	cmd := models.SaveDashboardCommand{
		OrgId:    orgID,
		FolderId: folderID,
		IsFolder: false,
		Dashboard: simplejson.NewFromAny(map[string]interface{}{
			"id":    nil,
			"title": title,
		}),
	}

	dto := SaveDashboardDTO{
		OrgId:     orgID,
		Dashboard: cmd.GetDashboardModel(),
		User: &models.SignedInUser{
			UserId:  1,
			OrgRole: models.ROLE_ADMIN,
		},
	}

	res, err := NewService(sqlStore).SaveDashboard(&dto, false)
	require.NoError(t, err)

	return res
}

func saveTestFolder(t *testing.T, title string, orgID int64, sqlStore *sqlstore.SQLStore) *models.Dashboard {
	t.Helper()
	cmd := models.SaveDashboardCommand{
		OrgId:    orgID,
		FolderId: 0,
		IsFolder: true,
		Dashboard: simplejson.NewFromAny(map[string]interface{}{
			"id":    nil,
			"title": title,
		}),
	}

	dto := SaveDashboardDTO{
		OrgId:     orgID,
		Dashboard: cmd.GetDashboardModel(),
		User: &models.SignedInUser{
			UserId:  1,
			OrgRole: models.ROLE_ADMIN,
		},
	}

	res, err := NewService(sqlStore).SaveDashboard(&dto, false)
	require.NoError(t, err)

	return res
}

func toSaveDashboardDto(cmd models.SaveDashboardCommand) SaveDashboardDTO {
	dash := (&cmd).GetDashboardModel()

	return SaveDashboardDTO{
		Dashboard: dash,
		Message:   cmd.Message,
		OrgId:     cmd.OrgId,
		User:      &models.SignedInUser{UserId: cmd.UserId},
		Overwrite: cmd.Overwrite,
	}
}
