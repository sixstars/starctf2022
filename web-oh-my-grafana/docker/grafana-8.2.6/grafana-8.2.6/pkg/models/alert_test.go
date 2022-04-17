package models

import (
	"testing"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlert_ContainsUpdates(t *testing.T) {
	settings, err := simplejson.NewJson([]byte(`{ "field": "value" }`))
	require.NoError(t, err)

	alert1 := &Alert{
		Settings: settings,
		Name:     "Name",
		Message:  "Message",
	}

	alert2 := &Alert{
		Settings: settings,
		Name:     "Name",
		Message:  "Message",
	}

	assert.False(t, alert1.ContainsUpdates(alert2))

	settingsUpdated, err := simplejson.NewJson([]byte(`{ "field": "newValue" }`))
	require.NoError(t, err)

	alert2.Settings = settingsUpdated

	assert.True(t, alert1.ContainsUpdates(alert2))
}

func TestAlert_GetTagsFromSettings(t *testing.T) {
	settings, err := simplejson.NewJson([]byte(`{
		"field": "value",
		"alertRuleTags": {
			"foo": "bar",
			"waldo": "fred",
			"tagMap": { "mapValue": "value" }
		}
	}`))
	require.NoError(t, err)

	alert := &Alert{
		Settings: settings,
		Name:     "Name",
		Message:  "Message",
	}

	expectedTags := []*Tag{
		{Id: 0, Key: "foo", Value: "bar"},
		{Id: 0, Key: "waldo", Value: "fred"},
		{Id: 0, Key: "tagMap", Value: ""},
	}
	actualTags := alert.GetTagsFromSettings()

	assert.ElementsMatch(t, actualTags, expectedTags)
}
