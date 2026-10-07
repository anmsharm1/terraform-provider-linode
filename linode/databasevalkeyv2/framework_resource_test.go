//go:build integration || databasevalkeyv2

package databasevalkeyv2_test

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/linode/linodego/v2"
	"github.com/linode/terraform-provider-linode/v4/linode/acceptance"
	"github.com/linode/terraform-provider-linode/v4/linode/databasevalkeyv2/tmpl"
	"github.com/linode/terraform-provider-linode/v4/linode/helper/databaseshared"
)

var testRegion, testEngine string

func init() {
	resource.AddTestSweepers("linode_database_valkey_v2", &resource.Sweeper{
		Name: "linode_database_valkey_v2",
		F:    sweep,
	})

	client, err := acceptance.GetTestClient()
	if err != nil {
		log.Fatal(err)
	}

	region, err := acceptance.GetRandomRegionWithCaps([]linodego.RegionCapability{linodego.CapabilityDBAAS}, "core")
	if err != nil {
		log.Fatal(err)
	}
	testRegion = region

	engine, err := databaseshared.ResolveValidDBEngine(context.Background(), *client, string(linodego.DatabaseEngineTypeValkey))
	if err != nil {
		log.Fatal(err)
	}
	testEngine = engine.ID
}

func sweep(prefix string) error {
	client, err := acceptance.GetTestClient()
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	databases, err := client.ListValkeyDatabases(context.Background(), acceptance.SweeperListOptions(prefix, "label"))
	if err != nil {
		return fmt.Errorf("failed to list Valkey databases: %w", err)
	}
	for _, database := range databases {
		if !acceptance.ShouldSweep(prefix, database.Label) {
			continue
		}
		if err := client.DeleteValkeyDatabase(context.Background(), database.ID); err != nil {
			return fmt.Errorf("failed to delete Valkey database %s during sweep: %w", database.Label, err)
		}
	}
	return nil
}

func checkValkeyDatabaseDestroy(state *terraform.State) error {
	client, err := acceptance.GetTestClient()
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}
	for _, instance := range state.RootModule().Resources {
		if instance.Type != "linode_database_valkey_v2" {
			continue
		}
		id, err := strconv.Atoi(instance.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to parse database ID %q: %w", instance.Primary.ID, err)
		}
		if _, err := client.GetValkeyDatabase(context.Background(), id); err == nil || !linodego.IsNotFound(err) {
			return fmt.Errorf("Valkey database %d still exists or returned an unexpected error: %v", id, err)
		}
	}
	return nil
}

func TestAccResourceDatabaseValkeyV2_basic(t *testing.T) {
	t.Parallel()

	const resourceName = "linode_database_valkey_v2.foobar"
	label := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: acceptance.ProtoV6ProviderFactories,
		CheckDestroy:             checkValkeyDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.Basic(t, label, testRegion, testEngine, "g7-dedicated-4-2"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("label"), knownvalue.StringExact(label)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("engine_id"), knownvalue.StringExact(testEngine)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("status"), knownvalue.StringExact("active")),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated", "members", "available_restore_times"},
			},
		},
	})
}
