package pegasus

// ConfigManager is the backend for the "Pegasus Config Manager" feature.
// For now it intentionally mirrors the ROM manager behavior, but lives in its
// own file so future changes won't affect other generators.
//
// Contract:
//   - Input: rootDir (Pegasus root dir), games list with Selected flag
//   - Output: GenerateResult with Created/Skipped/Failed and Errors
//
// NOTE: This is intentionally a thin wrapper around the existing generator.
// We'll diverge the logic later.

type ConfigGenerateResult = GenerateResult

func GenerateSelectedConfigFiles(rootDir string, games []GameModel) ConfigGenerateResult {
	return GenerateSelectedFiles(rootDir, games)
}
