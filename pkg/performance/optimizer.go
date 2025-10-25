package performance

import (
	"context"
	"runtime"
	"sync"
	"time"
	"fmt"
	"compress/gzip"
	"bytes"
	"io"
)

// ResourceOptimizer provides resource optimization capabilities
type ResourceOptimizer struct {
	memoryPool    *MemoryPool
	compressionPool *CompressionPool
	cacheManager  *CacheManager
	config        OptimizerConfig
	mu            sync.RWMutex
}

// OptimizerConfig contains configuration for the resource optimizer
type OptimizerConfig struct {
	EnableMemoryPooling   bool
	EnableCompression     bool
	EnableCaching         bool
	MaxMemoryUsage        uint64
	CompressionThreshold  int
	CacheSize             int
	GCInterval            time.Duration
}

// MemoryPool manages reusable memory buffers
type MemoryPool struct {
	pools map[int]*sync.Pool
	mu    sync.RWMutex
}

// CompressionPool manages compression/decompression resources
type CompressionPool struct {
	gzipWriters *sync.Pool
	gzipReaders *sync.Pool
}

// CacheManager manages cached resources
type CacheManager struct {
	cache     map[string]*CacheEntry
	maxSize   int
	currentSize int
	mu        sync.RWMutex
}

// CacheEntry represents a cached resource
type CacheEntry struct {
	Data      []byte
	Size      int
	LastUsed  time.Time
	UseCount  int64
}

// OptimizationResult represents the result of an optimization operation
type OptimizationResult struct {
	OriginalSize   int
	OptimizedSize  int
	CompressionRatio float64
	TimeSaved      time.Duration
	MemorySaved    uint64
}

// NewResourceOptimizer creates a new resource optimizer
func NewResourceOptimizer(config OptimizerConfig) *ResourceOptimizer {
	optimizer := &ResourceOptimizer{
		config: config,
	}
	
	if config.EnableMemoryPooling {
		optimizer.memoryPool = NewMemoryPool()
	}
	
	if config.EnableCompression {
		optimizer.compressionPool = NewCompressionPool()
	}
	
	if config.EnableCaching {
		optimizer.cacheManager = NewCacheManager(config.CacheSize)
	}
	
	// Start background optimization tasks
	if config.GCInterval > 0 {
		go optimizer.backgroundOptimization(config.GCInterval)
	}
	
	return optimizer
}

// OptimizeData optimizes data using available optimization techniques
func (ro *ResourceOptimizer) OptimizeData(key string, data []byte) ([]byte, OptimizationResult, error) {
	result := OptimizationResult{
		OriginalSize: len(data),
	}
	
	optimizedData := data
	var err error
	
	// Check cache first
	if ro.config.EnableCaching {
		if cached := ro.cacheManager.Get(key); cached != nil {
			result.OptimizedSize = len(cached)
			result.CompressionRatio = float64(result.OriginalSize) / float64(result.OptimizedSize)
			return cached, result, nil
		}
	}
	
	// Apply compression if enabled and data is large enough
	if ro.config.EnableCompression && len(data) > ro.config.CompressionThreshold {
		start := time.Now()
		optimizedData, err = ro.compressData(data)
		if err != nil {
			return data, result, fmt.Errorf("compression failed: %w", err)
		}
		result.TimeSaved = time.Since(start)
	}
	
	result.OptimizedSize = len(optimizedData)
	result.CompressionRatio = float64(result.OriginalSize) / float64(result.OptimizedSize)
	
	// Cache the optimized data
	if ro.config.EnableCaching {
		ro.cacheManager.Put(key, optimizedData)
	}
	
	return optimizedData, result, nil
}

// DeoptimizeData reverses optimization (e.g., decompression)
func (ro *ResourceOptimizer) DeoptimizeData(key string, data []byte) ([]byte, error) {
	// Check if data is compressed
	if ro.config.EnableCompression && ro.isCompressed(data) {
		return ro.decompressData(data)
	}
	
	return data, nil
}

// GetBuffer gets a reusable buffer from the memory pool
func (ro *ResourceOptimizer) GetBuffer(size int) []byte {
	if !ro.config.EnableMemoryPooling || ro.memoryPool == nil {
		return make([]byte, size)
	}
	
	return ro.memoryPool.Get(size)
}

// PutBuffer returns a buffer to the memory pool
func (ro *ResourceOptimizer) PutBuffer(buf []byte) {
	if !ro.config.EnableMemoryPooling || ro.memoryPool == nil {
		return
	}
	
	ro.memoryPool.Put(buf)
}

// OptimizeMemoryUsage performs memory optimization
func (ro *ResourceOptimizer) OptimizeMemoryUsage() OptimizationResult {
	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	
	// Force garbage collection
	runtime.GC()
	runtime.GC()
	
	// Clean up caches if memory usage is high
	if ro.config.EnableCaching && memBefore.Alloc > ro.config.MaxMemoryUsage {
		ro.cacheManager.Cleanup()
	}
	
	// Clean up memory pools
	if ro.config.EnableMemoryPooling {
		ro.memoryPool.Cleanup()
	}
	
	runtime.ReadMemStats(&memAfter)
	
	return OptimizationResult{
		MemorySaved: memBefore.Alloc - memAfter.Alloc,
	}
}

