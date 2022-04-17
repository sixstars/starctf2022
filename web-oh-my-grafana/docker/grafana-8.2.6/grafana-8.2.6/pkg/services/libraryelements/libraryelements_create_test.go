package libraryelements

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/util"
)

func TestCreateLibraryElement(t *testing.T) {
	scenarioWithPanel(t, "When an admin tries to create a library panel that already exists, it should fail",
		func(t *testing.T, sc scenarioContext) {
			command := getCreatePanelCommand(sc.folder.Id, "Text - Library Panel")
			resp := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 400, resp.Status())
		})

	scenarioWithPanel(t, "When an admin tries to create a library panel that does not exists, it should succeed",
		func(t *testing.T, sc scenarioContext) {
			var expected = libraryElementResult{
				Result: libraryElement{
					ID:          1,
					OrgID:       1,
					FolderID:    1,
					UID:         sc.initialResult.Result.UID,
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
						ConnectedDashboards: 0,
						Created:             sc.initialResult.Result.Meta.Created,
						Updated:             sc.initialResult.Result.Meta.Updated,
						CreatedBy: LibraryElementDTOMetaUser{
							ID:        1,
							Name:      "signed_in_user",
							AvatarURL: "/avatar/37524e1eb8b3e32850b57db0a19af93b",
						},
						UpdatedBy: LibraryElementDTOMetaUser{
							ID:        1,
							Name:      "signed_in_user",
							AvatarURL: "/avatar/37524e1eb8b3e32850b57db0a19af93b",
						},
					},
				},
			}
			if diff := cmp.Diff(expected, sc.initialResult, getCompareOptions()...); diff != "" {
				t.Fatalf("Result mismatch (-want +got):\n%s", diff)
			}
		})

	testScenario(t, "When an admin tries to create a library panel that does not exists using an nonexistent UID, it should succeed",
		func(t *testing.T, sc scenarioContext) {
			command := getCreatePanelCommand(sc.folder.Id, "Nonexistent UID")
			command.UID = util.GenerateShortUID()
			resp := sc.service.createHandler(sc.reqContext, command)
			var result = validateAndUnMarshalResponse(t, resp)
			var expected = libraryElementResult{
				Result: libraryElement{
					ID:          1,
					OrgID:       1,
					FolderID:    1,
					UID:         command.UID,
					Name:        "Nonexistent UID",
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
						ConnectedDashboards: 0,
						Created:             result.Result.Meta.Created,
						Updated:             result.Result.Meta.Updated,
						CreatedBy: LibraryElementDTOMetaUser{
							ID:        1,
							Name:      "signed_in_user",
							AvatarURL: "/avatar/37524e1eb8b3e32850b57db0a19af93b",
						},
						UpdatedBy: LibraryElementDTOMetaUser{
							ID:        1,
							Name:      "signed_in_user",
							AvatarURL: "/avatar/37524e1eb8b3e32850b57db0a19af93b",
						},
					},
				},
			}
			if diff := cmp.Diff(expected, result, getCompareOptions()...); diff != "" {
				t.Fatalf("Result mismatch (-want +got):\n%s", diff)
			}
		})

	scenarioWithPanel(t, "When an admin tries to create a library panel that does not exists using an existent UID, it should fail",
		func(t *testing.T, sc scenarioContext) {
			command := getCreatePanelCommand(sc.folder.Id, "Existing UID")
			command.UID = sc.initialResult.Result.UID
			resp := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 400, resp.Status())
		})

	scenarioWithPanel(t, "When an admin tries to create a library panel that does not exists using an invalid UID, it should fail",
		func(t *testing.T, sc scenarioContext) {
			command := getCreatePanelCommand(sc.folder.Id, "Invalid UID")
			command.UID = "Testing an invalid UID"
			resp := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 400, resp.Status())
		})

	scenarioWithPanel(t, "When an admin tries to create a library panel that does not exists using an UID that is too long, it should fail",
		func(t *testing.T, sc scenarioContext) {
			command := getCreatePanelCommand(sc.folder.Id, "Invalid UID")
			command.UID = "j6T00KRZzj6T00KRZzj6T00KRZzj6T00KRZzj6T00K"
			resp := sc.service.createHandler(sc.reqContext, command)
			require.Equal(t, 400, resp.Status())
		})

	testScenario(t, "When an admin tries to create a library panel where name and panel title differ, it should not update panel title",
		func(t *testing.T, sc scenarioContext) {
			command := getCreatePanelCommand(1, "Library Panel Name")
			resp := sc.service.createHandler(sc.reqContext, command)
			var result = validateAndUnMarshalResponse(t, resp)
			var expected = libraryElementResult{
				Result: libraryElement{
					ID:          1,
					OrgID:       1,
					FolderID:    1,
					UID:         result.Result.UID,
					Name:        "Library Panel Name",
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
						ConnectedDashboards: 0,
						Created:             result.Result.Meta.Created,
						Updated:             result.Result.Meta.Updated,
						CreatedBy: LibraryElementDTOMetaUser{
							ID:        1,
							Name:      "signed_in_user",
							AvatarURL: "/avatar/37524e1eb8b3e32850b57db0a19af93b",
						},
						UpdatedBy: LibraryElementDTOMetaUser{
							ID:        1,
							Name:      "signed_in_user",
							AvatarURL: "/avatar/37524e1eb8b3e32850b57db0a19af93b",
						},
					},
				},
			}
			if diff := cmp.Diff(expected, result, getCompareOptions()...); diff != "" {
				t.Fatalf("Result mismatch (-want +got):\n%s", diff)
			}
		})
}
