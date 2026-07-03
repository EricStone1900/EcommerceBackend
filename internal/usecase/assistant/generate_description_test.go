package assistant

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockAssistant_GenerateDescription_Success(t *testing.T) {
	a := newTestAssistant()

	ctx := context.Background()
	desc, err := a.GenerateProductDescription(ctx, 42)

	require.NoError(t, err)
	assert.Contains(t, desc, "42")
	assert.Contains(t, desc, "AI generated description")
}

func TestMockAssistant_GenerateDescription_ContextTimeout(t *testing.T) {
	a := newTestAssistant()

	// Create a context that expires before the 50ms simulated latency
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Wait for the timeout to actually fire before calling
	<-ctx.Done()

	desc, err := a.GenerateProductDescription(ctx, 1)

	require.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Empty(t, desc)
}

func TestMockAssistant_GenerateDescription_ContextCancelled(t *testing.T) {
	a := newTestAssistant()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	desc, err := a.GenerateProductDescription(ctx, 1)

	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Empty(t, desc)
}

func TestMockAssistant_GenerateDescription_DifferentProductIDs(t *testing.T) {
	a := newTestAssistant()

	testIDs := []uint{1, 100, 9999}
	for _, id := range testIDs {
		desc, err := a.GenerateProductDescription(context.Background(), id)
		require.NoError(t, err)
		assert.Contains(t, desc, strconv.FormatUint(uint64(id), 10))
	}
}
