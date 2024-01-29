package example

import (
	"context"
)

//go:generate wrappergen

type TestInterface interface {
	TestMethod(
		ctx context.Context,
		slice []int64,
		testMap map[string]context.Context,
		points ...string,
	) ([]string, error)
	TestMethod2(ctx context.Context) error
}
