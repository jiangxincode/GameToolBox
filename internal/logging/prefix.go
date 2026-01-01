package logging

import (
	"runtime"
	"strings"
)

// PrefixFromCallerSkip returns a prefix based on the caller's package name and
// function name (derived from runtime information).
//
// It does NOT read go.mod, so it works the same in standalone built binaries.
//
// Example:
//
//	full: github.com/game_tool_box/internal/ui/pegasusui.NewGameRemoverView
//	->   "pegasusui.NewGameRemoverView:"
//
// skip=0 means the direct caller of PrefixFromCallerSkip.
// skip=1 means caller's caller, etc.
func PrefixFromCallerSkip(skip int) string {
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return ""
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}
	full := strings.TrimSpace(fn.Name())
	if full == "" {
		return ""
	}

	// Split: "path/pkg.Func" -> pkg, Func
	pkgPath := full
	funcName := ""
	if idx := strings.LastIndex(full, "."); idx >= 0 {
		pkgPath = full[:idx]
		funcName = full[idx+1:]
	} else {
		funcName = full
	}

	pkgName := pkgPath
	if idx := strings.LastIndex(pkgPath, "/"); idx >= 0 {
		pkgName = pkgPath[idx+1:]
	}

	pkgName = strings.TrimSpace(pkgName)
	funcName = strings.TrimSpace(funcName)

	if pkgName == "" {
		return ""
	}
	if funcName == "" {
		return pkgName + ":"
	}
	return pkgName + "." + funcName + ":"
}
