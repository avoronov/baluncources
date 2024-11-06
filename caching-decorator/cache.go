package main

import (
	"container/list"
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/singleflight"
)

/*
	TODO:
		- how to implement memory boundaries on cache size (i.e. 1 Gb)?
		- how to implement periodically invalidation of cache?

	Q:
		- why `[]*string` in `MGet(context.Context, []string) ([]*string, error)`?
*/

type (
	Datasource interface {
		Get(context.Context, string) (string, error)
		MGet(context.Context, []string) ([]*string, error)
		Keys(context.Context) ([]string, error)
	}

	node struct {
		data string
		ptr  *list.Element
	}

	Cache struct {
		src   Datasource
		cap   uint
		group *singleflight.Group
		m     *sync.RWMutex
		lru   *list.List
		data  map[string]*node
	}
)

func NewCache(src Datasource, cap uint) *Cache {
	return &Cache{
		src:   src,
		cap:   cap,
		group: new(singleflight.Group),
		m:     new(sync.RWMutex),
		lru:   list.New(),
		data:  make(map[string]*node, cap),
	}
}

func doWithLock(l sync.Locker, f func()) {
	l.Lock()
	defer l.Unlock()

	f()
}

func (c *Cache) dropLRU() {
	back := c.lru.Back()
	c.lru.Remove(back)
	delete(c.data, back.Value.(string))
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	res, err, _ := c.group.Do(key, func() (any, error) {
		var (
			val   string
			found bool
		)

		doWithLock(
			c.m.RLocker(),
			func() {
				res, ok := c.data[key]
				if !ok {
					return
				}

				c.lru.MoveToFront(res.ptr)
				val = res.data
				found = true
			},
		)

		if found {
			return val, nil
		}

		val, err := c.src.Get(ctx, key)
		if err != nil {
			return "", fmt.Errorf("fetch value from datasource: %w", err)
		}

		doWithLock(
			c.m,
			func() {
				if int(c.cap) == len(c.data) {
					c.dropLRU()
				}

				c.data[key] = &node{data: val, ptr: c.lru.PushFront(key)}
			},
		)

		return val, err
	})

	if err != nil {
		return "", nil
	}

	return res.(string), nil
}

func (c *Cache) MGet(ctx context.Context, keys []string) ([]*string, error) {
	var (
		vals      map[string]string
		notCached []string
	)

	doWithLock(
		c.m.RLocker(),
		func() {
			for _, key := range keys {
				res, ok := c.data[key]
				if !ok {
					notCached = append(notCached, key)

					continue
				}

				c.lru.MoveToFront(res.ptr)
				vals[key] = res.data
			}
		},
	)

	if len(notCached) > 0 {
		raw, err := c.src.MGet(ctx, notCached)
		if err != nil {
			return nil, fmt.Errorf("fetch values from datasource: %w", err)
		}

		doWithLock(
			c.m,
			func() {
				for i, key := range notCached {
					if int(c.cap) == len(c.data) {
						c.dropLRU()
					}

					c.data[key] = &node{data: *raw[i], ptr: c.lru.PushFront(key)}

					vals[key] = *raw[i]
				}
			},
		)
	}

	res := make([]*string, 0, len(keys))
	for _, key := range keys {
		r := vals[key]
		res = append(res, &r)
	}

	return res, nil
}

func (c *Cache) Keys(ctx context.Context) ([]string, error) {
	return c.src.Keys(ctx) // It seems that there is no point in inspecting the local cache for all the keys
}
