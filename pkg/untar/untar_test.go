package untar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// helper function to create a tar.gz file for testing
func createTestTarGz(content map[string]string, dirs []string, symlinks map[string]string) ([]byte, error) {
	buf := new(bytes.Buffer)
	gzWriter := gzip.NewWriter(buf)
	tarWriter := tar.NewWriter(gzWriter)

	// Handle directories
	for _, dir := range dirs {
		hdr := &tar.Header{
			Name:     dir,
			Mode:     0755,
			Typeflag: tar.TypeDir,
		}
		if err := tarWriter.WriteHeader(hdr); err != nil {
			return nil, err
		}
	}

	// Handle regular files
	for name, body := range content {
		hdr := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(body)),
		}
		if err := tarWriter.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tarWriter.Write([]byte(body)); err != nil {
			return nil, err
		}
	}

	// Handle symbolic links
	for name, target := range symlinks {
		hdr := &tar.Header{
			Name:     name,
			Linkname: target,
			Mode:     0777,
			Size:     0,
			Typeflag: tar.TypeSymlink,
		}
		if err := tarWriter.WriteHeader(hdr); err != nil {
			return nil, err
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func TestUntar(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		dirs     []string
		symlinks map[string]string
		content  map[string]string
		wantErr  bool
		validate func(t *testing.T, destPath string)
	}{
		{
			name: "NormalFiles",
			content: map[string]string{
				"file1.txt": "hello world",
				"file2.txt": "goodbye world",
			},
			wantErr: false,
			validate: func(t *testing.T, destPath string) {
				// Validate the contents of the extracted files
				file1Content, err := os.ReadFile(filepath.Join(destPath, "file1.txt"))
				if err != nil {
					t.Error(err)
				}
				if string(file1Content) != "hello world" {
					t.Errorf("Expected 'hello world', got '%s'", string(file1Content))
				}

				file2Content, err := os.ReadFile(filepath.Join(destPath, "file2.txt"))
				if err != nil {
					t.Error(err)
				}
				if string(file2Content) != "goodbye world" {
					t.Errorf("Expected 'goodbye world', got '%s'", string(file2Content))
				}
			},
		},
		/*{
			name:    "EmptyArchive",
			content: map[string]string{},
			wantErr: false,
			validate: func(t *testing.T, destPath string) {
				// Check if the directory does not exist
				if _, err := os.Stat(destPath); !os.IsNotExist(err) {
					t.Errorf("Destination directory should not exist for empty archive, but it does")
				}
			},
		},*/

		{
			name: "DirectoriesAndSymlinks",
			content: map[string]string{
				"dir1/file1.txt": "file in dir1",
				"dir2/file2.txt": "file in dir2",
			},
			dirs: []string{"dir1/", "dir2/"},
			symlinks: map[string]string{
				"symlink1": "dir1/file1.txt",
			},
			wantErr: false,
			validate: func(t *testing.T, destPath string) {
				// Validate directories
				for _, dir := range []string{"dir1", "dir2"} {
					if _, err := os.Stat(filepath.Join(destPath, dir)); os.IsNotExist(err) {
						t.Errorf("Directory %s was not created", dir)
					}
				}

				// Validate symlink
				symlinkPath := filepath.Join(destPath, "symlink1")
				targetPath, err := os.Readlink(symlinkPath)
				if err != nil {
					t.Errorf("Failed to read symlink: %v", err)
				}
				if targetPath != "dir1/file1.txt" {
					t.Errorf("Symlink points to %s, expected dir1/file1.txt", targetPath)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir, err := ioutil.TempDir("", "untar_test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			tarGzData, err := createTestTarGz(tc.content, tc.dirs, tc.symlinks)
			if err != nil {
				t.Fatalf("Failed to create test tar.gz data: %v", err)
			}

			tarGzFile := filepath.Join(tempDir, "test.tar.gz")
			if err := os.WriteFile(tarGzFile, tarGzData, 0600); err != nil {
				t.Fatalf("Failed to write test tar.gz file: %v", err)
			}

			err = Untar(tarGzFile, tempDir)
			if (err != nil) != tc.wantErr {
				t.Errorf("Untar() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.validate != nil {
				tc.validate(t, tempDir)
			}
		})
	}
}
