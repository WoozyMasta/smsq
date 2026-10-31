// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package grpc

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mdp/qrterminal/v4"
	"github.com/woozymasta/flags"
	cm "github.com/woozymasta/steamcm"
	"github.com/woozymasta/steamcm/auth"
	"golang.org/x/term"
)

// cmLogin retains the connection and authenticated session for one operation.
type cmLogin struct {
	client  *cm.Client  // client owns the connected transport and session lifetime.
	session *cm.Session // session carries the authenticated Steam identity.
}

// authTimeoutTransport applies a per-call bound to unauthenticated auth requests.
type authTimeoutTransport struct {
	client  *cm.Client    // client sends the request over the established CM connection.
	timeout time.Duration // timeout bounds each authentication call independently.
}

// qrDisplay keeps one QR challenge visible and replaces it when Steam rotates it.
type qrDisplay struct {
	// output receives the prompt and rendered challenge.
	output io.Writer
	// terminal enables in-place refresh only for interactive stderr.
	terminal bool
	// lineCount tracks the block written by the previous challenge.
	lineCount int
}

// CallUnauthenticated delegates an auth request with a per-call deadline.
func (transport authTimeoutTransport) CallUnauthenticated(
	ctx context.Context,
	method string,
	body []byte,
) (cm.Message, error) {
	callCtx, cancel := context.WithTimeout(ctx, transport.timeout)
	defer cancel()

	return transport.client.CallUnauthenticated(callCtx, method, body)
}

// authenticateCM reuses a valid refresh token or runs an interactive login.
func authenticateCM(
	ctx context.Context,
	endpoints []cm.Endpoint,
	mode, account, device string,
	storage tokenStorage,
	stdin *os.File,
	output io.Writer,
	authTimeout time.Duration,
) (cmLogin, error) {
	token, err := storage.read()
	if err != nil {
		return cmLogin{}, err
	}

	if token != nil {
		login, logonErr := logonCMToken(ctx, endpoints, token)
		if logonErr == nil {
			return login, nil
		}

		// Only known token rejection results justify prompting for new credentials.
		if !isRejectedToken(logonErr) {
			return cmLogin{}, fmt.Errorf("CM token logon failed: %w", logonErr)
		}
	}

	newToken, err := acquireCMToken(ctx, endpoints, mode, account, device, stdin, output, authTimeout)
	if err != nil {
		return cmLogin{}, err
	}
	if err := storage.write(newToken); err != nil {
		return cmLogin{}, err
	}

	return logonCMToken(ctx, endpoints, newToken)
}

// logonCMToken establishes an authenticated session using a saved refresh token.
func logonCMToken(ctx context.Context, endpoints []cm.Endpoint, token *auth.Token) (cmLogin, error) {
	_, client, err := connectCM(ctx, endpoints)
	if err != nil {
		return cmLogin{}, err
	}

	session, err := client.LogonToken(ctx, cm.TokenLogonDetails{
		SteamID:      token.SteamID,
		AccountName:  token.AccountName,
		RefreshToken: token.RefreshToken,
	})
	if err != nil {
		_ = client.Close()
		return cmLogin{}, err
	}

	return cmLogin{client: client, session: session}, nil
}

// acquireCMToken completes QR or credential-based Steam authentication.
func acquireCMToken(
	ctx context.Context,
	endpoints []cm.Endpoint,
	mode, account, device string,
	stdin *os.File,
	output io.Writer,
	authTimeout time.Duration,
) (*auth.Token, error) {
	if err := requireInteractiveAuth(mode, stdin, output); err != nil {
		return nil, err
	}

	endpoint, client, err := connectCM(ctx, endpoints)
	if err != nil {
		return nil, err
	}
	defer func() { _ = client.Close() }()

	authClient, err := auth.New(authTimeoutTransport{client: client, timeout: authTimeout})
	if err != nil {
		return nil, err
	}

	var attempt *auth.Attempt
	display := newQRDisplay(output)
	if mode == "qr" {
		attempt, err = authClient.BeginQR(ctx, device)
		if err != nil {
			return nil, fmt.Errorf("start QR authentication: %w", err)
		}
		if err := display.show(attempt.ChallengeURL()); err != nil {
			return nil, err
		}
	} else {
		// Credential flows may require a Steam Guard code or mobile approval.
		var password []byte
		account, password, err = readLoginCredentials(stdin, output, account)
		if err != nil {
			return nil, err
		}
		defer clear(password)

		attempt, err = authClient.BeginCredentials(ctx, auth.Credentials{
			AccountName: account,
			DeviceName:  device,
			Password:    password,
		})
		if err != nil {
			return nil, fmt.Errorf("start credential authentication: %w", err)
		}
		if err := completeCMGuard(ctx, stdin, output, mode, attempt); err != nil {
			return nil, err
		}
	}

	result, err := pollCMToken(ctx, attempt, display, mode)
	if err != nil {
		return nil, err
	}
	if result.Token == nil {
		return nil, errors.New("steam authentication completed without a refresh token")
	}
	if err := client.Close(); err != nil {
		return nil, fmt.Errorf("close pre-login CM connection to %s: %w", endpoint.Address, err)
	}

	return result.Token, nil
}

