package metadata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseRealPSVITAFile tests parsing the actual PSVITA testdata file
func TestParseRealPSVITAFile(t *testing.T) {
	testFile := filepath.Join("..", "..", "..", "internal", "pegasus", "testdata", "PSVITA", "metadata.pegasus.txt")
	
	// Check if file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testFile)
	}

	doc, err := ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Verify collection was parsed
	collections := doc.Collections()
	if len(collections) == 0 {
		t.Error("Expected at least 1 collection, got 0")
	} else {
		c := collections[0]
		if c.Name == "" {
			t.Error("Collection name is empty")
		}
		t.Logf("Found collection: %s (shortname: %s)", c.Name, c.ShortName)
	}

	// Verify games were parsed
	games := doc.Games()
	if len(games) == 0 {
		t.Error("Expected at least 1 game, got 0")
	} else {
		t.Logf("Found %d games", len(games))
		for i, g := range games {
			if g.GameName == "" {
				t.Errorf("Game %d has empty name", i)
			}
			if g.FileName == "" {
				t.Errorf("Game %d (%s) has empty filename", i, g.GameName)
			}
		}
	}

	// Verify round-trip preserves content
	rendered := doc.Render()
	if len(rendered) == 0 {
		t.Error("Rendered output is empty")
	}

	// Verify collection line is preserved
	if !strings.Contains(rendered, "collection:") {
		t.Error("Collection line not preserved in rendered output")
	}

	// Parse the rendered output and verify it produces same results
	doc2 := Parse(rendered)
	games2 := doc2.Games()
	collections2 := doc2.Collections()

	if len(games2) != len(games) {
		t.Errorf("Round-trip changed game count: %d -> %d", len(games), len(games2))
	}
	if len(collections2) != len(collections) {
		t.Errorf("Round-trip changed collection count: %d -> %d", len(collections), len(collections2))
	}
}
