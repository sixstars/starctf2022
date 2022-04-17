package bus

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/opentracing/opentracing-go"
)

// HandlerFunc defines a handler function interface.
type HandlerFunc interface{}

// Msg defines a message interface.
type Msg interface{}

// ErrHandlerNotFound defines an error if a handler is not found
var ErrHandlerNotFound = errors.New("handler not found")

// TransactionManager defines a transaction interface
type TransactionManager interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// Bus type defines the bus interface structure
type Bus interface {
	Dispatch(msg Msg) error
	DispatchCtx(ctx context.Context, msg Msg) error
	Publish(msg Msg) error

	// InTransaction starts a transaction and store it in the context.
	// The caller can then pass a function with multiple DispatchCtx calls that
	// all will be executed in the same transaction. InTransaction will rollback if the
	// callback returns an error.
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	AddHandler(handler HandlerFunc)
	AddHandlerCtx(handler HandlerFunc)
	AddEventListener(handler HandlerFunc)

	// SetTransactionManager allows the user to replace the internal
	// noop TransactionManager that is responsible for managing
	// transactions in `InTransaction`
	SetTransactionManager(tm TransactionManager)
}

// InProcBus defines the bus structure
type InProcBus struct {
	logger          log.Logger
	handlers        map[string]HandlerFunc
	handlersWithCtx map[string]HandlerFunc
	listeners       map[string][]HandlerFunc
	txMng           TransactionManager
}

func ProvideBus() *InProcBus {
	return globalBus
}

// InTransaction defines an in transaction function
func (b *InProcBus) InTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return b.txMng.InTransaction(ctx, fn)
}

// temp stuff, not sure how to handle bus instance, and init yet
var globalBus = New()

// New initialize the bus
func New() *InProcBus {
	return &InProcBus{
		logger:          log.New("bus"),
		handlers:        make(map[string]HandlerFunc),
		handlersWithCtx: make(map[string]HandlerFunc),
		listeners:       make(map[string][]HandlerFunc),
		txMng:           &noopTransactionManager{},
	}
}

// Want to get rid of global bus
func GetBus() Bus {
	return globalBus
}

// SetTransactionManager function assign a transaction manager to the bus.
func (b *InProcBus) SetTransactionManager(tm TransactionManager) {
	b.txMng = tm
}

// DispatchCtx function dispatch a message to the bus context.
func (b *InProcBus) DispatchCtx(ctx context.Context, msg Msg) error {
	var msgName = reflect.TypeOf(msg).Elem().Name()

	span, ctx := opentracing.StartSpanFromContext(ctx, "bus - "+msgName)
	defer span.Finish()

	span.SetTag("msg", msgName)

	withCtx := true
	var handler = b.handlersWithCtx[msgName]
	if handler == nil {
		withCtx = false
		handler = b.handlers[msgName]
		if handler == nil {
			return ErrHandlerNotFound
		}
	}

	var params = []reflect.Value{}
	if withCtx {
		params = append(params, reflect.ValueOf(ctx))
	} else if setting.Env == setting.Dev {
		b.logger.Warn("DispatchCtx called with message handler registered using AddHandler and should be changed to use AddHandlerCtx", "msgName", msgName)
	}
	params = append(params, reflect.ValueOf(msg))

	ret := reflect.ValueOf(handler).Call(params)
	err := ret[0].Interface()
	if err == nil {
		return nil
	}
	return err.(error)
}

// Dispatch function dispatch a message to the bus.
func (b *InProcBus) Dispatch(msg Msg) error {
	var msgName = reflect.TypeOf(msg).Elem().Name()

	withCtx := true
	handler := b.handlersWithCtx[msgName]
	if handler == nil {
		withCtx = false
		handler = b.handlers[msgName]
		if handler == nil {
			return ErrHandlerNotFound
		}
	}

	var params = []reflect.Value{}
	if withCtx {
		if setting.Env == setting.Dev {
			b.logger.Warn("Dispatch called with message handler registered using AddHandlerCtx and should be changed to use DispatchCtx", "msgName", msgName)
		}
		params = append(params, reflect.ValueOf(context.Background()))
	}
	params = append(params, reflect.ValueOf(msg))

	ret := reflect.ValueOf(handler).Call(params)
	err := ret[0].Interface()
	if err == nil {
		return nil
	}
	return err.(error)
}

// Publish function publish a message to the bus listener.
func (b *InProcBus) Publish(msg Msg) error {
	var msgName = reflect.TypeOf(msg).Elem().Name()
	var listeners = b.listeners[msgName]

	var params = make([]reflect.Value, 1)
	params[0] = reflect.ValueOf(msg)

	for _, listenerHandler := range listeners {
		ret := reflect.ValueOf(listenerHandler).Call(params)
		e := ret[0].Interface()
		if e != nil {
			err, ok := e.(error)
			if ok {
				return err
			}
			return fmt.Errorf("expected listener to return an error, got '%T'", e)
		}
	}

	return nil
}

func (b *InProcBus) AddHandler(handler HandlerFunc) {
	handlerType := reflect.TypeOf(handler)
	queryTypeName := handlerType.In(0).Elem().Name()
	b.handlers[queryTypeName] = handler
}

func (b *InProcBus) AddHandlerCtx(handler HandlerFunc) {
	handlerType := reflect.TypeOf(handler)
	queryTypeName := handlerType.In(1).Elem().Name()
	b.handlersWithCtx[queryTypeName] = handler
}

// GetHandlerCtx returns the handler function for the given struct name.
func (b *InProcBus) GetHandlerCtx(name string) HandlerFunc {
	return b.handlersWithCtx[name]
}

func (b *InProcBus) AddEventListener(handler HandlerFunc) {
	handlerType := reflect.TypeOf(handler)
	eventName := handlerType.In(0).Elem().Name()
	_, exists := b.listeners[eventName]
	if !exists {
		b.listeners[eventName] = make([]HandlerFunc, 0)
	}
	b.listeners[eventName] = append(b.listeners[eventName], handler)
}

// AddHandler attaches a handler function to the global bus.
// Package level function.
func AddHandler(implName string, handler HandlerFunc) {
	globalBus.AddHandler(handler)
}

// AddHandlerCtx attaches a handler function to the global bus context.
// Package level function.
func AddHandlerCtx(implName string, handler HandlerFunc) {
	globalBus.AddHandlerCtx(handler)
}

// AddEventListener attaches a handler function to the event listener.
// Package level function.
func AddEventListener(handler HandlerFunc) {
	globalBus.AddEventListener(handler)
}

func Dispatch(msg Msg) error {
	return globalBus.Dispatch(msg)
}

func DispatchCtx(ctx context.Context, msg Msg) error {
	return globalBus.DispatchCtx(ctx, msg)
}

func Publish(msg Msg) error {
	return globalBus.Publish(msg)
}

func GetHandlerCtx(name string) HandlerFunc {
	return globalBus.GetHandlerCtx(name)
}

func ClearBusHandlers() {
	globalBus = New()
}

type noopTransactionManager struct{}

func (*noopTransactionManager) InTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
