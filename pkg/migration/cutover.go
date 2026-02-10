package migration

import "time"

// CutoverChecker checks whether the D14N migration cutover time has passed.
// It is used by MLS and Identity services to disable publishing after cutover.
type CutoverChecker struct {
	cutoverNs uint64
}

// NewCutoverChecker creates a new CutoverChecker with the given cutover timestamp.
func NewCutoverChecker(cutoverNs uint64) *CutoverChecker {
	return &CutoverChecker{cutoverNs: cutoverNs}
}

// IsMigrationComplete returns true if the cutover time has passed.
// Returns false if the checker is nil or cutoverNs is 0.
func (c *CutoverChecker) IsMigrationComplete() bool {
	if c == nil || c.cutoverNs == 0 {
		return false
	}
	return uint64(time.Now().UnixNano()) >= c.cutoverNs
}

// CutoverNs returns the configured cutover timestamp in nanoseconds.
func (c *CutoverChecker) CutoverNs() uint64 {
	if c == nil {
		return 0
	}
	return c.cutoverNs
}
