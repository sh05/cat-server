package entities

import (
	"testing"
	"time"
)

func TestDirectoryListing_NewDirectoryListing(t *testing.T) {
	testTime := time.Now()

	entry1, _ := NewFileSystemEntry("file1.txt", "/path/file1.txt", 100, testTime, false, 0644)
	entry2, _ := NewFileSystemEntry("file2.txt", "/path/file2.txt", 200, testTime, false, 0644)

	tests := []struct {
		name      string
		path      string
		entries   []FileSystemEntry
		wantErr   bool
		wantCount int
	}{
		{
			name:      "valid directory listing",
			path:      "/valid/path",
			entries:   []FileSystemEntry{*entry1, *entry2},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "empty directory listing",
			path:      "/empty/path",
			entries:   []FileSystemEntry{},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:      "empty path should fail",
			path:      "",
			entries:   []FileSystemEntry{*entry1},
			wantErr:   true,
			wantCount: 0,
		},
		{
			name:      "nil entries should fail",
			path:      "/valid/path",
			entries:   nil,
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listing, err := NewDirectoryListing(tt.path, tt.entries)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if listing.Path() != tt.path {
				t.Errorf("Expected path %s, got %s", tt.path, listing.Path())
			}

			if listing.TotalCount() != tt.wantCount {
				t.Errorf("Expected count %d, got %d", tt.wantCount, listing.TotalCount())
			}

			entries := listing.Entries()
			if len(entries) != tt.wantCount {
				t.Errorf("Expected %d entries, got %d", tt.wantCount, len(entries))
			}
		})
	}
}

func TestDirectoryListing_FilterByType(t *testing.T) {
	testTime := time.Now()

	file1, _ := NewFileSystemEntry("file1.txt", "/path/file1.txt", 100, testTime, false, 0644)
	dir1, _ := NewFileSystemEntry("dir1", "/path/dir1", 0, testTime, true, 0755)
	file2, _ := NewFileSystemEntry("file2.txt", "/path/file2.txt", 200, testTime, false, 0644)

	listing, _ := NewDirectoryListing("/path", []FileSystemEntry{*file1, *dir1, *file2})

	t.Run("filter files only", func(t *testing.T) {
		files := listing.FilterByType(false) // files only
		if len(files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(files))
		}
	})

	t.Run("filter directories only", func(t *testing.T) {
		dirs := listing.FilterByType(true) // directories only
		if len(dirs) != 1 {
			t.Errorf("Expected 1 directory, got %d", len(dirs))
		}
	})
}

func TestDirectoryListing_SortByName(t *testing.T) {
	testTime := time.Now()

	file1, _ := NewFileSystemEntry("zebra.txt", "/path/zebra.txt", 100, testTime, false, 0644)
	file2, _ := NewFileSystemEntry("alpha.txt", "/path/alpha.txt", 200, testTime, false, 0644)
	file3, _ := NewFileSystemEntry("beta.txt", "/path/beta.txt", 300, testTime, false, 0644)

	listing, _ := NewDirectoryListing("/path", []FileSystemEntry{*file1, *file2, *file3})

	sorted := listing.SortByName()
	expectedOrder := []string{"alpha.txt", "beta.txt", "zebra.txt"}

	if len(sorted) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(sorted))
	}

	for i, entry := range sorted {
		if entry.Name() != expectedOrder[i] {
			t.Errorf("Expected %s at position %d, got %s", expectedOrder[i], i, entry.Name())
		}
	}
}