// GetOptimizationStats returns optimization statistics
func (ro *ResourceOptimizer) GetOptimizationStats() OptimizationStats {
	stats := OptimizationStats{}
	
	if ro.config.EnableCaching {
		cacheStats := ro.cacheManager.GetStats()
		stats.CacheHits = cacheStats.Hits
		stats.CacheMisses = cacheStats.Misses
		stats.CacheSize = cacheStats.Size
	}
	
	if ro.config.EnableMemoryPooling {
		poolStats := ro.memoryPool.GetStats()
		stats.PoolHits = poolStats.Hits
		stats.PoolMisses = poolStats.Misses
		stats.PoolSize = poolStats.Size
	}
	
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	stats.MemoryUsage = memStats.Alloc
	stats.GCCount = memStats.NumGC
	
	return stats
}

// OptimizationStats represents optimization statistics
type OptimizationStats struct {
	CacheHits    int64
	CacheMisses  int64
	CacheSize    int
	PoolHits     int64
	PoolMisses   int64
	PoolSize     int
	MemoryUsage  uint64
	GCCount      uint32
}

// Internal methods

func (ro *ResourceOptimizer) compressData(data []byte) ([]byte, error) {
	if ro.compressionPool == nil {
		return ro.compressDataDirect(data)
	}
	
	// Get writer from pool
	writer := ro.compressionPool.GetWriter()
	defer ro.compressionPool.PutWriter(writer)
	
	var buf bytes.Buffer
	writer.Reset(&buf)
	
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

func (ro *ResourceOptimizer) decompressData(data []byte) ([]byte, error) {
	if ro.compressionPool == nil {
		return ro.decompressDataDirect(data)
	}
	
	// Get reader from pool
	reader := ro.compressionPool.GetReader()
	defer ro.compressionPool.PutReader(reader)
	
	buf := bytes.NewReader(data)
	err := reader.Reset(buf)
	if err != nil {
		return nil, err
	}
	
	var result bytes.Buffer
	_, err = io.Copy(&result, reader)
	if err != nil {
		return nil, err
	}
	
	return result.Bytes(), nil
}

func (ro *ResourceOptimizer) compressDataDirect(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

func (ro *ResourceOptimizer) decompressDataDirect(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	var result bytes.Buffer
	_, err = io.Copy(&result, reader)
	if err != nil {
		return nil, err
	}
	
	return result.Bytes(), nil
}

func (ro *ResourceOptimizer) isCompressed(data []byte) bool {
	// Check for gzip magic number
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

func (ro *ResourceOptimizer) backgroundOptimization(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for range ticker.C {
		// Perform background optimization tasks
		ro.OptimizeMemoryUsage()
		
		// Clean up expired cache entries
		if ro.config.EnableCaching {
			ro.cacheManager.CleanupExpired()
		}
	}
}

// MemoryPool implementation

func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		pools: make(map[int]*sync.Pool),
	}
}

func (mp *MemoryPool) Get(size int) []byte {
	// Round up to nearest power of 2 for better pooling
	poolSize := nextPowerOf2(size)
	
	mp.mu.RLock()
	pool, exists := mp.pools[poolSize]
	mp.mu.RUnlock()
	
	if !exists {
		mp.mu.Lock()
		// Double-check after acquiring write lock
		if pool, exists = mp.pools[poolSize]; !exists {
			pool = &sync.Pool{
				New: func() interface{} {
					return make([]byte, poolSize)
				},
			}
			mp.pools[poolSize] = pool
		}
		mp.mu.Unlock()
	}
	
	buf := pool.Get().([]byte)
	return buf[:size] // Return slice of requested size
}

func (mp *MemoryPool) Put(buf []byte) {
	if cap(buf) == 0 {
		return
	}
	
	poolSize := cap(buf)
	
	mp.mu.RLock()
	pool, exists := mp.pools[poolSize]
	mp.mu.RUnlock()
	
	if exists {
		// Reset buffer before returning to pool
		for i := range buf {
			buf[i] = 0
		}
		pool.Put(buf[:cap(buf)])
	}
}

func (mp *MemoryPool) Cleanup() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	// Clear all pools to free memory
	mp.pools = make(map[int]*sync.Pool)
}

func (mp *MemoryPool) GetStats() PoolStats {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	
	return PoolStats{
		Size: len(mp.pools),
	}
}

type PoolStats struct {
	Hits   int64
	Misses int64
	Size   int
}

// CompressionPool implementation

func NewCompressionPool() *CompressionPool {
	return &CompressionPool{
		gzipWriters: &sync.Pool{
			New: func() interface{} {
				return gzip.NewWriter(nil)
			},
		},
		gzipReaders: &sync.Pool{
			New: func() interface{} {
				reader, _ := gzip.NewReader(nil)
				return reader
			},
		},
	}
}

