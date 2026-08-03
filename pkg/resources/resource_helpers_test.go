package resources

import (
	"errors"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/require"
)

type stubDroppable struct {
	err     error
	dropped bool
}

func (s *stubDroppable) Drop() error {
	s.dropped = true
	return s.err
}

func TestRollbackCreate(t *testing.T) {
	o := materialize.MaterializeObject{ObjectType: materialize.Table, Name: "table"}
	cause := errors.New("could not set comment")

	t.Run("drop succeeds", func(t *testing.T) {
		b := &stubDroppable{}
		diags := rollbackCreate(b, o, "comment", cause)

		require.True(t, b.dropped, "The object should be dropped")
		require.Len(t, diags, 1, "Only the original error should be reported")
		require.Equal(t, diag.Error, diags[0].Severity)
	})

	t.Run("drop fails", func(t *testing.T) {
		b := &stubDroppable{err: errors.New("still referenced")}
		diags := rollbackCreate(b, o, "comment", cause)

		require.Len(t, diags, 2, "A failed drop should add a warning")
		require.Equal(t, diag.Error, diags[0].Severity)
		require.Equal(t, diag.Warning, diags[1].Severity)
		require.Contains(t, diags[1].Detail, "still referenced")
	})
}
