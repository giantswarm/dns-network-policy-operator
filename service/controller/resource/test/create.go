package test

import (
	"context"
	"fmt"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	fmt.Printf("Object created")
	return nil
}
