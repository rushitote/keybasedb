package main

type HashTable struct {
	m map[string]string
}

func NewHashTable() *HashTable {
	return &HashTable{make(map[string]string)}
}

func (h *HashTable) Get(key string) string {
	return h.m[key]
}

func (h *HashTable) Set(key, value string) {
	h.m[key] = value
}

func (h *HashTable) Delete(key string) {
	delete(h.m, key)
}
