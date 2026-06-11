package romconverter

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewToolManager(t *testing.T) {
	tm, err := NewToolManager()
	if err != nil {
		t.Fatalf("NewToolManager failed: %v", err)
	}
	if tm == nil {
		t.Fatal("NewToolManager returned nil")
	}
	if tm.toolsDir == "" {
		t.Fatal("toolsDir is empty")
	}
}

func TestGetToolPath(t *testing.T) {
	tm, err := NewToolManager()
	if err != nil {
		t.Fatalf("NewToolManager failed: %v", err)
	}

	tests := []struct {
		tool     ConversionTool
		wantFile string
	}{
		{
			tool:     ConversionTool{ID: "nsz"},
			wantFile: "nsz.py",
		},
		{
			tool:     ConversionTool{ID: "4nxci"},
			wantFile: "4nxci",
		},
		{
			tool:     ConversionTool{ID: "tnes2ines"},
			wantFile: "tnes2ines.py",
		},
		{
			tool:     ConversionTool{ID: "nesromtool"},
			wantFile: "inestool.py",
		},
		{
			tool:     ConversionTool{ID: "fdsconv"},
			wantFile: "fds_header_cleaner.py",
		},
	}

	for _, tt := range tests {
		t.Run(tt.tool.ID, func(t *testing.T) {
			path := tm.GetToolPath(tt.tool)
			if path == "" {
				t.Error("GetToolPath returned empty string")
			}
			if !filepath.IsAbs(path) {
				t.Error("GetToolPath should return absolute path")
			}
			// Check that path ends with expected file
			// On Windows, nsz tool uses .bat extension
			expectedFile := tt.wantFile
			if runtime.GOOS == "windows" && tt.tool.ID == "nsz" {
				expectedFile = "nsz.bat"
			}
			if filepath.Base(path) != expectedFile && filepath.Base(path) != expectedFile+".exe" {
				t.Errorf("Expected path to end with %s, got %s", expectedFile, filepath.Base(path))
			}
		})
	}
}

func TestIsToolInstalled_NotInstalled(t *testing.T) {
	tm, err := NewToolManager()
	if err != nil {
		t.Fatalf("NewToolManager failed: %v", err)
	}

	tool := ConversionTool{
		ID:   "nsz",
		Name: "nsz",
	}

	// Clean up any existing installation for test
	toolPath := tm.GetToolPath(tool)
	if toolPath != "" {
		os.RemoveAll(filepath.Dir(toolPath))
	}

	installed := tm.IsToolInstalled(tool)
	if installed {
		t.Error("Tool should not be installed in fresh test environment")
	}
}
