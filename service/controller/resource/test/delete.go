package test

import (
	"context"
	"fmt"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	fmt.Printf("Object deleted")
	return nil
}
