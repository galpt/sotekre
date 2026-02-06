package docs_test

import (
	"testing"

	"github.com/galpt/sotekre/backend/docs"
	"github.com/stretchr/testify/require"
)

func TestDocsPackageInitialization(t *testing.T) {
	require.NotNil(t, docs.SwaggerInfo, "generated docs should be present")
	require.NotEmpty(t, docs.SwaggerInfo.InstanceName(), "Swagger instance name should not be empty")
}