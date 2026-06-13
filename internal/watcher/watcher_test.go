package watcher_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr/testr"
	"go.opentelemetry.io/otel/trace/noop"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/watcher"
)

func TestNewAllowsEmptyConfig(t *testing.T) {
	t.Parallel()

	svc, err := watcher.New(
		context.Background(),
		noop.NewTracerProvider(),
		testr.New(t),
		&watcher.Config{},
	)

	assert.NilError(t, err)
	assert.Assert(t, svc != nil)
	assert.DeepEqual(t, svc.Watchers(), []watcher.Watcher{})
}