func (cp *CompressionPool) GetWriter() *gzip.Writer {
	return cp.gzipWriters.Get().(*gzip.Writer)
}

func (cp *CompressionPool) PutWriter(writer *gzip.Writer) {
	writer.Reset(nil)
	cp.gzipWriters.Put(writer)
}

func (cp *CompressionPool) GetReader() *gzip.Reader {
	return cp.gzipReaders.Get().(*gzip.Reader)
}

func (cp *CompressionPool) PutReader(reader *gzip.Reader) {
	reader.Reset(nil)
	cp.gzipReaders.Put(reader)
}

// CacheManager implementation

func NewCacheManager(maxSize int) *CacheManager {
	return &CacheManager{
		cache:   make(map[string]*CacheEntry),
		maxSize: maxSize,
	}
}

func (cm *CacheManager) Get(key string) []byte {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	if entry, exists := cm.cache[key]; exists {
		entry.LastUsed = time.Now()
		entry.UseCount++
		return entry.Data
	}
	
	return nil
}

func (cm *CacheManager) Put(key string, data []byte) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	// Check if we need to evict entries
	if len(cm.cache) >= cm.maxSize {
		cm.evictLRU()
	}
	
	entry := &CacheEntry{
		Data:     make([]byte, len(data)),
		Size:     len(data),
		LastUsed: time.Now(),
		UseCount: 1,
	}
	copy(entry.Data, data)
	
	cm.cache[key] = entry
	cm.currentSize += entry.Size
}

func (cm *CacheManager) Cleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	// Remove half of the least recently used entries
	entries := make([]*CacheEntry, 0, len(cm.cache))
	keys := make([]string, 0, len(cm.cache))
	
	for key, entry := range cm.cache {
		entries = append(entries, entry)
		keys = append(keys, key)
	}
	
	// Sort by last used time
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].LastUsed.After(entries[j].LastUsed) {
				entries[i], entries[j] = entries[j], entries[i]
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	
	// Remove half of the entries
	removeCount := len(entries) / 2
	for i := 0; i < removeCount; i++ {
		delete(cm.cache, keys[i])
		cm.currentSize -= entries[i].Size
	}
}

func (cm *CacheManager) CleanupExpired() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	expireTime := time.Now().Add(-1 * time.Hour) // Expire after 1 hour
	
	for key, entry := range cm.cache {
		if entry.LastUsed.Before(expireTime) {
			delete(cm.cache, key)
			cm.currentSize -= entry.Size
		}
	}
}

func (cm *CacheManager) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	
	for key, entry := range cm.cache {
		if oldestKey == "" || entry.LastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastUsed
		}
	}
	
	if oldestKey != "" {
		entry := cm.cache[oldestKey]
		delete(cm.cache, oldestKey)
		cm.currentSize -= entry.Size
	}
}

func (cm *CacheManager) GetStats() CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	return CacheStats{
		Size: len(cm.cache),
	}
}

type CacheStats struct {
	Hits   int64
	Misses int64
	Size   int
}

// Utility functions

func nextPowerOf2(n int) int {
	if n <= 0 {
		return 1
	}
	
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	
	return n
}

// DefaultOptimizerConfig returns a default optimizer configuration
func DefaultOptimizerConfig() OptimizerConfig {
	return OptimizerConfig{
		EnableMemoryPooling:  true,
		EnableCompression:    true,
		EnableCaching:        true,
		MaxMemoryUsage:       100 * 1024 * 1024, // 100MB
		CompressionThreshold: 1024,               // 1KB
		CacheSize:            1000,
		GCInterval:           5 * time.Minute,
	}
}

// Global optimizer instance
var globalOptimizer *ResourceOptimizer

func init() {
	globalOptimizer = NewResourceOptimizer(DefaultOptimizerConfig())
}

// Global functions for convenience

// OptimizeData optimizes data using the global optimizer
func OptimizeData(key string, data []byte) ([]byte, OptimizationResult, error) {
	return globalOptimizer.OptimizeData(key, data)
}

// DeoptimizeData deoptimizes data using the global optimizer
func DeoptimizeData(key string, data []byte) ([]byte, error) {
	return globalOptimizer.DeoptimizeData(key, data)
}

// GetBuffer gets a buffer from the global optimizer
func GetBuffer(size int) []byte {
	return globalOptimizer.GetBuffer(size)
}

// PutBuffer returns a buffer to the global optimizer
func PutBuffer(buf []byte) {
	globalOptimizer.PutBuffer(buf)
}

// OptimizeMemoryUsage optimizes memory usage globally
func OptimizeMemoryUsage() OptimizationResult {
	return globalOptimizer.OptimizeMemoryUsage()
}

// GetGlobalOptimizer returns the global optimizer instance
func GetGlobalOptimizer() *ResourceOptimizer {
	return globalOptimizer
}

// WithOptimization wraps a function with automatic resource optimization
func WithOptimization(ctx context.Context, fn func() error) error {
	// Get initial memory stats
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	
	// Execute function
	err := fn()
	
	// Optimize memory usage after execution
	globalOptimizer.OptimizeMemoryUsage()
	
	return err
}