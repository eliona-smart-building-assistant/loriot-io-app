package main

import (
	"github.com/eliona-smart-building-assistant/app-integration-tests/app"
	"github.com/eliona-smart-building-assistant/app-integration-tests/assert"
	"github.com/eliona-smart-building-assistant/app-integration-tests/test"
	"testing"
)

func TestApp(t *testing.T) {
	app.StartApp()
	test.AppWorks(t)
	t.Run("TestSchema", schema)
	t.Run("TestAssetTypes", assetTypes)
	app.StopApp()
}

func schema(t *testing.T) {
	t.Parallel()

	assert.SchemaExists(t, "loriot_io", []string{"configuration", "asset"})
}

func assetTypes(t *testing.T) {
	t.Parallel()

	assert.AssetTypeExists(t, "loriot_io_root", []string{})
}
