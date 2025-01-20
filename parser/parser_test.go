package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetColumnIdx(t *testing.T) {
	tests := []struct {
		name       string
		header     []string
		columnName string
		expected   int
	}{
		{
			name:       "column exists",
			header:     []string{"COUNTRY ISO2 CODE", "SWIFT CODE", "NAME"},
			columnName: "SWIFT CODE",
			expected:   1,
		},
		{
			name:       "column does not exist",
			header:     []string{"COUNTRY ISO2 CODE", "SWIFT CODE", "NAME"},
			columnName: "INVALID",
			expected:   -1,
		},
		{
			name:       "empty header",
			header:     []string{},
			columnName: "SWIFT CODE",
			expected:   -1,
		},
		{
			name:       "case sensitive match",
			header:     []string{"COUNTRY ISO2 CODE", "Swift Code", "NAME"},
			columnName: "SWIFT CODE",
			expected:   -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getColumnIdx(tt.header, tt.columnName)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateColumnMap(t *testing.T) {
	tests := []struct {
		name        string
		header      []string
		expected    map[string]int
		expectError bool
	}{
		{
			name: "all required columns present",
			header: []string{
				"COUNTRY ISO2 CODE", "SWIFT CODE", "CODE TYPE", "NAME",
				"ADDRESS", "TOWN NAME", "COUNTRY NAME", "TIME ZONE",
			},
			expected: map[string]int{
				"COUNTRY ISO2 CODE": 0,
				"SWIFT CODE":        1,
				"CODE TYPE":         2,
				"NAME":              3,
				"ADDRESS":           4,
				"TOWN NAME":         5,
				"COUNTRY NAME":      6,
				"TIME ZONE":         7,
			},
			expectError: false,
		},
		{
			name: "missing required column",
			header: []string{
				"COUNTRY ISO2 CODE", "SWIFT CODE", "NAME",
				"ADDRESS", "TOWN NAME", "COUNTRY NAME",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:        "empty header",
			header:      []string{},
			expected:    nil,
			expectError: true,
		},
		{
			name: "columns in different order",
			header: []string{
				"TIME ZONE", "COUNTRY ISO2 CODE", "SWIFT CODE", "CODE TYPE",
				"NAME", "ADDRESS", "TOWN NAME", "COUNTRY NAME",
			},
			expected: map[string]int{
				"TIME ZONE":         0,
				"COUNTRY ISO2 CODE": 1,
				"SWIFT CODE":        2,
				"CODE TYPE":         3,
				"NAME":              4,
				"ADDRESS":           5,
				"TOWN NAME":         6,
				"COUNTRY NAME":      7,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := createColumnMap(tt.header)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestParseCSV_FileErrors(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		_, err := ParseCSV("nonexistent.csv")
		require.Error(t, err)
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "empty.csv")
		err := os.WriteFile(tmpFile, []byte(""), 0666)
		require.NoError(t, err, "Failed to create empty test file")

		_, err = ParseCSV(tmpFile)
		require.Error(t, err)
	})
}
