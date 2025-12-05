package main

import (
	"fmt"
	"time"
)

//go:generate go run ../../irpc

// KVStore is the interface irpc will generate an RPC client & server for.
type KVStore interface {
	Put(key string, value []byte, ttl time.Duration) error
	Get(key string) ([]byte, error)
	Delete(key string) error

	ModifiedSince(since time.Time) ([]string, error)
}

// --- Simple in-memory implementation for example purposes ----

var _ KVStore = &kvMemory{}

type kvMemory struct {
	data map[string]entry
}

type entry struct {
	value     []byte
	expiresAt time.Time
	modified  time.Time
}

func newKVMemory() *kvMemory {
	return &kvMemory{data: make(map[string]entry)}
}

func (kv *kvMemory) Put(key string, value []byte, ttl time.Duration) error {
	kv.data[key] = entry{
		value:     value,
		modified:  time.Now(),
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (kv *kvMemory) Get(key string) ([]byte, error) {
	e, ok := kv.data[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, fmt.Errorf("not found")
	}
	return e.value, nil
}

func (kv *kvMemory) Delete(key string) error {
	delete(kv.data, key)
	return nil
}

func (kv *kvMemory) ModifiedSince(t time.Time) ([]string, error) {
	out := []string{}
	for k, e := range kv.data {
		if e.modified.After(t) {
			out = append(out, k)
		}
	}
	return out, nil
}
