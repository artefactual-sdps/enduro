package event

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestNopServicePublishEvent(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	svc := NewServiceNop[string]()

	// Should not panic - nop service accepts any event.
	svc.PublishEvent(ctx, "test-event")
}

func TestNopServiceSubscribe(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	svc := NewServiceNop[string]()

	_, err := svc.Subscribe(ctx)
	assert.ErrorContains(t, err, "Subscribe not supported by nop service")
}
