// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package grpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/woozymasta/steamcm/auth"
)

// maxTokenFileSize bounds token files before JSON decoding.
const maxTokenFileSize = 1 << 20

// tokenDocument is the versioned on-disk representation of a Steam refresh token.
type tokenDocument struct {
	// AccountName identifies the Steam account bound to this token.
	AccountName string `json:"account_name"`
	// RefreshToken authorizes future CM logons without repeating Steam Guard.
	RefreshToken string `json:"refresh_token"`
	// GuardData preserves Steam's device confirmation state when supplied.
	GuardData string `json:"guard_data,omitempty"`
	// SteamID binds the token to the expected Steam account.
	SteamID uint64 `json:"steam_id"`
}

// tokenStorage reads and writes one plaintext refresh-token file.
type tokenStorage struct {
	// path is empty when tokens should remain in memory only.
	path string
}

// read loads a token when configured and treats a missing file as a fresh login.
func (storage tokenStorage) read() (*auth.Token, error) {
	if storage.path == "" {
		return nil, nil
	}

	file, err := os.Open(storage.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read token file: %w", err)
	}

	data, err := io.ReadAll(io.LimitReader(file, maxTokenFileSize+1))
	closeErr := file.Close()
	if err != nil {
		return nil, fmt.Errorf("read token file: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close token file: %w", closeErr)
	}
	if len(data) > maxTokenFileSize {
		return nil, errors.New("token file exceeds 1 MiB")
	}

	var document tokenDocument
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("decode token file: %w", err)
	}

	if document.SteamID == 0 ||
		document.AccountName == "" ||
		document.RefreshToken == "" ||
		len(document.RefreshToken) > 64<<10 {
		return nil, errors.New("token file is incomplete")
	}

	return &auth.Token{
		SteamID:      document.SteamID,
		AccountName:  document.AccountName,
		RefreshToken: document.RefreshToken,
		GuardData:    document.GuardData,
	}, nil
}

// write atomically persists a complete token with owner-only permissions.
func (storage tokenStorage) write(token *auth.Token) error {
	if storage.path == "" {
		return nil
	}
	if token == nil || token.SteamID == 0 || token.AccountName == "" || token.RefreshToken == "" {
		return errors.New("refusing to persist incomplete Steam token")
	}

	// #nosec G117 -- token persistence is intentional and the file is owner-only.
	data, err := json.Marshal(tokenDocument{
		SteamID:      token.SteamID,
		AccountName:  token.AccountName,
		RefreshToken: token.RefreshToken,
		GuardData:    token.GuardData,
	})
	if err != nil {
		return fmt.Errorf("encode token: %w", err)
	}
	if len(data) > maxTokenFileSize {
		return errors.New("token document exceeds 1 MiB")
	}

	directory := filepath.Dir(storage.path)
	temporary, err := os.CreateTemp(directory, ".smsq-token-*")
	if err != nil {
		return fmt.Errorf("create temporary token file: %w", err)
	}

	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("restrict token file permissions: %w", err)
	}

	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write token file: %w", err)
	}

	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("flush token file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close token file: %w", err)
	}
	if err := os.Rename(temporaryPath, storage.path); err != nil {
		return fmt.Errorf("replace token file: %w", err)
	}

	return nil
}
