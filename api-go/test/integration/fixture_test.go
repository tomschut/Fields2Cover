// Copyright (C) 2026 Wageningen University — BSD-3-Clause

//go:build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/Fields2Cover/fields2cover/api-go/internal/f2cclient"
	"github.com/Fields2Cover/fields2cover/api-go/internal/pipeline"
)

// testSuite owns one f2c-grpc-server subprocess + one Go API
// httptest.Server. Shared across all tests in the package via TestMain.
type testSuite struct {
	// Shim subprocess
	shimBinary string // resolved absolute path
	socketPath string // <tmpdir>/f2c.sock
	tmpDir     string
	shimProc   *exec.Cmd

	// Go API
	f2cClient  *f2cclient.Client
	httpServer *httptest.Server
}

var suite *testSuite

// baseURL returns the http://127.0.0.1:PORT address of the in-process
// Go API for tests to dial.
func baseURL(t *testing.T) string {
	t.Helper()
	if suite == nil || suite.httpServer == nil {
		t.Fatalf("integration suite not started")
	}
	return suite.httpServer.URL
}

// TestMain builds the shim binary if missing, starts it, wires up the
// Go API, runs the tests, then tears everything down.
func TestMain(m *testing.M) {
	s := &testSuite{}
	if err := s.setup(); err != nil {
		fmt.Fprintf(os.Stderr, "integration fixture setup failed: %v\n", err)
		_ = s.teardown()
		os.Exit(1)
	}
	suite = s
	code := m.Run()
	if err := s.teardown(); err != nil {
		fmt.Fprintf(os.Stderr, "integration fixture teardown error: %v\n", err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}

// setup builds + starts the shim and the Go API.
func (s *testSuite) setup() error {
	// 1. Locate or build the shim binary.
	bin, err := resolveShimBinary()
	if err != nil {
		return fmt.Errorf("resolve shim binary: %w", err)
	}
	s.shimBinary = bin

	// 2. Pick a temp socket path. Unix socket paths have a ~108 byte
	// limit on Linux; /tmp is always short enough.
	tmp, err := os.MkdirTemp("", "f2c-int-*")
	if err != nil {
		return fmt.Errorf("mktemp: %w", err)
	}
	s.tmpDir = tmp
	s.socketPath = filepath.Join(tmp, "f2c.sock")

	// 3. Launch the shim subprocess.
	cmd := exec.Command(s.shimBinary)
	cmd.Env = append(os.Environ(), "F2C_GRPC_SOCKET="+s.socketPath)
	cmd.Stdout = os.Stderr // route both to test output for debugging
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // own process group so we can signal cleanly
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start f2c-grpc-server: %w", err)
	}
	s.shimProc = cmd

	// 4. Wait for the socket to appear.
	if err := waitForSocket(s.socketPath, 10*time.Second); err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("shim socket did not appear: %w", err)
	}

	// 5. Dial gRPC + build pipeline.
	client, err := f2cclient.Dial(s.socketPath)
	if err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("dial shim: %w", err)
	}
	s.f2cClient = client

	// 6. Build the Go API router and expose it via httptest.Server.
	router := buildIntegrationRouter(client)
	s.httpServer = httptest.NewServer(router)

	// 7. Prove the stack is reachable via /healthz.
	if err := pingHealthz(s.httpServer.URL); err != nil {
		s.httpServer.Close()
		_ = cmd.Process.Kill()
		return fmt.Errorf("healthz smoke: %w", err)
	}

	return nil
}

// teardown reverses setup, in reverse order.
func (s *testSuite) teardown() error {
	var firstErr error
	setErr := func(e error) {
		if firstErr == nil && e != nil {
			firstErr = e
		}
	}

	if s.httpServer != nil {
		s.httpServer.Close()
	}
	if s.f2cClient != nil {
		setErr(s.f2cClient.Close())
	}
	if s.shimProc != nil && s.shimProc.Process != nil {
		_ = s.shimProc.Process.Signal(syscall.SIGTERM)
		done := make(chan error, 1)
		go func() { done <- s.shimProc.Wait() }()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			_ = s.shimProc.Process.Kill()
			<-done
		}
	}
	if s.socketPath != "" {
		_ = os.Remove(s.socketPath)
	}
	if s.tmpDir != "" {
		_ = os.RemoveAll(s.tmpDir)
	}
	return firstErr
}

// resolveShimBinary finds f2c-grpc-server or builds it on the fly.
// Priority:
//  1. F2C_GRPC_SERVER env var (explicit path)
//  2. <repo>/build/f2c-grpc/f2c-grpc-server
//  3. `cmake --build <repo>/build --target f2c-grpc-server`, then #2
func resolveShimBinary() (string, error) {
	if p := os.Getenv("F2C_GRPC_SERVER"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
		return "", fmt.Errorf("F2C_GRPC_SERVER=%s but file missing", p)
	}

	repo, err := repoRoot()
	if err != nil {
		return "", err
	}
	candidate := filepath.Join(repo, "build", "f2c-grpc", "f2c-grpc-server")
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	// Build on the fly.
	fmt.Fprintf(os.Stderr, "integration: building f2c-grpc-server via cmake...\n")
	build := exec.Command("cmake", "--build",
		filepath.Join(repo, "build"), "--target", "f2c-grpc-server")
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return "", fmt.Errorf("cmake build failed (is <repo>/build configured with -DBUILD_F2C_GRPC=ON?): %w", err)
	}
	if _, err := os.Stat(candidate); err != nil {
		return "", fmt.Errorf("cmake build reported success but binary missing at %s", candidate)
	}
	return candidate, nil
}

// repoRoot walks up from this source file's directory looking for a
// `.git` sibling.
func repoRoot() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repo root (.git) not found above %s", filepath.Dir(thisFile))
		}
		dir = parent
	}
}

// waitForSocket polls until the socket file exists or timeout elapses.
func waitForSocket(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("socket %s did not appear within %s", path, timeout)
}

// pingHealthz fetches /healthz and checks for 200.
func pingHealthz(baseURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/healthz", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("healthz returned %d: %s", resp.StatusCode, body)
	}
	return nil
}

// buildIntegrationRouter mirrors cmd/api/main.go's buildRouter.
// Kept as a tiny copy here because cmd/api is package main and can't
// be imported. Integration tests assert on the same HTTP surface so
// drift is caught by the tests themselves.
func buildIntegrationRouter(client f2cclient.F2CClient) http.Handler {
	p := pipeline.NewFromClient(client)
	return buildRouterForIntegration(p)
}
