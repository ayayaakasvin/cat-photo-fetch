package catphotofetch

import (
	"log"
	"sync"
)

type CatCache struct {
	Data 		[]byte
	mu			sync.RWMutex
}

var cache = &CatCache{}

func RefreshImage() error {
	data, err := FetchRandomPhoto()
	if err != nil {
		cache = nil
		return err
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	cache.Data = data

	return nil
}

func GetCatImage() ([]byte, error) {
	cache.mu.RLock()
	data := cache.Data
	cache.mu.RUnlock()

	if data == nil {
		if err := RefreshImage(); err != nil {
			return nil, err
		}
		cache.mu.RLock()
		defer cache.mu.RUnlock()
		return cache.Data, nil
	}

	go func() {
		if err := RefreshImage(); err != nil {
			log.Println("background refresh failed:", err)
		}
	}()

	return data, nil
}