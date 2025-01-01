// Copyright 2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package gencache

import (
	"errors"
	"testing"

	"github.com/vedranvuk/strutils"
)

type CacheTestItem struct {
	ID   string
	Data []byte
}

func equalByteSlice(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func RandomKey() string {
	return strutils.RandomString(true, true, true, 16)
}

func TestGenCache(t *testing.T) {
	var data = []CacheTestItem{
		{RandomKey(), []byte{1, 2, 3, 4}},
		{RandomKey(), []byte{5, 6, 7, 8}},
		{RandomKey(), []byte{9, 10, 11, 12}},
		{RandomKey(), []byte{13, 14, 15, 16}},
	}
	var cache = NewGenCache[string, []byte](8, 8)
	cache.Put(data[0].ID, data[0].Data)
	cache.Put(data[1].ID, data[1].Data)
	cache.Put(data[2].ID, data[2].Data)
	cache.Put(data[3].ID, data[3].Data)

	var (
		buf []byte
		found bool
	)
	if buf, found = cache.Get(data[2].ID); !found {
		t.Fatal("not found")
	}
	if !equalByteSlice(buf, data[2].Data) {
		t.Fatal(errors.New("unexpected data"))
	}

	if buf, found = cache.Get(data[3].ID); !found {
		t.Fatal("not found")
	}
	if !equalByteSlice(buf, data[3].Data) {
		t.Fatal(errors.New("unexpected data"))
	}
}

func BenchmarkCachePut(b *testing.B) {
	var data = []CacheTestItem{
		{RandomKey(), []byte{1, 2, 3, 4}},
		{RandomKey(), []byte{5, 6, 7, 8}},
		{RandomKey(), []byte{9, 10, 11, 12}},
		{RandomKey(), []byte{13, 14, 15, 16}},
	}
	var cache = NewGenCache[string, []byte](8, 8)
	cache.Put(data[0].ID, data[0].Data)
	cache.Put(data[1].ID, data[1].Data)
	cache.Put(data[2].ID, data[2].Data)
	cache.Put(data[3].ID, data[3].Data)
	var idx int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx = i % 4
		cache.Put(data[idx].ID, data[idx].Data)
	}
	b.StopTimer()
}

func BenchmarkCacheGet(b *testing.B) {
	var data = []CacheTestItem{
		{RandomKey(), []byte{1, 2, 3, 4}},
		{RandomKey(), []byte{5, 6, 7, 8}},
		{RandomKey(), []byte{9, 10, 11, 12}},
		{RandomKey(), []byte{13, 14, 15, 16}},
	}
	var cache = NewGenCache[string, []byte](8, 8)
	cache.Put(data[0].ID, data[0].Data)
	cache.Put(data[1].ID, data[1].Data)
	cache.Put(data[2].ID, data[2].Data)
	cache.Put(data[3].ID, data[3].Data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(data[i%4].ID)
	}
	b.StopTimer()
}
