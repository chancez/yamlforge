package generator

import (
	"context"
)

type Generator interface {
	Generate(context.Context) (any, error)
}
