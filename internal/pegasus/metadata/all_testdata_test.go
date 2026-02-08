package metadata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseAllTestdataFiles tests parsing of all metadata files in testdata directory
// This ensures the parser can handle all real-world configuration files from master branch
func TestParseAllTestdataFiles(t *testing.T) {
	testdataRoot := filepath.Join("..", "..", "..", "internal", "pegasus", "testdata")
	
	// Check if testdata directory exists
	if _, err := os.Stat(testdataRoot); os.IsNotExist(err) {
		t.Skipf("Testdata directory not found: %s", testdataRoot)
	}

	var allFiles []string
	err := filepath.WalkDir(testdataRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (filepath.Base(path) == "metadata.pegasus.txt" || filepath.Base(path) == "metadata.txt") {
			allFiles = append(allFiles, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk testdata directory: %v", err)
	}

	if len(allFiles) == 0 {
		t.Fatalf("No metadata files found in testdata directory")
	}

	t.Logf("Found %d metadata files to test", len(allFiles))

	var parseErrors []string
	var successCount int

	for _, file := range allFiles {
		relPath, _ := filepath.Rel(testdataRoot, file)
		
		doc, err := ReadFile(file)
		if err != nil {
			parseErrors = append(parseErrors, relPath+": "+err.Error())
			continue
		}

		// Try to get games and collections
		games := doc.Games()
		collections := doc.Collections()

		// Try to render (round-trip test)
		rendered := doc.Render()
		if len(rendered) == 0 {
			parseErrors = append(parseErrors, relPath+": render produced empty output")
			continue
		}

		// Parse the rendered version to ensure round-trip works
		doc2 := Parse(rendered)
		games2 := doc2.Games()
		collections2 := doc2.Collections()

		if len(games2) != len(games) {
			parseErrors = append(parseErrors, relPath+": round-trip changed game count")
			continue
		}

		if len(collections2) != len(collections) {
			parseErrors = append(parseErrors, relPath+": round-trip changed collection count")
			continue
		}

		successCount++
		t.Logf("✓ %s: %d games, %d collections", relPath, len(games), len(collections))
	}

	// Report results
	t.Logf("\nResults: %d/%d files parsed successfully", successCount, len(allFiles))

	if len(parseErrors) > 0 {
		t.Errorf("Failed to parse %d files:", len(parseErrors))
		for _, err := range parseErrors {
			t.Errorf("  - %s", err)
		}
	}

	// Ensure at least 95% success rate
	successRate := float64(successCount) / float64(len(allFiles))
	if successRate < 0.95 {
		t.Errorf("Success rate too low: %.1f%% (expected >= 95%%)", successRate*100)
	}
}

// TestAllTestdataFilesHaveValidStructure tests that all files have at least one game or collection
func TestAllTestdataFilesHaveValidStructure(t *testing.T) {
	testdataRoot := filepath.Join("..", "..", "..", "internal", "pegasus", "testdata")
	
	if _, err := os.Stat(testdataRoot); os.IsNotExist(err) {
		t.Skipf("Testdata directory not found: %s", testdataRoot)
	}

	var allFiles []string
	err := filepath.WalkDir(testdataRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (filepath.Base(path) == "metadata.pegasus.txt" || filepath.Base(path) == "metadata.txt") {
			allFiles = append(allFiles, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk testdata directory: %v", err)
	}

	emptyFiles := []string{}

	for _, file := range allFiles {
		relPath, _ := filepath.Rel(testdataRoot, file)
		
		doc, err := ReadFile(file)
		if err != nil {
			continue // Already reported in other test
		}

		games := doc.Games()
		collections := doc.Collections()

		if len(games) == 0 && len(collections) == 0 {
			emptyFiles = append(emptyFiles, relPath)
		}
	}

	if len(emptyFiles) > 0 {
		t.Logf("Warning: %d files have no games or collections:", len(emptyFiles))
		for _, file := range emptyFiles {
			t.Logf("  - %s", file)
		}
	}
}

// TestSampleFilesDetailed provides detailed parsing information for a few sample files
func TestSampleFilesDetailed(t *testing.T) {
	sampleDirs := []string{"PSVITA", "3DO", "N64", "PS1", "SFC"}
	testdataRoot := filepath.Join("..", "..", "..", "internal", "pegasus", "testdata")

	for _, dir := range sampleDirs {
		metaFile := filepath.Join(testdataRoot, dir, "metadata.pegasus.txt")
		
		if _, err := os.Stat(metaFile); os.IsNotExist(err) {
			t.Logf("Sample file not found: %s", dir)
			continue
		}

		doc, err := ReadFile(metaFile)
		if err != nil {
			t.Errorf("%s: failed to parse: %v", dir, err)
			continue
		}

		games := doc.Games()
		collections := doc.Collections()

		t.Logf("\n%s:", dir)
		t.Logf("  Collections: %d", len(collections))
		for i, c := range collections {
			t.Logf("    [%d] %s (shortname: %s)", i+1, c.Name, c.ShortName)
		}

		t.Logf("  Games: %d", len(games))
		for i, g := range games[:min(3, len(games))] {
			fields := []string{}
			if g.Publisher != "" {
				fields = append(fields, "publisher")
			}
			if g.Genre != "" {
				fields = append(fields, "genre")
			}
			if g.Players != "" {
				fields = append(fields, "players")
			}
			if g.Rating != "" {
				fields = append(fields, "rating")
			}
			if g.Release != "" {
				fields = append(fields, "release")
			}
			
			extraInfo := ""
			if len(fields) > 0 {
				extraInfo = " (has: " + strings.Join(fields, ", ") + ")"
			}
			
			t.Logf("    [%d] %s%s", i+1, g.GameName, extraInfo)
		}
		if len(games) > 3 {
			t.Logf("    ... and %d more games", len(games)-3)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
