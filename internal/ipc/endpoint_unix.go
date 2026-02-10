//go:build !windows

package ipc

import (
	"net"
	"os"
	"path/filepath"
)

func ResolveEndpoint() (Endpoint, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return Endpoint{}, err
	}
	path := filepath.Join(configDir, "valvefm", "ctl.sock")
	return Endpoint{Network: "unix", Address: path}, nil
}

func Listen() (net.Listener, Endpoint, error) {
	ep, err := ResolveEndpoint()
	if err != nil {
		return nil, Endpoint{}, err
	}

	if err := os.MkdirAll(filepath.Dir(ep.Address), 0o755); err != nil {
		return nil, Endpoint{}, err
	}
	_ = os.Remove(ep.Address)

	listener, err := net.Listen(ep.Network, ep.Address)
	if err != nil {
		return nil, Endpoint{}, err
	}
	_ = os.Chmod(ep.Address, 0o600)
	return listener, ep, nil
}

func Cleanup(ep Endpoint) error {
	if ep.Address == "" {
		return nil
	}
	_ = os.Remove(ep.Address)
	return nil
}
