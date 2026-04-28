// Copyright (C) 2026 Wageningen University — BSD-3-Clause

package f2cclient

import (
	"strings"
	"testing"
)

func TestDial_EmptySocketPath_Errors(t *testing.T) {
	t.Parallel()
	c, err := Dial("")
	if err == nil {
		t.Fatalf("Dial(\"\") = (%v, nil), want non-nil error", c)
	}
	if c != nil {
		t.Errorf("Dial(\"\") returned non-nil client = %v, want nil", c)
	}
	if !strings.Contains(err.Error(), "empty socket path") {
		t.Errorf("error message = %q, want substring %q", err.Error(), "empty socket path")
	}
}

func TestDial_NonexistentSocket_LazyOK(t *testing.T) {
	t.Parallel()
	// grpc.NewClient is lazy — it does not connect until the first RPC.
	// Dial against a path that does not exist must therefore succeed.
	c, err := Dial("/tmp/f2c-test-does-not-exist.sock")
	if err != nil {
		t.Fatalf("Dial(nonexistent) = (_, %v), want nil err (NewClient is lazy)", err)
	}
	if c == nil {
		t.Fatalf("Dial(nonexistent) = (nil, nil), want non-nil client")
	}
	defer c.Close()
}

func TestClose_AfterDial_OK(t *testing.T) {
	t.Parallel()
	c, err := Dial("/tmp/f2c-test-close-1.sock")
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("Close: %v, want nil", err)
	}
}

func TestClose_DoubleClose_NoPanic(t *testing.T) {
	t.Parallel()
	c, err := Dial("/tmp/f2c-test-close-2.sock")
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("first Close: %v, want nil", err)
	}
	// Second close is allowed to return an error but must not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("second Close panicked: %v", r)
		}
	}()
	_ = c.Close()
}

func TestClose_NilClient_NoPanic(t *testing.T) {
	t.Parallel()
	var c *Client
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Close on nil client panicked: %v", r)
		}
	}()
	if err := c.Close(); err != nil {
		t.Errorf("nil-receiver Close = %v, want nil", err)
	}
}

// Compile-time check duplicated in the test file so a future refactor
// that breaks the embedding is caught even if the production assertion
// is removed.
var _ F2CClient = (*Client)(nil)
