package camillo

import (
	"net/http"
	"sync"

	"golang.org/x/net/context"
)

var sharedContextStore contextStore

type contextStore struct {
	mtx      sync.RWMutex
	contexts map[*http.Request][]context.Context
}

func (s *contextStore) Get(req *http.Request) context.Context {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.contexts == nil {
		return nil
	}

	stack := s.contexts[req]
	if len(stack) == 0 {
		return nil
	}

	return stack[len(stack)-1]
}

func (s *contextStore) Push(req *http.Request, ctx context.Context) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.contexts == nil {
		s.contexts = make(map[*http.Request][]context.Context)
	}

	s.contexts[req] = append(s.contexts[req], ctx)
}

func (s *contextStore) Pop(req *http.Request, ctx context.Context) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.contexts == nil {
		return
	}

	stack := s.contexts[req]
	if len(stack) == 0 {
		panic("unbalanced Push/Pop calls")
	}
	if stack[len(stack)-1] != ctx {
		panic("unbalanced Push/Pop calls")
	}

	stack = stack[:len(stack)-1]
	if len(stack) == 0 {
		delete(s.contexts, req)
	} else {
		s.contexts[req] = stack
	}
}
