//go:build unit

package databaseaccesscontrols

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessControlsTemplatesRenderAllowList(t *testing.T) {
	tmpl, err := template.ParseFiles("tmpl/mysql.gotf", "tmpl/postgresql.gotf", "tmpl/valkey.gotf")
	require.NoError(t, err)

	data := struct {
		Engine, Label, Region, AllowedIP string
	}{
		Engine: "engine/version", Label: "test-db", Region: "us-mia", AllowedIP: "192.168.0.25/32",
	}
	for _, name := range []string{
		"database_access_controls_mysql",
		"database_access_controls_postgresql",
		"database_access_controls_valkey",
	} {
		t.Run(name, func(t *testing.T) {
			var config bytes.Buffer
			require.NoError(t, tmpl.ExecuteTemplate(&config, name, data))
			assert.Contains(t, config.String(), `allow_list = ["192.168.0.25/32"]`)
		})
	}
}
