package generator

import (
	"context"
)

type Generator interface {
	Generate(context.Context) ([]byte, error)
}
