//go:build darwin

package app

import "codex-proxy/app/darwin"

func newPlatform() Platform {
	return &darwin.Darwin{}
}
