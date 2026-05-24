//go:build integration
// +build integration

// Tests for pool, due to http server issues, context cancel error can fail
package imagepool_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	imagepool "github.com/ayayaakasvin/cat-photo-fetch/image-pool"
)

// TestSave25ImagesToCatsDir verifies that the pool can provide 25 images and
// that each image can be written to the ./cats directory without error.
func TestSave25ImagesToCatsDir(t *testing.T) {
	dir := "./cats"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create cats directory: %v", err)
	}

	pool, err := imagepool.NewCatImagePool()
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Stop()

	for i := 1; i <= 25; i++ {
		img := pool.Get()
		if img == nil {
			t.Fatalf("failed to get image %d", i)
		}

		reader := img.Reader()
		data, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatalf("failed to read image %d data: %v", i, err)
		}

		if len(data) == 0 {
			t.Fatalf("image %d has empty data", i)
		}

		path := filepath.Join(dir, fmt.Sprintf("cat_%02d.jpg", i))
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatalf("failed to write image %d to file: %v", i, err)
		}

		stat, err := os.Stat(path)
		if err != nil {
			t.Fatalf("failed to stat saved image %d: %v", i, err)
		}
		if stat.Size() == 0 {
			t.Fatalf("saved image %d is empty", i)
		}
	}

	t.Logf("saved 25 images to %s", dir)
}
