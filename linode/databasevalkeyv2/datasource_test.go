//go:build integration || databasevalkeyv2

package databasevalkeyv2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/linode/terraform-provider-linode/v4/linode/acceptance"
	"github.com/linode/terraform-provider-linode/v4/linode/databasevalkeyv2/tmpl"
)

func TestAccDataSourceDatabaseValkeyV2_basic(t *testing.T) {
	t.Parallel()

	const dataSourceName = "data.linode_database_valkey_v2.foobar"
	label := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: acceptance.ProtoV6ProviderFactories,
		CheckDestroy:             checkValkeyDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.DataBasic(t, label, testRegion, testEngine, "g7-dedicated-4-2"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("label"), knownvalue.StringExact(label)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("engine_id"), knownvalue.StringExact(testEngine)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("status"), knownvalue.StringExact("active")),
				},
			},
		},
	})
}
