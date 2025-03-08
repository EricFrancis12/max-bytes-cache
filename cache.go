package main

import "sync"

type MaxBytesCache[T any] struct {
	mu       sync.Mutex
	data     map[string]T
	order    []string // Keeps track of the order of keys added to data
	maxBytes uint64   // The max number of bytes that data can hold
}

func NewMaxBytesCache[T any](maxBytes uint64) (*MaxBytesCache[T], error) {
	var t T
	_, err := calcSize(t)
	if err != nil {
		return nil, err
	}

	return &MaxBytesCache[T]{
		mu:       sync.Mutex{},
		data:     make(map[string]T),
		order:    []string{},
		maxBytes: maxBytes,
	}, nil
}

func (c *MaxBytesCache[T]) Get(key string) T {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.data[key]
}

func (c *MaxBytesCache[T]) Set(key string, value T) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = value
	c.order = append(c.order, key)

	if c.maxBytesExceeded() {
		return c.freeToMaxBytes()
	}
	return 0
}

func (c *MaxBytesCache[T]) maxBytesExceeded() bool {
	return c.dataSize() > c.maxBytes
}

func (c *MaxBytesCache[T]) freeToMaxBytes() uint64 {
	var count uint64 = 0
	for {
		if !c.maxBytesExceeded() {
			break
		}

		t := c.shift()
		if t != nil {
			count += mustCalcSize(t)
		}
	}
	return count
}

func (c *MaxBytesCache[T]) dataSize() uint64 {
	return mustCalcSize(c.data)
}

func (c *MaxBytesCache[T]) shift() *T {
	key := c.shiftKey()
	if key == nil {
		return nil
	}
	t, ok := c.data[*key]

	delete(c.data, *key)

	if !ok {
		return nil
	}
	return &t
}

func (c *MaxBytesCache[T]) shiftKey() *string {
	if len(c.order) == 0 {
		return nil
	}
	key := c.order[0]
	c.order = c.order[1:]
	return &key
}
