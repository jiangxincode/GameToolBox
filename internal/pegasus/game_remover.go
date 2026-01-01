package pegasus

// GameRemover is the backend for the "Pegasus Game Remover" feature.
// For now it mirrors the ROM manager generator behavior, but lives in its own
// file so future delete logic won't affect other features.
//
// Contract (temporary):
//   - Input: rootDir, games with Selected flag
//   - Output: GenerateResult (Created/Skipped/Failed) to keep UI wiring simple
//
// TODO: replace with real removal result once delete behavior is implemented.

type GameRemoveResult = GenerateResult

func RemoveSelectedGames(rootDir string, games []GameModel) GameRemoveResult {
	// Placeholder: reuse existing behavior until remover logic is defined.
	return GenerateSelectedFiles(rootDir, games)
}
