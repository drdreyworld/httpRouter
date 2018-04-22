package httpRouter

import "context"

type ActionFunc func(ctx context.Context) (interface{}, error)

type HttpAction struct {
	action ActionFunc
	ctx    context.Context
}



