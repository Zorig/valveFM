//go:build windows

package ipc

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func ResolveEndpoint() (Endpoint, error) {
	path, err := endpointFilePath()
	if err != nil {
		return Endpoint{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Endpoint{}, err
	}
	addr := strings.TrimSpace(string(data))
	if addr == "" {
		return Endpoint{}, errors.New("empty ipc address")
	}
	return Endpoint{Network: "tcp", Address: addr}, nil
}

func Listen() (net.Listener, Endpoint, error) {
	path, err := endpointFilePath()
	if err != nil {
		return nil, Endpoint{}, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, Endpoint{}, err
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, Endpoint{}, err
	}

	addr := listener.Addr().String()
	if err := os.WriteFile(path, []byte(addr), 0o600); err != nil {
		_ = listener.Close()
		return nil, Endpoint{}, err
	}

	return listener, Endpoint{Network: "tcp", Address: addr}, nil
}

func Cleanup(ep Endpoint) error {
	path, err := endpointFilePath()
	if err != nil {
		return err
	}
	_ = os.Remove(path)
	return nil
}

func endpointFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "valvefm", "ctl.addr"), nil
}
