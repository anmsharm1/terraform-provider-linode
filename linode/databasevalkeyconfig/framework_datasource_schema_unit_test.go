//go:build unit

package databasevalkeyconfig

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestConfigOptionSchema(t *testing.T) {
	intFields := func(valueField string) map[string]attr.Type {
		return map[string]attr.Type{
			valueField: types.Int64Type,
			"minimum":  types.Int64Type,
			"maximum":  types.Int64Type,
		}
	}
	options := map[string]map[string]attr.Type{
		"backup_hour":                              intFields("example"),
		"backup_minute":                            intFields("example"),
		"frequent_snapshots":                       {"default": types.BoolType},
		"valkey_acl_channels_default":              {"enum": types.ListType{ElemType: types.StringType}},
		"valkey_active_expire_effort":              intFields("default"),
		"valkey_activedefrag":                      {"default": types.BoolType},
		"valkey_lfu_decay_time":                    intFields("default"),
		"valkey_lfu_log_factor":                    intFields("default"),
		"valkey_maxmemory_policy":                  {"default": types.StringType, "enum": types.ListType{ElemType: types.StringType}},
		"valkey_number_of_databases":               intFields("example"),
		"valkey_persistence":                       {"enum": types.ListType{ElemType: types.StringType}},
		"valkey_pubsub_client_output_buffer_limit": intFields("example"),
		"valkey_timeout":                           intFields("default"),
	}

	require.Len(t, frameworkDataSourceSchema.Attributes, len(options)+1)
	for name, fields := range options {
		t.Run(name, func(t *testing.T) {
			attribute, ok := frameworkDataSourceSchema.Attributes[name].(schema.ListAttribute)
			require.True(t, ok)
			require.True(t, attribute.Computed)
			require.NotEmpty(t, attribute.Description)

			fields["description"] = types.StringType
			fields["requires_restart"] = types.BoolType
			fields["type"] = types.ListType{ElemType: types.StringType}
			require.Equal(t, types.ObjectType{AttrTypes: fields}, attribute.ElementType)
		})
	}
}
