package catphotofetch_test

import (
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	catphotofetch "github.com/ayayaakasvin/cat-photo-fetch"
)

func TestFetch(t *testing.T) {
	start := time.Now()
	photoData, err := catphotofetch.FetchViaCustomClient(&http.Client{Timeout: time.Minute * 1})
	if err != nil {
		t.Fatalf("failed to fetch file: %s", err)
	}

	file, err := os.Create("cat.jpg")
	if err != nil {
		t.Fatalf("failed to create file: %s", err)
	}

	_, err = io.Copy(file, photoData.ReaderCloser())
	if err != nil {
		t.Fatalf("failed to fetch file: %s", err)
	}

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("failed to get file stat: %s", err)
	}

	t.Logf("Stats: Size = %d | Data = %d\nTime = %v", stat.Size(), len(photoData.Data), time.Since(start).Seconds())
}
