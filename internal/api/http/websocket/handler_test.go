package websocket

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestStaticResolverConsoleSockPath(t *testing.T) {
	resolver := StaticResolver{ContainerRoot: "/tmp/containers", SockName: ""}
	got, err := resolver.ConsoleSockPath("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join("/tmp/containers", "abc", "tty.sock")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	resolver = StaticResolver{ContainerRoot: "/tmp/containers", SockName: "exec.sock"}
	got, err = resolver.ConsoleSockPath("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want = filepath.Join("/tmp/containers", "abc", "exec.sock")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestIsRetryableUnixDialErr(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		expect bool
	}{
		{name: "conn refused", err: syscall.ECONNREFUSED, expect: true},
		{name: "no such file", err: syscall.ENOENT, expect: true},
		{name: "permission", err: syscall.EACCES, expect: false},
		{name: "op error conn refused", err: &net.OpError{Err: syscall.ECONNREFUSED}, expect: true},
		{name: "op error syscall", err: &net.OpError{Err: &os.SyscallError{Err: syscall.ENOENT}}, expect: true},
		{name: "op error other", err: &net.OpError{Err: &os.SyscallError{Err: syscall.EPERM}}, expect: false},
		{name: "wrapped", err: errors.New("wrapped"), expect: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := isRetryableUnixDialErr(tc.err)
			if got != tc.expect {
				t.Fatalf("expected %v, got %v", tc.expect, got)
			}
		})
	}
}
