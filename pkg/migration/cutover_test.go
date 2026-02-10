package migration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCutoverChecker_IsMigrationComplete_NilChecker(t *testing.T) {
	var checker *CutoverChecker
	require.False(t, checker.IsMigrationComplete())
}

func TestCutoverChecker_IsMigrationComplete_ZeroCutover(t *testing.T) {
	checker := NewCutoverChecker(0)
	require.False(t, checker.IsMigrationComplete())
}

func TestCutoverChecker_IsMigrationComplete_FutureTime(t *testing.T) {
	// Set cutover to 1 hour in the future
	futureNs := uint64(time.Now().Add(time.Hour).UnixNano())
	checker := NewCutoverChecker(futureNs)
	require.False(t, checker.IsMigrationComplete())
}

func TestCutoverChecker_IsMigrationComplete_PastTime(t *testing.T) {
	// Set cutover to 1 hour in the past
	pastNs := uint64(time.Now().Add(-time.Hour).UnixNano())
	checker := NewCutoverChecker(pastNs)
	require.True(t, checker.IsMigrationComplete())
}

func TestCutoverChecker_IsMigrationComplete_ExactlyNow(t *testing.T) {
	// Set cutover to now (should be complete since we use >=)
	nowNs := uint64(time.Now().UnixNano())
	checker := NewCutoverChecker(nowNs)
	// Give a tiny bit of slack for test execution
	require.True(t, checker.IsMigrationComplete())
}

func TestCutoverChecker_CutoverNs(t *testing.T) {
	cutover := uint64(1704067200000000000)
	checker := NewCutoverChecker(cutover)
	require.Equal(t, cutover, checker.CutoverNs())
}

func TestCutoverChecker_CutoverNs_NilChecker(t *testing.T) {
	var checker *CutoverChecker
	require.Equal(t, uint64(0), checker.CutoverNs())
}
