package internal

import (
	"errors"
	"sync"
)

type Cache struct {
	mu sync.RWMutex
	m  map[string][]byte
}

func NewCache() *Cache {
	return &Cache{
		m: make(map[string][]byte),
	}
}

func (c *Cache) Set(key string, value []byte) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
	return nil
}

func (c *Cache) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, errors.New("key cannot be empty")
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, exists := c.m[key]; exists {
		return v, nil
	}
	return nil, errors.New("key not found: " + key)
}

func (c *Cache) Remove(key string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, key)
	return nil
}

func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.m)
}
