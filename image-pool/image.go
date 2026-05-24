package imagepool

import (
	"fmt"
	"sync"

	catphotofetch "github.com/ayayaakasvin/cat-photo-fetch"
)

const (
	maxSize             = 50
	minpoolBoundary     = 25
	initialPoolFillSize = 10
	boundaryFillAmount  = 2
)

type CatImagePool struct {
	pool chan *catphotofetch.Image

	requestRefill chan struct{}
	stop          chan struct{}
	stopOnce      sync.Once
}

// Constructor of CatImagePool with size of 50
// creates 1 goroutine that will fill the pool up to max
func NewCatImagePool() (*CatImagePool, error) {
	pool := &CatImagePool{
		pool:          make(chan *catphotofetch.Image, maxSize),
		requestRefill: make(chan struct{}, 1),
		stop:          make(chan struct{}),
	}

	if err := pool.fill(initialPoolFillSize); err != nil {
		return nil, fmt.Errorf("failed to init pool: fill error %s", err)
	}

	go pool.refillLoop()

	return pool, nil
}

func (c *CatImagePool) fill(size int) error {
	for i := 0; i < size; i++ {
		img, err := catphotofetch.FetchRandomPhoto()
		if err != nil {
			return err
		}

		c.pool <- img
	}

	return nil
}

func (p *CatImagePool) refillLoop() {
	for {
		select {
		case <-p.stop:
			return

		case <-p.requestRefill:
			needed := boundaryFillAmount

			for i := 0; i < needed; i++ {
				// respect max capacity
				if len(p.pool) >= maxSize {
					break
				}

				img, err := catphotofetch.FetchRandomPhoto()
				if err != nil {
					continue
				}

				p.pool <- img
			}
		}
	}
}

func (p *CatImagePool) Get() *catphotofetch.Image {
	img := <-p.pool

	// non-blocking signal
	if len(p.pool) < minpoolBoundary {
		select {
		case p.requestRefill <- struct{}{}:
		default:
		}
	}

	return img
}

// Stop gracefully stops the pool's refill goroutine
func (p *CatImagePool) Stop() {
	p.stopOnce.Do(func() {
		close(p.stop)
	})
}
