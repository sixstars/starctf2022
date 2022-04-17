package models

import (
	"testing"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDashboardUrl(t *testing.T) {
	origAppURL := setting.AppUrl
	t.Cleanup(func() { setting.AppUrl = origAppURL })

	setting.AppUrl = ""
	url := GetDashboardUrl("uid", "my-dashboard")
	assert.Equal(t, "/d/uid/my-dashboard", url)
}

func TestGetFullDashboardUrl(t *testing.T) {
	origAppURL := setting.AppUrl
	t.Cleanup(func() { setting.AppUrl = origAppURL })

	setting.AppUrl = "http://grafana.local/"
	url := GetFullDashboardUrl("uid", "my-dashboard")
	assert.Equal(t, "http://grafana.local/d/uid/my-dashboard", url)
}

func TestDashboard_UpdateSlug(t *testing.T) {
	dashboard := NewDashboard("Grafana Play Home")
	assert.Equal(t, "grafana-play-home", dashboard.Slug)

	dashboard.UpdateSlug()
	assert.Equal(t, "grafana-play-home", dashboard.Slug)
}

func TestNewDashboardFromJson(t *testing.T) {
	json := simplejson.New()
	json.Set("title", "test dash")
	json.Set("tags", "")

	dash := NewDashboardFromJson(json)
	assert.Equal(t, "test dash", dash.Title)
	require.Empty(t, dash.GetTags())
}

func TestSaveDashboardCommand_GetDashboardModel(t *testing.T) {
	t.Run("should set IsFolder", func(t *testing.T) {
		json := simplejson.New()
		json.Set("title", "test dash")

		cmd := &SaveDashboardCommand{Dashboard: json, IsFolder: true}
		dash := cmd.GetDashboardModel()

		assert.Equal(t, "test dash", dash.Title)
		assert.True(t, dash.IsFolder)
	})

	t.Run("should set FolderId", func(t *testing.T) {
		json := simplejson.New()
		json.Set("title", "test dash")

		cmd := &SaveDashboardCommand{Dashboard: json, FolderId: 1}
		dash := cmd.GetDashboardModel()

		assert.Equal(t, int64(1), dash.FolderId)
	})
}

func TestSlugifyTitle(t *testing.T) {
	testCases := map[string]string{
		"Grafana Play Home": "grafana-play-home",
		"snöräv-över-ån":    "snorav-over-an",
		"漢字":                "han-zi",      // Hanzi for hanzi
		"🇦🇶":                "8J-HpvCfh7Y", // flag of Antarctica-emoji, using fallback
		"𒆠":                 "8JKGoA",      // cuneiform Ki, using fallback
	}

	for input, expected := range testCases {
		t.Run(input, func(t *testing.T) {
			slug := SlugifyTitle(input)
			assert.Equal(t, expected, slug)
		})
	}
}
