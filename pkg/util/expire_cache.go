/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"sync"
	"time"

	"github.com/golang/groupcache/lru"
)

type ExpireCache struct {
	cache *lru.Cache
	lock  sync.RWMutex
}

func NewExpireCache(maxSize int) *ExpireCache {
	return &ExpireCache{cache: lru.New(maxSize)}
}

type Expirable interface {
	Expiration() time.Time
}

func (c *ExpireCache) Add(key lru.Key, e Expirable) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache.Add(key, e)
	// Remove entry from cache upon expiry.
	time.AfterFunc(e.Expiration().Sub(time.Now()), func() {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.cache.Remove(key)
	})
}

func (c *ExpireCache) Get(key lru.Key) (Expirable, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	e, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	entry := e.(Expirable)
	if time.Now().After(entry.Expiration()) {
		return nil, false
	}
	return entry, true
}
