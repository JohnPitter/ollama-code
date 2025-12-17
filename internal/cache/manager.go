package cache

import (
	"sync"
	"time"
)

// Entry entrada de cache
type Entry struct {
	Value      interface{}
	Expiration time.Time
}

// Manager gerenciador de cache
type Manager struct {
	cache map[string]Entry
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewManager cria novo gerenciador
func NewManager(ttl time.Duration) *Manager {
	m := &Manager{
		cache: make(map[string]Entry),
		ttl:   ttl,
	}

	// Cleanup periódico
	go m.cleanupLoop()

	return m
}

// Get obtém valor do cache
func (m *Manager) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.Expiration) {
		return nil, false
	}

	return entry.Value, true
}

// Set define valor no cache
func (m *Manager) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache[key] = Entry{
		Value:      value,
		Expiration: time.Now().Add(m.ttl),
	}
}

// Delete remove do cache
func (m *Manager) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.cache, key)
}

// Clear limpa todo cache
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache = make(map[string]Entry)
}

// cleanupLoop loop de limpeza
func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.cleanup()
	}
}

// cleanup remove entradas expiradas
func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, entry := range m.cache {
		if now.After(entry.Expiration) {
			delete(m.cache, key)
		}
	}
}
