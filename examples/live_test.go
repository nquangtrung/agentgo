package main

import (
	"os"
	"strings"
	"testing"
)

// liveEnvVar gates every example test that calls a real provider API.
//
// Live tests make billable network calls and need the matching API key in
// ../.env. They are skipped unless AGENTGO_LIVE_TESTS is explicitly set to 1 or
// true, so a plain `go test ./...` stays offline and free.
//
// Run them only when the user explicitly asks, for example:
//
//	AGENTGO_LIVE_TESTS=1 go test -v ./examples/ -run TestGenerateTextDeepSeek
//
// Do not set AGENTGO_LIVE_TESTS for routine test runs or CI. Values other than
// 1/true (including 0) leave the live tests skipped on purpose.
const liveEnvVar = "AGENTGO_LIVE_TESTS"

// requireLiveTest skips the calling test unless live tests are opted into.
//
// It must be called before utils.LoadEnv, because LoadEnv calls log.Fatal when
// ../.env is missing and would turn a skipped run into a hard failure.
func requireLiveTest(t *testing.T) {
	t.Helper()
	switch strings.ToLower(os.Getenv(liveEnvVar)) {
	case "1", "true":
	default:
		t.Skipf("live API test skipped: set %s=1 to run (only on explicit user request)", liveEnvVar)
	}
}
