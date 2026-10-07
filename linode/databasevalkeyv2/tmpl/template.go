package tmpl

import (
	"testing"

	"github.com/linode/terraform-provider-linode/v4/linode/acceptance"
)

type TemplateData struct {
	Label    string
	Region   string
	EngineID string
	Type     string
}

func Basic(t testing.TB, label, region, engineID, nodeType string) string {
	return acceptance.ExecuteTemplate(t, "database_valkey_v2_basic", TemplateData{
		Label: label, Region: region, EngineID: engineID, Type: nodeType,
	})
}

func DataBasic(t testing.TB, label, region, engineID, nodeType string) string {
	return acceptance.ExecuteTemplate(t, "database_valkey_v2_data_basic", TemplateData{
		Label: label, Region: region, EngineID: engineID, Type: nodeType,
	})
}
