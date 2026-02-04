package logger

import (
	"net/http"
	"testing"
)

func TestPeerIp(t *testing.T) {
	req := &http.Request{RemoteAddr: "192.168.0.1:1234"}
	if got := peerIp(req); got != "192.168.0.1" {
		t.Fatalf("expected host only, got %q", got)
	}

	req = &http.Request{RemoteAddr: "not-a-host-port"}
	if got := peerIp(req); got != "not-a-host-port" {
		t.Fatalf("expected passthrough, got %q", got)
	}
}

func TestSeverityForAction(t *testing.T) {
	if got := severityForAction("hook.poststart"); got != SEV_MEDIUM {
		t.Fatalf("expected %d, got %d", SEV_MEDIUM, got)
	}
	if got := severityForAction("unknown.action"); got != SEV_LOW {
		t.Fatalf("expected %d, got %d", SEV_LOW, got)
	}
}

func TestBump(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect string
	}{
		{name: "info", input: "information", expect: "low"},
		{name: "low", input: "low", expect: "medium"},
		{name: "medium", input: "medium", expect: "high"},
		{name: "high", input: "high", expect: "critical"},
		{name: "unknown", input: "custom", expect: "custom"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := bump(tc.input)
			if got != tc.expect {
				t.Fatalf("expected %q, got %q", tc.expect, got)
			}
		})
	}
}