// connectCM tries directory endpoints in order and returns the first live connection.
func connectCM(ctx context.Context, endpoints []cm.Endpoint) (cm.Endpoint, *cm.Client, error) {
	var lastErr error
	for _, endpoint := range endpoints {
		client, err := cm.New()
		if err != nil {
			return cm.Endpoint{}, nil, err
		}

		if err := client.Connect(ctx, endpoint); err == nil {
			return endpoint, client, nil
		}

		lastErr = err
		_ = client.Close()
	}

	if lastErr == nil {
		return cm.Endpoint{}, nil, errors.New("CM directory returned no endpoints")
	}

	return cm.Endpoint{}, nil, fmt.Errorf("could not connect to a CM endpoint: %w", lastErr)
}

// completeCMGuard submits a code or waits for the selected mobile approval.
func completeCMGuard(
	ctx context.Context,
	stdin *os.File,
	output io.Writer,
	mode string,
	attempt *auth.Attempt,
) error {
	var selected *auth.Confirmation
	// Prefer a device code over email when both are offered.
	for _, confirmation := range attempt.Confirmations() {
		if mode == "approve" && confirmation.Type == auth.ConfirmationDeviceConfirmation {
			selected = &confirmation
			break
		}

		if mode == "code" &&
			(confirmation.Type == auth.ConfirmationDeviceCode || confirmation.Type == auth.ConfirmationEmailCode) {
			selected = &confirmation
			if confirmation.Type == auth.ConfirmationDeviceCode {
				break
			}
		}
	}

	if selected == nil {
		if len(attempt.Confirmations()) == 0 {
			return nil
		}

		return fmt.Errorf("steam did not offer the guard method selected by -auth-mode %q", mode)
	}

	if selected.Type == auth.ConfirmationDeviceConfirmation {
		_, err := fmt.Fprintln(output, "Approve the sign-in request in the Steam mobile app.")
		return err
	}

	code, err := readHiddenInput(stdin, output, "Steam Guard code: ")
	if err != nil {
		return err
	}
	defer clear(code)

	if err := attempt.SubmitCode(ctx, selected.Type, string(code)); err != nil {
		return fmt.Errorf("submit Steam Guard code: %w", err)
	}

	return nil
}

// pollCMToken refreshes changing QR challenges and waits until auth completes.
func pollCMToken(
	ctx context.Context,
	attempt *auth.Attempt,
	display *qrDisplay,
	mode string,
) (auth.PollResult, error) {
	lastChallenge := attempt.ChallengeURL()
	for {
		challenge := attempt.ChallengeURL()
		if challenge != "" && challenge != lastChallenge {
			if err := display.show(challenge); err != nil {
				return auth.PollResult{}, err
			}

			lastChallenge = challenge
		}

		result, err := attempt.PollOnce(ctx)
		if err != nil {
			return auth.PollResult{}, authPollError(ctx, mode, err)
		}

		if !result.Pending {
			return result, nil
		}

		if result.ChallengeURL != "" && result.ChallengeURL != lastChallenge {
			if err := display.show(result.ChallengeURL); err != nil {
				return auth.PollResult{}, err
			}

			lastChallenge = result.ChallengeURL
		}
		timer := time.NewTimer(result.NextPollInterval)

		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return auth.PollResult{}, authPollError(ctx, mode, ctx.Err())
		}
	}
}

// authPollError turns auth deadline and cancellation failures into user-facing guidance.
func authPollError(ctx context.Context, mode string, err error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		stage := "Steam Guard confirmation"
		if mode == "qr" {
			stage = "QR/mobile approval"
		}

		return fmt.Errorf(
			"steam %s timed out waiting for authentication; complete the prompt before --timeout expires or increase --timeout: %w",
			stage,
			context.DeadlineExceeded,
		)
	}
	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) && closeErr.Code == websocket.CloseNormalClosure {
		if mode == "qr" {
			return errors.New(
				"steam QR approval did not finish before the login session closed; the QR may have expired " +
					"or approval timed out. Run the command again, scan the new QR, and approve it promptly",
			)
		}

		return errors.New("steam closed the authentication session before login completed; retry the login")
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return fmt.Errorf("steam authentication interrupted: %w", context.Canceled)
	}

	return err
}

