package camillo

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
)

// Handler handler is an interface that objects can implement to be registered to serve as middleware
// in the Camillo middleware stack.
// ServeHTTP should yield to the next middleware in the chain by invoking the next http.HandlerFunc
// passed in.
//
// If the Handler writes to the ResponseWriter, the next http.HandlerFunc should not be invoked.
type Handler interface {
	ServeHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request, next NextFunc)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as Camillo handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler object that calls f.
type HandlerFunc func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next NextFunc)

// NextFunc passes the request to the next middleware layer
type NextFunc func(ctx context.Context, rw http.ResponseWriter, r *http.Request)

func (h HandlerFunc) ServeHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request, next NextFunc) {
	h(ctx, rw, r, next)
}

type middleware struct {
	handler Handler
	next    *middleware
}

func (m middleware) ServeHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	m.handler.ServeHTTP(ctx, rw, r, m.next.ServeHTTP)
}

// Wrap converts a http.Handler into a camillo.Handler so it can be used as a Camillo
// middleware. The next http.HandlerFunc is automatically called after the Handler
// is executed.
func Wrap(handler http.Handler) Handler {
	return HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next NextFunc) {
		sharedContextStore.Add(r, ctx)

		handler.ServeHTTP(rw, r)

		ctx = sharedContextStore.Get(r)
		next(ctx, rw, r)
	})
}

// Camillo is a stack of Middleware Handlers that can be invoked as an http.Handler.
// Camillo middleware is evaluated in the order that they are added to the stack using
// the Use and UseHandler methods.
type Camillo struct {
	ctx        context.Context
	middleware middleware
	handlers   []Handler
}

// New returns a new Camillo instance with no middleware preconfigured.
func New(handlers ...Handler) *Camillo {
	return NewWithContext(context.Background(), handlers...)
}

// NewWithContext returns a new Camillo instance with no middleware preconfigured.
func NewWithContext(ctx context.Context, handlers ...Handler) *Camillo {
	return &Camillo{
		ctx:        ctx,
		handlers:   handlers,
		middleware: build(handlers),
	}
}

// Classic returns a new Camillo instance with the default middleware already
// in the stack.
//
// Recovery - Panic Recovery Middleware
// Logger - Request/Response Logging
// Static - Static File Serving
func Classic() *Camillo {
	return New(NewRecovery(), NewLogger(), NewStatic(http.Dir("public")))
}

func (n *Camillo) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var ctx context.Context

	ctx = sharedContextStore.Get(r)
	if ctx != nil {
		n.middleware.ServeHTTP(ctx, NewResponseWriter(rw), r)
		return
	}

	ctx = n.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sharedContextStore.Add(r, ctx)
	defer sharedContextStore.Remove(r)

	n.middleware.ServeHTTP(ctx, NewResponseWriter(rw), r)
}

// Use adds a Handler onto the middleware stack. Handlers are invoked in the order they are added to a Camillo.
func (n *Camillo) Use(handler Handler) {
	n.handlers = append(n.handlers, handler)
	n.middleware = build(n.handlers)
}

// UseFunc adds a Camillo-style handler function onto the middleware stack.
func (n *Camillo) UseFunc(handlerFunc func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next NextFunc)) {
	n.Use(HandlerFunc(handlerFunc))
}

// UseHandler adds a http.Handler onto the middleware stack. Handlers are invoked in the order they are added to a Camillo.
func (n *Camillo) UseHandler(handler http.Handler) {
	n.Use(Wrap(handler))
}

// UseHandlerFunc adds a http.HandlerFunc-style handler function onto the middleware stack.
func (n *Camillo) UseHandlerFunc(handlerFunc func(rw http.ResponseWriter, r *http.Request)) {
	n.UseHandler(http.HandlerFunc(handlerFunc))
}

// Run is a convenience function that runs the camillo stack as an HTTP
// server. The addr string takes the same format as http.ListenAndServe.
func (n *Camillo) Run(addr string) {
	l := log.New(os.Stdout, "[camillo] ", 0)
	l.Printf("listening on %s", addr)
	l.Fatal(http.ListenAndServe(addr, n))
}

// Handlers returns a list of all the handlers in the current Camillo middleware chain.
func (n *Camillo) Handlers() []Handler {
	return n.handlers
}

func build(handlers []Handler) middleware {
	var next middleware

	if len(handlers) == 0 {
		return voidMiddleware()
	} else if len(handlers) > 1 {
		next = build(handlers[1:])
	} else {
		next = voidMiddleware()
	}

	return middleware{handlers[0], &next}
}

func voidMiddleware() middleware {
	return middleware{
		HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next NextFunc) {}),
		&middleware{},
	}
}
