package animation

import (
	"bytes"
	"testing"

	appcontext "github.com/svvly2/lettuce/internal/app/context"
)

func TestNewUploadHandlerRequiresEitherClientOrOAuth(t *testing.T) {
	ctx := &appcontext.Context{}

	if _, err := newUploadHandler(ctx, nil, "name", "", bytes.NewBufferString("data"), 0, false); err == nil {
		t.Fatal("expected an error when neither a cookie client nor OAuth client is available")
	}
}
