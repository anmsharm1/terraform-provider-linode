//go:build unit

package databasevalkeyv2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCACertIsSensitive(t *testing.T) {
	resourceAttr, ok := frameworkResourceSchema.Attributes["ca_cert"]
	require.True(t, ok)
	assert.True(t, resourceAttr.IsSensitive())

	dataSourceAttr, ok := frameworkDatasourceSchema.Attributes["ca_cert"]
	require.True(t, ok)
	assert.True(t, dataSourceAttr.IsSensitive())
}
