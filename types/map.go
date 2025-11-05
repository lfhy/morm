package types

import "sync"

// OrderedMap 并发安全的有序 map
type OrderedMap struct {
	mu       sync.RWMutex
	data     map[string]interface{}
	keys     []string       // 记录插入顺序
	keyIndex map[string]int // 记录每个 key 在 keys 中的位置，便于快速删除
}

// NewOrderedMap 创建一个新的 OrderedMap
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		data:     make(map[string]interface{}),
		keys:     make([]string, 0),
		keyIndex: make(map[string]int),
	}
}

// Store 存储键值对
func (om *OrderedMap) Store(key string, value interface{}) {
	om.mu.Lock()
	defer om.mu.Unlock()
	if _, exists := om.data[key]; !exists {
		// 新增 key
		om.keys = append(om.keys, key)
		om.keyIndex[key] = len(om.keys) - 1
	}
	om.data[key] = value
}

// Load 获取键对应的值
func (om *OrderedMap) Load(key string) (value interface{}, ok bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()
	value, ok = om.data[key]
	return
}

// Delete 删除指定键
func (om *OrderedMap) Delete(key string) {
	om.mu.Lock()
	defer om.mu.Unlock()
	if idx, exists := om.keyIndex[key]; exists {
		delete(om.data, key)
		delete(om.keyIndex, key)
		// 将最后一个元素移到被删元素位置以避免拷贝整个数组
		lastIdx := len(om.keys) - 1
		if idx != lastIdx {
			om.keys[idx] = om.keys[lastIdx]
			om.keyIndex[om.keys[idx]] = idx
		}
		om.keys = om.keys[:lastIdx]
	}
}

// Range 按照插入顺序遍历所有键值对
func (om *OrderedMap) Range(fn func(key string, value interface{}) bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()
	for _, k := range om.keys {
		v := om.data[k]
		if !fn(k, v) {
			break
		}
	}
}