// newQRDisplay detects whether challenge updates can replace a QR in place.
func newQRDisplay(output io.Writer) *qrDisplay {
	return &qrDisplay{output: output, terminal: flags.DetectWriterTTY(output)}
}

// requireInteractiveAuth rejects login modes that cannot be completed without a terminal.
func requireInteractiveAuth(mode string, stdin *os.File, output io.Writer) error {
	if !flags.DetectWriterTTY(output) {
		return errors.New("steam authentication requires a terminal on stderr")
	}
	if mode != "qr" && !flags.DetectFileTTY(stdin) {
		return errors.New("steam account authentication requires a terminal on stdin")
	}

	return nil
}

// show renders the initial QR and refreshes it in place on interactive terminals.
func (display *qrDisplay) show(value string) error {
	if display.lineCount > 0 && !display.terminal {
		return nil
	}

	var block bytes.Buffer
	if _, err := fmt.Fprintln(&block, "Scan this QR code with the Steam mobile app:"); err != nil {
		return err
	}
	qrterminal.GenerateWithConfig(value, qrterminal.Config{
		Level:      qrterminal.M,
		Writer:     &block,
		HalfBlocks: true,
	})

	if display.terminal && display.lineCount > 0 {
		if _, err := fmt.Fprintf(display.output, "\x1b[%dA\r", display.lineCount); err != nil {
			return err
		}

		for range display.lineCount {
			if _, err := io.WriteString(display.output, "\x1b[2K\n"); err != nil {
				return err
			}
		}

		if _, err := fmt.Fprintf(display.output, "\x1b[%dA\r", display.lineCount); err != nil {
			return err
		}
	}

	rendered := block.String()
	if _, err := io.WriteString(display.output, rendered); err != nil {
		return err
	}

	display.lineCount = strings.Count(rendered, "\n")
	return nil
}

// readLoginCredentials reads an account name and password without echoing the password.
func readLoginCredentials(stdin *os.File, output io.Writer, account string) (string, []byte, error) {
	if account == "" {
		reader := bufio.NewReader(stdin)
		if _, err := fmt.Fprint(output, "Steam account: "); err != nil {
			return "", nil, err
		}

		value, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", nil, err
		}

		account = strings.TrimSpace(value)
		if account == "" || len(account) > 128 {
			return "", nil, errors.New("steam account is empty or too long")
		}
	}

	account = strings.TrimSpace(account)
	if account == "" || len(account) > 128 {
		return "", nil, errors.New("steam account is empty or too long")
	}
	if !flags.DetectFileTTY(stdin) {
		return "", nil, errors.New("steam password input requires an interactive terminal")
	}
	if _, err := fmt.Fprint(output, "Steam password: "); err != nil {
		return "", nil, err
	}

	password, readErr := term.ReadPassword(int(stdin.Fd()))
	_, writeErr := fmt.Fprintln(output)
	if readErr != nil {
		clear(password)
		return "", nil, readErr
	}
	if writeErr != nil {
		clear(password)
		return "", nil, writeErr
	}
	if len(password) == 0 || len(password) > 4096 {
		clear(password)
		return "", nil, errors.New("steam password is empty or too long")
	}

	return account, password, nil
}

// readHiddenInput reads a bounded secret from an interactive terminal.
func readHiddenInput(stdin *os.File, output io.Writer, prompt string) ([]byte, error) {
	if !flags.DetectFileTTY(stdin) {
		return nil, errors.New("steam guard code input requires an interactive terminal")
	}
	if _, err := fmt.Fprint(output, prompt); err != nil {
		return nil, err
	}

	code, readErr := term.ReadPassword(int(stdin.Fd()))
	_, writeErr := fmt.Fprintln(output)
	if readErr != nil {
		clear(code)
		return nil, readErr
	}
	if writeErr != nil {
		clear(code)
		return nil, writeErr
	}
	if len(code) == 0 || len(code) > 64 {
		clear(code)
		return nil, errors.New("steam Guard code is empty or too long")
	}

	return code, nil
}

// isRejectedToken recognizes EResults that allow interactive reauthentication.
func isRejectedToken(err error) bool {
	var logonErr *cm.LogonError
	if !errors.As(err, &logonErr) {
		return false
	}

	// These EResults were observed for invalid, revoked, and expired refresh tokens.
	return logonErr.EResult == 5 || logonErr.EResult == 26 || logonErr.EResult == 27
}
