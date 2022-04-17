package libraryelements

import (
	"testing"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"gopkg.in/macaron.v1"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/models"
)

func TestGetLibraryElement(t *testing.T) {
	scenarioWithPanel(t, "When an admin tries to get a library panel that does not exist, it should fail",
		func(t *testing.T, sc scenarioContext) {
			// by uid
			sc.ctx.Req = macaron.SetURLParams(sc.ctx.Req, map[string]string{":uid": "unknown"})
			resp := sc.service.getHandler(sc.reqContext)
			require.Equal(t, 404, resp.Status())

			// by name
			sc.ctx.Req = macaron.SetURLParams(sc.ctx.Req, map[string]string{":name": "unknown"})
			resp = sc.service.getByNameHandler(sc.reqContext)
			require.Equal(t, 404, resp.Status())
		})

	scenarioWithPanel(t, "When an admin tries to get a library panel that exists, it should succeed and return correct result",
		func(t *testing.T, sc scenarioContext) {
			var expected = func(res libraryElementResult) libraryElementResult {
				return libraryElementResult{
					Result: libraryElement{
						ID:          1,
						OrgID:       1,
						FolderID:    1,
						UID:         res.Result.UID,
						Name:        "Text - Library Panel",
						Kind:        int64(models.PanelElement),
						Type:        "text",
						Description: "A description",
						Model: map[string]interface{}{
							"datasource":  "${DS_GDEV-TESTDATA}",
							"description": "A description",
							"id":          float64(1),
							"title":       "Text - Library Panel",
							"type":        "text",
						},
						Version: 1,
						Meta: LibraryElementDTOMeta{
							FolderName:          "ScenarioFolder",
							FolderUID:           sc.folder.Uid,
							ConnectedDashboards: 0,
							Created:             res.Result.Meta.Created,
							Updated:             res.Result.Meta.Updated,
							CreatedBy: LibraryElementDTOMetaUser{
								ID:        1,
								Name:      userInDbName,
								AvatarURL: userInDbAvatar,
							},
							UpdatedBy: LibraryElementDTOMetaUser{
								ID:        1,
								Name:      userInDbName,
								AvatarURL: userInDbAvatar,
							},
						},
					},
				}
			}

			// by uid
			sc.ctx.Req = macaron.SetURLParams(sc.ctx.Req, map[string]string{":uid": sc.initialResult.Result.UID})
			resp := sc.service.getHandler(sc.reqContext)
			var result = validateAndUnMarshalResponse(t, resp)

			if diff := cmp.Diff(expected(result), result, getCompareOptions()...); diff != "" {
				t.Fatalf("Result mismatch (-want +got):\n%s", diff)
			}

			// by name
			sc.ctx.Req = macaron.SetURLParams(sc.ctx.Req, map[string]string{":name": sc.initialResult.Result.Name})
			resp = sc.service.getByNameHandler(sc.reqContext)
			arrayResult := validateAndUnMarshalArrayResponse(t, resp)

			if diff := cmp.Diff(libraryElementArrayResult{Result: []libraryElement{expected(result).Result}}, arrayResult, getCompareOptions()...); diff != "" {
				t.Fatalf("Result mismatch (-want +got):\n%s", diff)
			}
		})

	scenarioWithPanel(t, "When an admin tries to get a connected library panel, it should succeed and return correct connected dashboards",
		func(t *testing.T, sc scenarioContext) {
			dashJSON := map[string]interface{}{
				"panels": []interface{}{
					map[string]interface{}{
						"id": int64(1),
						"gridPos": map[string]interface{}{
							"h": 6,
							"w": 6,
							"x": 0,
							"y": 0,
						},
					},
					map[string]interface{}{
						"id": int64(2),
						"gridPos": map[string]interface{}{
							"h": 6,
							"w": 6,
							"x": 6,
							"y": 0,
						},
						"libraryPanel": map[string]interface{}{
							"uid":  sc.initialResult.Result.UID,
							"name": sc.initialResult.Result.Name,
						},
					},
				},
			}
			dash := models.Dashboard{
				Title: "Testing getHandler",
				Data:  simplejson.NewFromAny(dashJSON),
			}
			dashInDB := createDashboard(t, sc.sqlStore, sc.user, &dash, sc.folder.Id)
			err := sc.service.ConnectElementsToDashboard(sc.reqContext, []string{sc.initialResult.Result.UID}, dashInDB.Id)
			require.NoError(t, err)

			expected := func(res libraryElementResult) libraryElementResult {
				return libraryElementResult{
					Result: libraryElement{
						ID:          1,
						OrgID:       1,
						FolderID:    1,
						UID:         res.Result.UID,
						Name:        "Text - Library Panel",
						Kind:        int64(models.PanelElement),
						Type:        "text",
						Description: "A description",
						Model: map[string]interface{}{
							"datasource":  "${DS_GDEV-TESTDATA}",
							"description": "A description",
							"id":          float64(1),
							"title":       "Text - Library Panel",
							"type":        "text",
						},
						Version: 1,
						Meta: LibraryElementDTOMeta{
							FolderName:          "ScenarioFolder",
							FolderUID:           sc.folder.Uid,
							ConnectedDashboards: 1,
							Created:             res.Result.Meta.Created,
							Updated:             res.Result.Meta.Updated,
							CreatedBy: LibraryElementDTOMetaUser{
								ID:        1,
								Name:      userInDbName,
								AvatarURL: userInDbAvatar,
							},
							UpdatedBy: LibraryElementDTOMetaUser{
								ID:        1,
								Name:      userInDbName,
								AvatarURL: userInDbAvatar,
							},
						},
					},
				}
			}

			// by uid
			sc.ctx.Req = macaron.SetURLParams(sc.ctx.Req, map[string]string{":uid": sc.initialResult.Result.UID})
			resp := sc.service.getHandler(sc.reqContext)
			result := validateAndUnMarshalResponse(t, resp)

			if diff := cmp.Diff(expected(result), result, getCompareOptions()...); diff != "" {
				t.Fatalf("Result mismatch (-want +got):\n%s", diff)
			}

			// by name
			sc.ctx.Req = macaron.SetURLParams(sc.ctx.Req, map[string]string{":name": sc.initialResult.Result.Name})
			resp = sc.service.getByNameHandler(sc.reqContext)
			arrayResult := validateAndUnMarshalArrayResponse(t, resp)
			if diff := cmp.Diff(libraryElementArrayResult{Result: []libraryElement{expected(result).Result}}, arrayResult, getCompareOptions()...); diff != "" {
				t.Fatalf("Result mismatch (-want +got):\n%s", diff)
			}
		})

	scenarioWithPanel(t, "When an admin tries to get a library panel that exists in an other org, it should fail",
		func(t *testing.T, sc scenarioContext) {
			sc.reqContext.SignedInUser.OrgId = 2
			sc.reqContext.SignedInUser.OrgRole = models.ROLE_ADMIN

			// by uid
			sc.ctx.Req = macaron.SetURLParams(sc.ctx.Req, map[string]string{":uid": sc.initialResult.Result.UID})
			resp := sc.service.getHandler(sc.reqContext)
			require.Equal(t, 404, resp.Status())

			// by name
			sc.ctx.Req = macaron.SetURLParams(sc.ctx.Req, map[string]string{":name": sc.initialResult.Result.Name})
			resp = sc.service.getByNameHandler(sc.reqContext)
			require.Equal(t, 404, resp.Status())
		})
}
