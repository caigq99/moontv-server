package cache

import (
	"strconv"
	"sync"
	"time"

	"github.com/moontv/server/pkg/upstream"
)

type entry struct {
	results   []upstream.SearchResult
	pageCount int
	expiresAt time.Time
}

type SearchCache struct {
	data sync.Map
	ttl  time.Duration
}

func NewSearchCache(ttl time.Duration) *SearchCache {
	c := &SearchCache{ttl: ttl}
	go c.cleanup()
	return c
}

func (c *SearchCache) makeKey(sourceKey, query string, page int) string {
	return sourceKey + "::" + query + "::" + strconv.Itoa(page)
}

func (c *SearchCache) Get(sourceKey, query string, page int) ([]upstream.SearchResult, int, bool) {
	key := c.makeKey(sourceKey, query, page)
	val, ok := c.data.Load(key)
	if !ok {
		return nil, 0, false
	}
	e := val.(*entry)
	if time.Now().After(e.expiresAt) {
		c.data.Delete(key)
		return nil, 0, false
	}
	return e.results, e.pageCount, true
}

func (c *SearchCache) Set(sourceKey, query string, page int, results []upstream.SearchResult, pageCount int) {
	key := c.makeKey(sourceKey, query, page)
	c.data.Store(key, &entry{
		results:   results,
		pageCount: pageCount,
		expiresAt: time.Now().Add(c.ttl),
	})
}

func (c *SearchCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		now := time.Now()
		c.data.Range(func(key, val any) bool {
			if e := val.(*entry); now.After(e.expiresAt) {
				c.data.Delete(key)
			}
			return true
		})
	}
}
