package main

import (
	"fmt"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
)

type cacheImpl struct {
	downstream CalendarInterface
	cache      *cache.Cache[string, any]
	expiration time.Duration
}

func NewCache(downstream CalendarInterface, expiration time.Duration) CalendarInterface {
	c := cache.New[string, any]()
	return &cacheImpl{downstream: downstream, cache: c, expiration: expiration}
}
func (c *cacheImpl) Years() []int {
	path := "/"
	if years, ok := c.cache.Get(path); ok {
		return years.([]int)
	}
	years := c.downstream.Years()
	c.cache.Set(path, years, cache.WithExpiration(c.expiration))
	return years
}

func (c *cacheImpl) Months(year int) []time.Month {
	path := fmt.Sprintf("/%d", year)
	if months, ok := c.cache.Get(path); ok {
		return months.([]time.Month)
	}
	months := c.downstream.Months(year)
	c.cache.Set(path, months, cache.WithExpiration(c.expiration))
	return months
}

func (c *cacheImpl) Days(year int, month time.Month) []int {
	path := fmt.Sprintf("/%d/%d", year, month)
	if days, ok := c.cache.Get(path); ok {
		return days.([]int)
	}
	days := c.downstream.Days(year, month)
	c.cache.Set(path, days, cache.WithExpiration(c.expiration))
	return days
}

func (c *cacheImpl) Entries(t time.Time) []CalendarEntry {
	path := fmt.Sprintf("/%d/%d/%d", t.Year(), t.Month(), t.Day())
	if entries, ok := c.cache.Get(path); ok {
		return entries.([]CalendarEntry)
	}
	entries := c.downstream.Entries(t)
	c.cache.Set(path, entries, cache.WithExpiration(c.expiration))
	return entries
}
