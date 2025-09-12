package generator

import (
	"context"
)

type Generator interface {
	Generate(context.Context) (*Result, error)
}

type Result struct {
	Output any
	// TODO: Indicate if the output is expected to be a stream
	Format string
}
