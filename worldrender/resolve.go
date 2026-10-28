//go:build !js

package worldrender

import "github.com/yh2237/UtauWLD/runtimepath"

// ResolveEnginePathは、WORLDのライブラリをruntimeから探す。
func ResolveEnginePath() (string, error) {
	return runtimepath.Find("", runtimepath.LibraryName("utautts-world-engine"), false)
}
