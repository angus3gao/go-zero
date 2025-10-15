package redis

import (
	_ "embed"
)

var (
	//go:embed getorsetscript.lua
	getOrSetLuaScript string
	getOrSetScript    = NewScript(getOrSetLuaScript)

	//go:embed compareandswapscript.lua
	compareAndSwapLuaScript string
	compareAndSwapscript    = NewScript(compareAndSwapLuaScript)

	//go:embed compareanddelscript.lua
	compareAndDelLuaScript string
	compareAndDelscript    = NewScript(compareAndDelLuaScript)

	//go:embed zcomparehigherscript.lua
	zCompareHigherLuaScript string
	zCompareHigherScript    = NewScript(zCompareHigherLuaScript)
)
