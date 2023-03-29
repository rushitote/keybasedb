package main

// Storage Engine Interface
type Engine struct {
	m map[string]string
}

func (e *Engine) Read(key string) string {
	return e.m[key]
}

func (e *Engine) Write(key, value string) {
	e.m[key] = value
}

func (e *Engine) Delete(key string) {
	e.m[key] = DeletedHash
}

const (
	DeletedHash = "hefiwhe783d7qdiq83"
)
