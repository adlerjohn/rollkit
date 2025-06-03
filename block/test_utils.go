package block

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"testing"

	"cosmossdk.io/log"
	"github.com/rollkit/rollkit/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// GenerateHeaderHash creates a deterministic hash for a test header based on height and proposer.
// This is useful for predicting expected hashes in tests without needing full header construction.
func GenerateHeaderHash(t *testing.T, height uint64, proposer []byte) []byte {
	t.Helper()
	// Create a simple deterministic representation of the header's identity
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, height)

	hasher := sha256.New()
	_, err := hasher.Write([]byte("testheader:")) // Prefix to avoid collisions
	require.NoError(t, err)
	_, err = hasher.Write(heightBytes)
	require.NoError(t, err)
	_, err = hasher.Write(proposer)
	require.NoError(t, err)

	return hasher.Sum(nil)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, keyvals ...any) { m.Called(msg, keyvals) }
func (m *MockLogger) Info(msg string, keyvals ...any)  { m.Called(msg, keyvals) }
func (m *MockLogger) Warn(msg string, keyvals ...any)  { m.Called(msg, keyvals) }
func (m *MockLogger) Error(msg string, keyvals ...any) { m.Called(msg, keyvals) }
func (m *MockLogger) With(keyvals ...any) log.Logger   { return m }
func (m *MockLogger) Impl() any                        { return m }

// fillPendingHeaders populates the given PendingHeaders with a sequence of mock SignedHeader objects for testing.
// It generates headers with consecutive heights and stores them in the underlying store so that PendingHeaders logic can retrieve them.
//
// Parameters:
//
//	ctx: context for store operations
//	t: the testing.T instance
//	pendingHeaders: the PendingHeaders instance to fill
//	chainID: the chain ID to use for generated headers
//	startHeight: the starting height for headers (default 1 if 0)
//	count: the number of headers to generate (default 3 if 0)
func fillPendingHeaders(ctx context.Context, t *testing.T, pendingHeaders *PendingHeaders, chainID string, startHeightAndCount ...uint64) {
	t.Helper()
	startHeight := uint64(1)
	count := uint64(3)
	if len(startHeightAndCount) > 0 && startHeightAndCount[0] != 0 {
		startHeight = startHeightAndCount[0]
	}
	if len(startHeightAndCount) > 1 && startHeightAndCount[1] != 0 {
		count = startHeightAndCount[1]
	}

	store := pendingHeaders.base.store
	for i := uint64(0); i < count; i++ {
		height := startHeight + i
		header, data := types.GetRandomBlock(height, 0, chainID)
		sig := &header.Signature
		err := store.SaveBlockData(ctx, header, data, sig)
		require.NoError(t, err, "failed to save block data for header at height %d", height)
		err = store.SetHeight(ctx, height)
		require.NoError(t, err, "failed to set store height for header at height %d", height)
	}
}
