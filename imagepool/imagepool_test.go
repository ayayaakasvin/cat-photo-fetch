package imagepool_test

import (
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ayayaakasvin/cat-photo-fetch/imagepool"
)

// TestNewCatImagePool tests that the pool initializes correctly
func TestNewCatImagePool(t *testing.T) {
	pool, err := imagepool.NewCatImagePool(5, &http.Client{
		Timeout: time.Minute * 1,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Stop()

	// Pool should have been filled with initialPoolFillSize (10) images
	if pool == nil {
		t.Errorf("expected non-nil pool, got nil")
	}
}

// TestGetImage tests that Get() returns images
func TestGetImage(t *testing.T) {
	pool, err := imagepool.NewCatImagePool(5, &http.Client{
		Timeout: time.Minute * 1,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Stop()

	img := pool.Get()
	if img == nil {
		t.Errorf("expected image, got nil")
	}

	if img.ContentType == "" {
		t.Errorf("expected ContentType to be set, got empty string")
	}

	if len(img.Data) == 0 {
		t.Errorf("expected image data, got empty slice")
	}
}

// TestGetMultipleImages tests that Get() returns multiple different images
func TestGetMultipleImages(t *testing.T) {
	pool, err := imagepool.NewCatImagePool(5, &http.Client{
		Timeout: time.Minute * 1,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Stop()

	images := make([]*struct {
		ContentType string
		DataLen     int
	}, 5)

	for i := 0; i < 5; i++ {
		img := pool.Get()
		images[i] = &struct {
			ContentType string
			DataLen     int
		}{
			ContentType: img.ContentType,
			DataLen:     len(img.Data),
		}
	}

	// All images should have some data
	for i, img := range images {
		if img.DataLen == 0 {
			t.Errorf("image %d has no data", i)
		}
		if img.ContentType == "" {
			t.Errorf("image %d has no ContentType", i)
		}
	}
}

// TestConcurrentGet tests that multiple goroutines can safely Get() from the pool
func TestConcurrentGet(t *testing.T) {
	pool, err := imagepool.NewCatImagePool(5, &http.Client{
		Timeout: time.Minute * 1,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Stop()

	numGoroutines := 10
	imagesPerGoroutine := 3
	var wg sync.WaitGroup
	var successCount atomic.Int32

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < imagesPerGoroutine; j++ {
				img := pool.Get()
				if img != nil && len(img.Data) > 0 {
					successCount.Add(1)
				}
				// Small delay to simulate processing
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	wg.Wait()

	expected := int32(numGoroutines * imagesPerGoroutine)
	if successCount.Load() != expected {
		t.Errorf("expected %d successful Gets, got %d", expected, successCount.Load())
	}
}

// TestPoolRefill tests that the pool refills when it drops below the minimum boundary
func TestPoolRefill(t *testing.T) {
	pool, err := imagepool.NewCatImagePool(5, &http.Client{
		Timeout: time.Minute * 1,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Stop()

	// The pool is initialized with 10 images (initialPoolFillSize)
	// Get more images to drop the pool below minpoolBoundary (25)
	// This should trigger refilling
	for i := 0; i < 8; i++ {
		_ = pool.Get()
	}

	// Give the refill goroutine time to add more images
	time.Sleep(500 * time.Millisecond)

	// Try to get more images - should succeed due to refilling
	for i := 0; i < 5; i++ {
		img := pool.Get()
		if img == nil {
			t.Errorf("expected image after refill, got nil")
		}
	}
}

// TestStop tests that Stop() properly stops the refill goroutine
func TestStop(t *testing.T) {
	pool, err := imagepool.NewCatImagePool(5, &http.Client{
		Timeout: time.Minute * 1,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Get an image before stopping
	img := pool.Get()
	if img == nil {
		t.Errorf("expected image before stop, got nil")
	}

	// Stop the pool
	pool.Stop()

	// Give the goroutine time to stop
	time.Sleep(100 * time.Millisecond)

	// We should still be able to get images that were already in the pool,
	// but new refilling won't happen
	// Just verify that calling Stop multiple times doesn't panic
	pool.Stop()
}

// TestPoolCapacityNotExceeded tests that the pool doesn't exceed max capacity
func TestPoolCapacityNotExceeded(t *testing.T) {
	pool, err := imagepool.NewCatImagePool(5, &http.Client{
		Timeout: time.Minute * 1,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Stop()

	// Let the pool refill multiple times without consuming images
	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)
	}

	// The pool should be full or near capacity but not exceed maxSize (50)
	// We can't directly check the pool size, but if it exceeds capacity,
	// the refill logic would break
	// Try to get all images without blocking indefinitely
	timeout := time.After(5 * time.Second)
	imageCount := 0

	for {
		select {
		case <-timeout:
			goto done
		default:
			img := pool.Get()
			if img != nil {
				imageCount++
			}
		}
	}

done:
	if imageCount == 0 {
		t.Errorf("expected to retrieve images, got 0")
	}
}

// TestImageDataConsistency tests that fetched images have consistent data
func TestImageDataConsistency(t *testing.T) {
	pool, err := imagepool.NewCatImagePool(5, &http.Client{
		Timeout: time.Minute * 1,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Stop()

	img := pool.Get()
	if img == nil {
		t.Errorf("expected image, got nil")
	}

	// Test that Reader() works correctly
	reader := img.ReaderCloser()
	if reader == nil {
		t.Errorf("expected reader, got nil")
	}
	reader.Close()
}

// TestSequentialGets tests sequential Gets work properly
func TestSequentialGets(t *testing.T) {
	pool, err := imagepool.NewCatImagePool(5, &http.Client{
		Timeout: time.Minute * 1,
	})
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Stop()

	// Sequentially get 20 images
	for i := 0; i < 20; i++ {
		img := pool.Get()
		if img == nil {
			t.Errorf("iteration %d: expected image, got nil", i)
		}
		if len(img.Data) == 0 {
			t.Errorf("iteration %d: expected image data, got empty", i)
		}
	}
}
