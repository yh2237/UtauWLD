//go:build windows

package worldrender

import (
	"syscall"
	"testing"
)

func TestWindowsWorldEngineCloseForgetsExtras(t *testing.T) {
	dll := &syscall.DLL{}
	windowsWorldExtrasMu.Lock()
	windowsWorldExtras[dll] = &windowsWorldProcs{}
	windowsWorldExtrasMu.Unlock()

	engine := &windowsWorldEngine{dll: dll}
	_ = engine.Close() // FreeLibrary(0) fails; the extras entry must still be removed.

	windowsWorldExtrasMu.Lock()
	_, found := windowsWorldExtras[dll]
	windowsWorldExtrasMu.Unlock()
	if found {
		t.Fatal("closed engine extras were retained")
	}
}
