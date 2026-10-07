//go:build unit

package databasevalkeyv2

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linode/linodego/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModel_FlattenAvailableRestoreTimes(t *testing.T) {
	timestamp := time.Date(2026, time.October, 5, 12, 30, 0, 0, time.UTC)
	tests := []struct {
		name           string
		availableTimes []time.Time
		wantNull       bool
		want           []string
	}{
		{name: "nil is null", wantNull: true},
		{name: "empty stays empty", availableTimes: []time.Time{}, want: []string{}},
		{name: "populated", availableTimes: []time.Time{timestamp}, want: []string{timestamp.Format(time.RFC3339)}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := &Model{}
			database := &linodego.ValkeyDatabase{AvailableRestoreTimes: test.availableTimes}

			diags := model.Flatten(t.Context(), database, nil, nil, false)
			require.False(t, diags.HasError(), "%v", diags)
			if test.wantNull {
				assert.True(t, model.AvailableRestoreTimes.IsNull())
				return
			}

			var got []string
			require.False(t, model.AvailableRestoreTimes.ElementsAs(t.Context(), &got, false).HasError())
			assert.Equal(t, test.want, got)
		})
	}
}

func TestModel_GetFork(t *testing.T) {
	restoreTime := time.Date(2026, time.October, 5, 12, 30, 0, 0, time.UTC)
	tests := []struct {
		name       string
		model      Model
		wantNil    bool
		wantSource int
		wantTime   *time.Time
	}{
		{name: "no fork inputs", model: Model{}, wantNil: true},
		{name: "source only", model: Model{ForkSource: types.Int64Value(123)}, wantSource: 123},
		{
			name:       "source and snapshot",
			model:      Model{ForkSource: types.Int64Value(123), ForkRestoreTime: timetypes.NewRFC3339TimeValue(restoreTime)},
			wantSource: 123,
			wantTime:   &restoreTime,
		},
		{name: "snapshot requires source", model: Model{ForkRestoreTime: timetypes.NewRFC3339TimeValue(restoreTime)}, wantNil: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var diags diag.Diagnostics
			fork := test.model.GetFork(&diags)
			if test.name == "snapshot requires source" {
				require.True(t, diags.HasError())
				assert.Nil(t, fork)
				return
			}
			require.False(t, diags.HasError(), "%v", diags)
			if test.wantNil {
				assert.Nil(t, fork)
				return
			}
			require.NotNil(t, fork)
			assert.Equal(t, test.wantSource, fork.Source)
			if test.wantTime == nil {
				assert.Nil(t, fork.RestoreTime)
			} else {
				require.NotNil(t, fork.RestoreTime)
				assert.True(t, test.wantTime.Equal(*fork.RestoreTime))
			}
		})
	}
}

func TestModel_GetEngineConfigPreservesExplicitZeroValuesAndOmitsNull(t *testing.T) {
	model := Model{
		EngineConfigBackupHour:            types.Int64Value(0),
		EngineConfigFrequentSnapshots:     types.BoolValue(false),
		EngineConfigValkeyMaxmemoryPolicy: types.StringNull(),
	}
	var diags diag.Diagnostics
	config := model.GetEngineConfig(&diags)
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, config.BackupHour)
	assert.Equal(t, 0, *config.BackupHour)
	require.NotNil(t, config.FrequentSnapshots)
	assert.False(t, *config.FrequentSnapshots)
	assert.Nil(t, config.ValkeyMaxmemoryPolicy)

	encoded, err := json.Marshal(config)
	require.NoError(t, err)
	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(encoded, &fields))
	assert.JSONEq(t, "0", string(fields["backup_hour"]))
	assert.JSONEq(t, "false", string(fields["frequent_snapshots"]))
	assert.NotContains(t, fields, "valkey_maxmemory_policy")
}

func TestModel_GetEngineConfigOmitsUnsetConfig(t *testing.T) {
	var diags diag.Diagnostics
	assert.Nil(t, (&Model{}).GetEngineConfig(&diags))
	assert.False(t, diags.HasError(), "%v", diags)
}

func TestModel_CopyFromCopiesRestoreTimes(t *testing.T) {
	list, diags := types.ListValueFrom(t.Context(), types.StringType, []string{"2026-10-05T12:30:00Z"})
	require.False(t, diags.HasError())
	original := Model{AvailableRestoreTimes: list}
	var copied Model
	copied.CopyFrom(&original, false)
	assert.Equal(t, original.AvailableRestoreTimes, copied.AvailableRestoreTimes)
}
