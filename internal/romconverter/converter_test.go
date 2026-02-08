package romconverter

import (
"testing"
)

func TestGetSwitchFormats(t *testing.T) {
formats := GetSwitchFormats()
if len(formats) != 4 {
t.Errorf("expected 4 formats, got %d", len(formats))
}
expectedFormats := []SwitchFormat{FormatXCI, FormatNSP, FormatNSZ, FormatXCZ}
for i, format := range formats {
if format != expectedFormats[i] {
t.Errorf("format[%d]: expected %s, got %s", i, expectedFormats[i], format)
}
}
}

func TestGetSwitchTools(t *testing.T) {
tools := GetSwitchTools()
if len(tools) < 2 {
t.Errorf("expected at least 2 tools, got %d", len(tools))
}
for _, tool := range tools {
if tool.Name == "" {
t.Error("tool name should not be empty")
}
if tool.ID == "" {
t.Error("tool ID should not be empty")
}
}
}

func TestConverterValidate(t *testing.T) {
converter := NewConverter(
ConversionTool{ID: "test", Name: "Test Tool"},
"",
"",
FormatNSP,
FormatXCI,
nil,
)

err := converter.Validate()
if err == nil {
t.Error("expected validation error for empty directories")
}
}
