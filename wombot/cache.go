package main

import (
	"strconv" // need FormatBool and Itoa for `Print()`
	"strings" // need builder for `Print()`
)

type IIUCache struct {
	Cache *IIUCacheElement
	Cap   uint
}

type IIUCacheElement struct {
	ID         int64
	Value      bool
	Prev, Next *IIUCacheElement
}

func NewIIUCache(capacity uint) *IIUCache {
	return &IIUCache{
		Cap:   capacity,
		Cache: nil,
	}
}

func (c *IIUCache) Len() uint {
	var i uint
	for el := c.Cache; el != nil; el = el.Next {
		i++
	}
	return i
}

func (c *IIUCache) Pop() {
	if c.Cache == nil {
		return
	}
	var el = c.Cache
	for ; el.Next != nil; el = el.Next {

	}
	if el.Prev != nil {
		el.Prev.Next = nil
	} else {
		c.Cache = nil
	}
}

func (c *IIUCache) Get(id int64) (bool, *IIUCacheElement) {
	var need *IIUCacheElement
	for el := c.Cache; el != nil; el = el.Next {
		if el.ID == id {
			need = el
			break
		}
	}
	if need == nil {
		return false, nil
	}
	if need.Prev != nil {
		need.Prev.Next = need.Next
		need.Next = c.Cache
	}
	c.Cache = need
	return true, need
}

func (c *IIUCache) Put(id int64, val bool) {
	if c.Len() >= c.Cap {
		c.Pop()
	}

	if is, el := c.Get(id); is && el != nil {
		el.Value = val
		return
	}

	c.Cache = &IIUCacheElement{
		ID:    id,
		Value: val,
		Prev:  nil,
		Next:  c.Cache,
	}
}

func (c IIUCache) Print() {
	var result strings.Builder
	for el := c.Cache; el != nil; el = el.Next {
		result.WriteString(" -> ")
		result.WriteString(strconv.Itoa(int(el.ID)))
		result.WriteRune(':')
		result.WriteString(strconv.FormatBool(el.Value))
	}
	result.WriteString(" -> nil")
	debl.Println(result.String())
}
