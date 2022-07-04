package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateOrUpdate(t *testing.T) {
	wrapper := NamespaceWrapper{}
	sql, args := wrapper.CreateOrUpdate()
	require.Equal(t, "asdf", sql)
	require.Equal(t, []interface{}{"asdf"}, args)
}
