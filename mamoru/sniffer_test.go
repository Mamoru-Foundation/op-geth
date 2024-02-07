package mamoru

import (
	"fmt"
	"os"
	"testing"

	"github.com/Mamoru-Foundation/mamoru-sniffer-go/mamoru_sniffer"
	"github.com/ethereum/go-ethereum/mamoru/sync_state"
	"github.com/stretchr/testify/assert"
)

var _ sync_state.SyncClient = (*mockSyncClient)(nil)

type mockSyncClient struct {
	currentBlock uint64
	highestBlock uint64
	isSync       bool
}

func (m mockSyncClient) CurrentBlock() uint64 {
	return m.currentBlock
}

func (m mockSyncClient) HighestBlock() uint64 {
	return m.highestBlock
}

func (m mockSyncClient) IsSync() bool {
	return m.isSync
}

func TestSniffer_isSnifferEnable(t *testing.T) {
	t.Run("TRUE env is set 1", func(t *testing.T) {
		_ = os.Setenv("MAMORU_SNIFFER_ENABLE", "1")
		defer unsetEnvSnifferEnable()
		s := NewSniffer()
		got := s.isSnifferEnable()
		assert.True(t, got)
	})
	t.Run("TRUE env is set true", func(t *testing.T) {
		_ = os.Setenv("MAMORU_SNIFFER_ENABLE", "true")
		defer unsetEnvSnifferEnable()
		s := NewSniffer()
		got := s.isSnifferEnable()
		assert.True(t, got)
	})
	t.Run("FALSE env is set 0", func(t *testing.T) {
		_ = os.Setenv("MAMORU_SNIFFER_ENABLE", "0")
		defer unsetEnvSnifferEnable()
		s := NewSniffer()
		got := s.isSnifferEnable()
		assert.False(t, got)
	})
	t.Run("FALSE env is set 0", func(t *testing.T) {
		_ = os.Setenv("MAMORU_SNIFFER_ENABLE", "false")
		defer unsetEnvSnifferEnable()
		s := NewSniffer()
		got := s.isSnifferEnable()
		assert.False(t, got)
	})
	t.Run("FALSE env is not set", func(t *testing.T) {
		_ = os.Setenv("MAMORU_SNIFFER_ENABLE", "")
		defer unsetEnvSnifferEnable()
		s := NewSniffer()
		got := s.isSnifferEnable()
		assert.False(t, got)
	})
}

func unsetEnvSnifferEnable() {
	_ = os.Unsetenv("MAMORU_SNIFFER_ENABLE")
}

func TestSniffer_connect(t *testing.T) {
	t.Run("TRUE ", func(t *testing.T) {
		InitConnectFunc(func() (*mamoru_sniffer.Sniffer, error) { return nil, nil })
		got := connect()
		assert.True(t, got)
	})
	t.Run("FALSE connect have error", func(t *testing.T) {
		InitConnectFunc(func() (*mamoru_sniffer.Sniffer, error) { return nil, fmt.Errorf("Some err") })
		got := connect()
		assert.False(t, got)
	})
}

func TestSniffer_CheckRequirements1(t *testing.T) {
	currentBlock := uint64(1000)
	t.Run("TRUE ", func(t *testing.T) {
		_ = os.Setenv("MAMORU_SNIFFER_ENABLE", "true")
		defer unsetEnvSnifferEnable()
		InitConnectFunc(func() (*mamoru_sniffer.Sniffer, error) { return nil, nil })
		InitSyncProcess(func() (sync_state.SyncClient, error) {
			return &mockSyncClient{isSync: true, highestBlock: 1000, currentBlock: 1000}, nil
		})
		s := &Sniffer{}
		assert.True(t, s.CheckRequirements(currentBlock))
	})
	t.Run("FALSE chain not sync ", func(t *testing.T) {
		_ = os.Setenv("MAMORU_SNIFFER_ENABLE", "true")
		defer unsetEnvSnifferEnable()
		SnifferConnectFunc = func() (*mamoru_sniffer.Sniffer, error) { return nil, nil }
		InitSyncProcess(func() (sync_state.SyncClient, error) {
			return &mockSyncClient{isSync: false, highestBlock: 1000, currentBlock: 1000}, nil
		})
		s := &Sniffer{}
		assert.False(t, s.CheckRequirements(currentBlock))
	})
	t.Run("FALSE connect error", func(t *testing.T) {
		_ = os.Setenv("MAMORU_SNIFFER_ENABLE", "true")
		defer unsetEnvSnifferEnable()
		InitConnectFunc(func() (*mamoru_sniffer.Sniffer, error) { return nil, fmt.Errorf("Some err") })
		InitSyncProcess(func() (sync_state.SyncClient, error) {
			return &mockSyncClient{isSync: true, highestBlock: 1000, currentBlock: 1000}, nil
		})
		s := &Sniffer{}
		assert.False(t, s.CheckRequirements(currentBlock))
	})
	t.Run("FALSE env not set", func(t *testing.T) {
		_ = os.Setenv("MAMORU_SNIFFER_ENABLE", "0")
		defer unsetEnvSnifferEnable()
		InitConnectFunc(func() (*mamoru_sniffer.Sniffer, error) { return nil, nil })
		InitSyncProcess(func() (sync_state.SyncClient, error) {
			return &mockSyncClient{isSync: true, highestBlock: 1000, currentBlock: 1000}, nil
		})
		s := &Sniffer{}
		assert.False(t, s.CheckRequirements(currentBlock))
	})
}
