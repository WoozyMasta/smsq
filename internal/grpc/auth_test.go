// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package grpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestAuthPollErrorExplainsTimeout(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	err := authPollError(ctx, "qr", errors.New("websocket closed normally"))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("authPollError() = %v, want deadline exceeded", err)
	}
	for _, expected := range []string{"QR/mobile approval", "timed out", "increase --timeout"} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("authPollError() = %q, missing %q", err, expected)
		}
	}
}

func TestAuthPollErrorPreservesUnderlyingFailure(t *testing.T) {
	cause := errors.New("websocket closed normally")
	if got := authPollError(context.Background(), "qr", cause); got != cause {
		t.Fatalf("authPollError() = %v, want original error %v", got, cause)
	}

	err := authPollError(context.Background(), "qr", context.DeadlineExceeded)
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("authPollError(deadline) = %v, want a clear timeout error", err)
	}
}

func TestAuthPollErrorExplainsNormalQRSessionClose(t *testing.T) {
	err := authPollError(context.Background(), "qr", &websocket.CloseError{
		Code: websocket.CloseNormalClosure,
	})
	if err == nil || !strings.Contains(err.Error(), "QR may have expired") ||
		strings.Contains(err.Error(), "websocket") {
		t.Fatalf("authPollError() = %v, want a clear expired-QR message", err)
	}
}

func TestQRDisplayReplacesChallengeInPlace(t *testing.T) {
	var output bytes.Buffer
	display := &qrDisplay{output: &output, terminal: true}
	if err := display.show("https://s.team/q/first"); err != nil {
		t.Fatalf("show(first) error = %v", err)
	}
	previousLines := display.lineCount
	output.Reset()

	if err := display.show("https://s.team/q/second"); err != nil {
		t.Fatalf("show(second) error = %v", err)
	}

	got := output.String()
	if !strings.HasPrefix(got, fmt.Sprintf("\x1b[%dA\r", previousLines)) {
		t.Fatalf("refresh did not move to the previous QR block: %q", got)
	}
	if count := strings.Count(got, "\x1b[2K\n"); count != previousLines {
		t.Fatalf("refresh cleared %d lines, want %d", count, previousLines)
	}
	if !strings.Contains(got, "Scan this QR code with the Steam mobile app:") {
		t.Fatal("refresh omitted the QR prompt")
	}
}

func TestRequireInteractiveAuthNeedsTTY(t *testing.T) {
	stdin, stdinWrite, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(stdin) error = %v", err)
	}
	defer func() { _ = stdin.Close() }()
	defer func() { _ = stdinWrite.Close() }()

	_, stderr, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(stderr) error = %v", err)
	}
	defer func() { _ = stderr.Close() }()

	if err := requireInteractiveAuth("qr", nil, stderr); err == nil || !strings.Contains(err.Error(), "stderr") {
		t.Fatalf("requireInteractiveAuth(qr) error = %v, want stderr TTY error", err)
	}
	if err := requireInteractiveAuth("code", stdin, &bytes.Buffer{}); err == nil ||
		!strings.Contains(err.Error(), "stdin") {
		t.Fatalf("requireInteractiveAuth(code) error = %v, want stdin TTY error", err)
	}
	if err := requireInteractiveAuth("qr", nil, &bytes.Buffer{}); err != nil {
		t.Fatalf("requireInteractiveAuth(qr with TTY writer) error = %v", err)
	}
}

func TestQRDisplayPrintsOnceWithoutTerminal(t *testing.T) {
	var output bytes.Buffer
	display := &qrDisplay{output: &output}
	if err := display.show("https://s.team/q/first"); err != nil {
		t.Fatalf("show(first) error = %v", err)
	}
	first := output.String()

	if err := display.show("https://s.team/q/second"); err != nil {
		t.Fatalf("show(second) error = %v", err)
	}
	if got := output.String(); got != first {
		t.Fatalf("non-terminal refresh changed output; got %q, want original block", got)
	}
}
