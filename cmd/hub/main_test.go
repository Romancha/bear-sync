package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_MissingRequired(t *testing.T) {
	for _, key := range []string{"HUB_DB_PATH", "HUB_OPENCLAW_TOKEN", "HUB_BRIDGE_TOKEN"} {
		t.Run(key, func(t *testing.T) {
			required := map[string]string{
				"HUB_DB_PATH":        "/tmp/test.db",
				"HUB_OPENCLAW_TOKEN": "oc-test",
				"HUB_BRIDGE_TOKEN":   "br-test",
			}
			for k, v := range required {
				if k != key {
					t.Setenv(k, v)
				}
			}
			_, err := loadConfig()
			require.Error(t, err)
			assert.Contains(t, err.Error(), key)
		})
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("HUB_DB_PATH", "/tmp/test.db")
	t.Setenv("HUB_OPENCLAW_TOKEN", "oc-test")
	t.Setenv("HUB_BRIDGE_TOKEN", "br-test")

	cfg, err := loadConfig()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.port)
	assert.Equal(t, "attachments", cfg.attachmentsDir)
	assert.Equal(t, "oc-test", cfg.openclawToken)
	assert.Equal(t, "br-test", cfg.bridgeToken)
}

func TestLoadConfig_CustomValues(t *testing.T) {
	t.Setenv("HUB_DB_PATH", "/tmp/hub.db")
	t.Setenv("HUB_OPENCLAW_TOKEN", "oc-test")
	t.Setenv("HUB_BRIDGE_TOKEN", "br-test")
	t.Setenv("HUB_PORT", "9090")
	t.Setenv("HUB_ATTACHMENTS_DIR", "/tmp/att")

	cfg, err := loadConfig()
	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.port)
	assert.Equal(t, "/tmp/att", cfg.attachmentsDir)
	assert.Equal(t, "/tmp/hub.db", cfg.dbPath)
}
