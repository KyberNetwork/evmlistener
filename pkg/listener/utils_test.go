package listener

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/KyberNetwork/evmlistener/pkg/types"
)

func makeLog(addr, topic0, topic1 string) types.Log {
	l := types.Log{Address: addr}
	if topic0 != "" {
		l.Topics = append(l.Topics, topic0)
	}
	if topic1 != "" {
		l.Topics = append(l.Topics, topic1)
	}
	return l
}

func TestFilterSpamLogs_BelowThreshold(t *testing.T) {
	logs := []types.Log{
		makeLog("0xA", "0xT0", "0xT1"),
		makeLog("0xA", "0xT0", "0xT1"),
	}
	got := filterSpamLogs(logs, 3)
	assert.Equal(t, logs, got)
}

func TestFilterSpamLogs_AboveThreshold_KeepsLast(t *testing.T) {
	// 4 logs sharing the same (addr, topic0, topic1) key, threshold = 3.
	logs := []types.Log{
		makeLog("0xA", "0xT0", "0xT1"), // index 0 — dropped
		makeLog("0xA", "0xT0", "0xT1"), // index 1 — dropped
		makeLog("0xA", "0xT0", "0xT1"), // index 2 — dropped
		makeLog("0xA", "0xT0", "0xT1"), // index 3 — kept (last)
	}
	got := filterSpamLogs(logs, 3)
	assert.Len(t, got, 1)
	assert.Equal(t, logs[3], got[0])
}

func TestFilterSpamLogs_DifferentTopic1_AllKept(t *testing.T) {
	// Simulates Uniswap V4 / Balancer: same address+topic0, different pool ID in topic1.
	logs := []types.Log{
		makeLog("0xVault", "0xSwap", "0xPool1"),
		makeLog("0xVault", "0xSwap", "0xPool2"),
		makeLog("0xVault", "0xSwap", "0xPool3"),
		makeLog("0xVault", "0xSwap", "0xPool4"),
	}
	got := filterSpamLogs(logs, 3)
	assert.Equal(t, logs, got)
}

func TestFilterSpamLogs_AnonymousEvents_KeyedByAddress(t *testing.T) {
	// 4 anonymous logs (no topics) from the same address, threshold = 3.
	logs := []types.Log{
		{Address: "0xSpam"}, // index 0 — dropped
		{Address: "0xSpam"}, // index 1 — dropped
		{Address: "0xSpam"}, // index 2 — dropped
		{Address: "0xSpam"}, // index 3 — kept
	}
	got := filterSpamLogs(logs, 3)
	assert.Len(t, got, 1)
	assert.Equal(t, logs[3], got[0])
}

func TestFilterSpamLogs_AnonymousEvents_DifferentAddress_AllKept(t *testing.T) {
	logs := []types.Log{
		{Address: "0xA"},
		{Address: "0xB"},
		{Address: "0xC"},
		{Address: "0xD"},
	}
	got := filterSpamLogs(logs, 3)
	assert.Equal(t, logs, got)
}

func TestFilterSpamLogs_PreservesOrderAndNonSpam(t *testing.T) {
	// Mix of spam and legitimate logs; verify order is preserved.
	logs := []types.Log{
		makeLog("0xSpam", "0xT0", "0xT1"),  // 0 — dropped
		makeLog("0xLegit", "0xT0", "0xP1"), // 1 — kept (unique pool)
		makeLog("0xSpam", "0xT0", "0xT1"),  // 2 — dropped
		makeLog("0xLegit", "0xT0", "0xP2"), // 3 — kept (unique pool)
		makeLog("0xSpam", "0xT0", "0xT1"),  // 4 — dropped
		makeLog("0xLegit", "0xT0", "0xP3"), // 5 — kept (unique pool)
		makeLog("0xSpam", "0xT0", "0xT1"),  // 6 — kept (last spam)
	}
	got := filterSpamLogs(logs, 3)
	assert.Equal(t, []types.Log{logs[1], logs[3], logs[5], logs[6]}, got)
}
