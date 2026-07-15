//go:build windows

package app

import "codex-proxy/app/windows"

func newPlatform() Platform {
	return &windows.Windows{}
}
