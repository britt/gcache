package gcache

import "container/heap"

// TODO: See if there is a way to get rid of the flag arguments

// ScoreCache Discards the lowest scored items first.
// It uses an aggregate metric like total bytes to
// decide when evictions are necessary
type ScoreCache struct {
	baseCache
	items         map[interface{}]*scoredItem
	evictList     *priorityHeap
	computeScore  ScoringFunc
	computeWeight WeightingFunc
	totalWeight   int
}

// ScoringFunc computes the eviction priority for the queue
type ScoringFunc func(value interface{}) int

// WeightingFunc computes the weight of the item to determine
// when evictions are necessary.
type WeightingFunc func(value interface{}) int

func newScoreCache(cb *CacheBuilder) *ScoreCache {
	c := &ScoreCache{}
	buildCache(&c.baseCache, cb)
	c.computeScore = cb.scoringFunc
	c.computeWeight = cb.weightingFunc

	c.reset()
	c.loadGroup.cache = c
	return c
}

func (sc *ScoreCache) reset() {
	newHeap := priorityHeap([]*scoredItem{})
	sc.evictList = &newHeap
	heap.Init(sc.evictList)
	sc.items = make(map[interface{}]*scoredItem)
}

// Get returns an item from the cache if it is present. If it is not present
// it attempts to load it using the LoaderFunc.
func (sc *ScoreCache) Get(key interface{}) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	item, err := sc.getItem(key, true)
	if err != nil {
		return sc.getWithLoader(key, true)
	}
	return item.value, nil
}

// GetIFPresent returns an item from the cache if it is present in cache and a KeyNotFoundError if it is not.
// It does not attempt to load the item
func (sc *ScoreCache) GetIFPresent(key interface{}) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	item, err := sc.getItem(key, true)

	if item != nil {
		return item.value, nil
	}
	return nil, err
}

// GetALL returns all if the cached values
func (sc *ScoreCache) GetALL() map[interface{}]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	m := make(map[interface{}]interface{})
	for k, v := range sc.items {
		m[k] = v.value
	}

	return m
}

// Set adds a key, value pair to the cache
func (sc *ScoreCache) Set(key, value interface{}) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.set(key, value)
}

// set an item without locking and return the item
func (sc *ScoreCache) set(key, value interface{}) *scoredItem {
	// Check for existing item
	existing, err := sc.getItem(key, false)
	if err == nil {
		sc.totalWeight -= existing.weight
		existing.value = value
		existing.score = sc.computeScore(value)
		existing.weight = sc.computeWeight(value)
		sc.totalWeight += existing.weight
		idx, _ := sc.getIndex(key)
		heap.Fix(sc.evictList, idx)
		return existing
	}

	// Otherwise add to cache
	item := sc.newScoredItem(key, value)
	// Verify item will not exceed total weight
	if sc.totalWeight+item.weight > sc.size {
		sc.evictUntil(item.weight)
	}
	heap.Push(sc.evictList, item)
	sc.items[key] = item
	sc.totalWeight += item.weight

	sc.addedCallback(key, value)

	return item
}

// Remove deletes an item
func (sc *ScoreCache) Remove(key interface{}) bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if item, ok := sc.items[key]; ok {
		delete(sc.items, key)
		index := -1

		for i, it := range []*scoredItem(*sc.evictList) {
			if it.key == key {
				index = i
				break
			}
		}

		if index > 0 {
			heap.Remove(sc.evictList, index)
			sc.totalWeight -= item.weight
			sc.evictedCallback(item.key, item.value)
			return true
		}
	}
	return false
}

// Purge removes all items from the cache without calling eviction handlers
func (sc *ScoreCache) Purge() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.reset()
}

// Keys returns all of the keys in the cache
func (sc *ScoreCache) Keys() []interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	keys := make([]interface{}, len(sc.items))
	i := 0
	for k := range sc.items {
		keys[i] = k
		i++
	}

	return keys
}

// Len returns the number of items in the cache
func (sc *ScoreCache) Len() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return len(sc.items)
}

// loads an item using the loaderFunc
func (sc *ScoreCache) getWithLoader(key interface{}, isWait bool) (interface{}, error) {
	if sc.loaderFunc == nil {
		return nil, KeyNotFoundError
	}

	item, _, err := sc.load(key, func(v interface{}, e error) (interface{}, error) {
		if e == nil {
			return sc.set(key, v), nil
		}
		return nil, e
	}, isWait)
	if err != nil {
		return nil, err
	}
	return item.(*scoredItem).value, nil
}

// gets an item from the cache with an options load flag
func (sc *ScoreCache) get(key interface{}, onLoad bool) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.getItem(key, onLoad)
}

// gets an item from the cache (not threadsafe!)
func (sc *ScoreCache) getItem(key interface{}, count bool) (*scoredItem, error) {
	item, ok := sc.items[key]
	if !ok {
		if count {
			sc.IncrMissCount()
		}
		return item, KeyNotFoundError
	}
	if count {
		sc.IncrHitCount()
	}
	return item, nil
}

func (sc *ScoreCache) evictUntil(w int) {
	targetWeight := sc.totalWeight - w
	var item *scoredItem
	for sc.totalWeight > targetWeight {
		item = heap.Pop(sc.evictList).(*scoredItem)
		delete(sc.items, item.key)
		sc.evictedCallback(item.key, item.value)
		sc.totalWeight -= item.weight
	}
}

func (sc *ScoreCache) addedCallback(key, value interface{}) {
	if sc.addedFunc != nil {
		(*sc.addedFunc)(key, value)
	}
}

func (sc *ScoreCache) evictedCallback(key, value interface{}) {
	if sc.evictedFunc != nil {
		(*sc.evictedFunc)(key, value)
	}
}

func (sc *ScoreCache) getIndex(key interface{}) (int, error) {
	for i, item := range []*scoredItem(*sc.evictList) {
		if item.key == key {
			return i, nil
		}
	}
	return -1, KeyNotFoundError
}

type scoredItem struct {
	key    interface{}
	value  interface{}
	score  int
	weight int
}

func (sc *ScoreCache) newScoredItem(key, value interface{}) *scoredItem {
	score := sc.computeScore(value)
	weight := sc.computeWeight(value)

	return &scoredItem{key: key, value: value, score: score, weight: weight}
}

type priorityHeap []*scoredItem

func (h *priorityHeap) Push(x interface{}) {
	item := x.(*scoredItem)
	*h = append(*h, item)
}

func (h *priorityHeap) Pop() interface{} {
	old := *h
	item := old[len(old)-1]
	*h = old[0 : len(old)-1]
	return item
}

func (h priorityHeap) Len() int {
	return len(h)
}

func (h priorityHeap) Less(i, j int) bool {
	return h[i].score < h[j].score
}

func (h priorityHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
