package camillo

import (
	"net/http"
	"sync"

	"golang.org/x/net/context"
)

var sharedContextStore contextStore

type contextStore struct {
	mtx      sync.RWMutex
	contexts map[*http.Request]context.Context
}

func (s *contextStore) Get(req *http.Request) context.Context {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.contexts == nil {
		return nil
	}

	return s.contexts[req]
}

func (s *contextStore) Add(req *http.Request, ctx context.Context) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.contexts == nil {
		s.contexts = make(map[*http.Request]context.Context)
	}

	s.contexts[req] = ctx
}

func (s *contextStore) Remove(req *http.Request) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.contexts == nil {
		return
	}

	delete(s.contexts, req)
}
