// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// Package f2cclient provides connection management around the generated
// f2c.v1 gRPC client. Handler code in internal/api obtains a *Client at
// startup via Dial, uses the embedded F2CClient interface to make RPCs,
// and calls Close on graceful shutdown.

package f2cclient

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps a generated F2CClient with ownership of its underlying
// *grpc.ClientConn. Embed-by-interface gives callers the full RPC
// surface without re-declaring methods on this type.
type Client struct {
	conn *grpc.ClientConn
	F2CClient
}

// Dial opens a connection to the f2c-grpc-server on the given Unix
// domain socket path and returns a Client wrapping the generated F2C
// stub.
//
// socketPath is a filesystem path (e.g., "/tmp/f2c.sock"); Dial
// constructs the corresponding "unix://"+path target string itself.
//
// grpc.NewClient is lazy: it returns a *grpc.ClientConn immediately and
// defers the actual transport handshake until the first RPC. A
// successful Dial therefore does NOT prove the server is reachable;
// callers wanting that guarantee must issue a probe RPC (see /readyz
// in internal/api).
func Dial(socketPath string) (*Client, error) {
	if socketPath == "" {
		return nil, fmt.Errorf("f2cclient.Dial: empty socket path")
	}
	target := "unix://" + socketPath
	// Phase-12 integration tests produce ~5 MB Path responses (tens of
	// thousands of PathState poses) — comfortably above the gRPC 4 MiB
	// default. Raise both send and receive caps to 64 MiB so the client
	// can round-trip realistic coverage outputs without surfacing
	// ResourceExhausted on /paths:plan.
	const maxMsgBytes = 64 * 1024 * 1024
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMsgBytes),
			grpc.MaxCallSendMsgSize(maxMsgBytes),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("f2cclient.Dial(%s): %w", target, err)
	}
	return &Client{
		conn:      conn,
		F2CClient: NewF2CClient(conn),
	}, nil
}

// Close releases the underlying *grpc.ClientConn. Safe to call multiple
// times — the second call may return a non-nil error from grpc but will
// not panic.
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Compile-time assertion: *Client must satisfy F2CClient via embedding.
var _ F2CClient = (*Client)(nil)
