package untar

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Untar extracts a tar.gz file asynchronously using goroutine pools.
func Untar(filePath, destPath string) error {
	fmt.Println("Untar: Starting extraction process")

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	var wg sync.WaitGroup
	maxWorkers := 10
	workerPool := make(chan struct{}, maxWorkers)
	errChan := make(chan error, maxWorkers)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			fmt.Println("Untar: Reached end of file")
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read next item: %v", err)
		}

		// Read file content for non-directory entries
		var fileContent []byte
		if header.Typeflag != tar.TypeDir && header.Typeflag != tar.TypeSymlink {
			fileContent, err = io.ReadAll(tarReader)
			if err != nil {
				return fmt.Errorf("failed to read file content: %v", err)
			}
		}

		fmt.Printf("Untar: Processing %s\n", header.Name)
		workerPool <- struct{}{}
		wg.Add(1)

		go func(header *tar.Header, fileContent []byte) {
			defer func() {
				<-workerPool
				wg.Done()
			}()

			destFilePath := filepath.Join(destPath, header.Name)

			// Ensure parent directories exist
			if err := os.MkdirAll(filepath.Dir(destFilePath), 0755); err != nil {
				errChan <- fmt.Errorf("failed to create directory: %v", err)
				return
			}

			if err := extractFile(header, destFilePath, fileContent); err != nil {
				fmt.Printf("Untar: Error extracting file %s: %v\n", destFilePath, err)
				errChan <- err
			}
		}(header, fileContent)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return fmt.Errorf("error in goroutine: %v", err)
		}
	}

	fmt.Println("Untar: Extraction process completed successfully")
	return nil
}

// extractFile handles the extraction of a single file from the tar archive.
func extractFile(header *tar.Header, destFilePath string, fileContent []byte) error {
	switch header.Typeflag {
	case tar.TypeDir:
		// Directory creation is handled in the main goroutine
		return nil
	case tar.TypeSymlink:
		return os.Symlink(header.Linkname, destFilePath)
	default:
		destFile, err := os.Create(destFilePath)
		if err != nil {
			return err
		}
		defer destFile.Close()
		if _, err := destFile.Write(fileContent); err != nil {
			return err
		}
	}
	return nil
}
