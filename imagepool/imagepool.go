package imagepool

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	catphotofetch "github.com/ayayaakasvin/cat-photo-fetch"
)

const (
	maxSize             = 64
	minpoolBoundary     = 32
	initialPoolFillSize = 16
	boundaryFillAmount  = 2
	defaultTimeout      = time.Second * 10
)

type CatImagePool struct {
	pool chan *catphotofetch.Image

	requestRefill chan struct{}
	stop          chan struct{}
	stopOnce      sync.Once
	mu            *sync.WaitGroup

	poolFillWorkers int

	c *http.Client
}

// Constructor of CatImagePool with size of 50
// creates n goroutine that will fill the pool up to max
func NewCatImagePool(poolFillWorkers int, c *http.Client) (*CatImagePool, error) {
	if poolFillWorkers < 1 {
		return nil, fmt.Errorf("Pool Fill Worker number is less 1: %d", poolFillWorkers)
	}

	if c == nil {
		c = &http.Client{
			Timeout: defaultTimeout,
		}
	}

	pool := &CatImagePool{
		pool:          make(chan *catphotofetch.Image, maxSize),
		requestRefill: make(chan struct{}, 1),
		stop:          make(chan struct{}),
		mu:            &sync.WaitGroup{},

		poolFillWorkers: poolFillWorkers,

		c: c,
	}

	if err := pool.fill(initialPoolFillSize); err != nil {
		return nil, fmt.Errorf("failed to init pool: fill error %s", err)
	}

	for i := 0; i < poolFillWorkers; i++ {
		pool.mu.Add(1)
		go pool.refillLoop(pool.mu)
	}

	return pool, nil
}

func (c *CatImagePool) fill(size int) error {
	for i := 0; i < size; i++ {
		img, err := catphotofetch.FetchViaCustomClient(c.c)
		if err != nil {
			return err
		}

		c.pool <- img
	}

	return nil
}

func (p *CatImagePool) refillLoop(mu *sync.WaitGroup) {
	defer mu.Done()

	for {
		select {
		case <-p.stop:
			return

		case <-p.requestRefill:
			needed := boundaryFillAmount

			for i := 0; i < needed; i++ {
				img, err := catphotofetch.FetchViaCustomClient(p.c)
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
		p.mu.Wait()
	})
}
