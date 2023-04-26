package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Cache is the main struct representing the in-memory cache
type Cache struct {
	mu        sync.RWMutex
	items     map[string]*cacheItem
	maxSize   int
	expireTTL time.Duration
}

// cacheItem represents an item in the cache with its value and expiration time
type cacheItem struct {
	value      interface{}
	expireTime time.Time
}

// cacheMap represents a map of caches, where each cache is identified by a string key
type cacheMap struct {
	mu     sync.RWMutex
	caches map[string]*Cache
}

var (
	cacheMapInstance = cacheMap{
		caches: make(map[string]*Cache),
	}
)

// NewCache creates a new Cache with the given maximum size and expiration TTL
func NewCache(maxSize int, expireTTL time.Duration) *Cache {
	return &Cache{
		items:     make(map[string]*cacheItem),
		maxSize:   maxSize,
		expireTTL: expireTTL,
	}
}

// WriteJSONResponse writes a JSON response to the HTTP response writer with the given status code and response body
func WriteJSONResponse(w http.ResponseWriter, statusCode int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(body)
}

// Set stores a value in the cache with the given key
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the cache is full and delete the oldest item if necessary
	if len(c.items) >= c.maxSize {
		c.deleteOldestItem()
	}

	// Set the new item in the cache with its expiration time
	expireTime := time.Now().Add(c.expireTTL)
	c.items[key] = &cacheItem{
		value:      value,
		expireTime: expireTime,
	}
}

// Get retrieves a value from the cache given a key, returns nil if not found or expired
func (c *Cache) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if ok && item.expireTime.After(time.Now()) {
		return item.value
	}

	return nil
}

// Delete removes a value from the cache given a key
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// deleteOldestItem removes the oldest item from the cache based on its expiration time
func (c *Cache) deleteOldestItem() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestTime.IsZero() || item.expireTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.expireTime
		}
	}

	delete(c.items, oldestKey)
}

// HandleCreateCache is the handler for creating a new cache with a given maximum size and expiration TTL
func HandleCreateCache(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	maxSize := vars["maxSize"]
	expireTTL := vars["expireTTL"]

	maxSizeInt := 0
	expireTTLInt := 0

	fmt.Sscanf(maxSize, "%d", &maxSizeInt)
	fmt.Sscanf(expireTTL, "%d", &expireTTLInt)

	if maxSizeInt <= 0 || expireTTLInt <= 0 {
		http.Error(w, "Invalid input. Maximum size and expiration TTL must be greater than 0.", http.StatusBadRequest)
		return
	}

	cache := NewCache(maxSizeInt, time.Duration(expireTTLInt)*time.Second)

	cacheID := fmt.Sprintf("cache%d", time.Now().UnixNano())

	cacheMapInstance.mu.Lock()
	defer cacheMapInstance.mu.Unlock()

	cacheMapInstance.caches[cacheID] = cache

	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Cache created with ID: %s", cacheID),
	}

	WriteJSONResponse(w, http.StatusCreated, response)
}

// In this implementation, a unique cache ID is generated based on the current UnixNano timestamp, and the created cache is stored in the cacheMapInstance
// which is a global instance of cacheMap that holds multiple caches identified by their cache IDs. Also,
// the WriteJSONResponse function is used to write the J6SON response to the HTTP response writer with the appropriate status code and response body.
// Plz you have to the Gorilla Mux library installed in your Go environment. You can install it using this command:go get -u github.com/gorilla/mux
// Thank you.
