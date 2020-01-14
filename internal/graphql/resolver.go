package graphql

import (
	"context"
)

// Resolver : Root resolver for service.
type Resolver struct{}

// Mutation : Mutation resolver for service.
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Query : Query resolver for service.
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) NewPhysicalVolume(ctx context.Context, input NewTodo) (*Todo, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) ServiceTypes(ctx context.Context) ([]*ServiceType, error) {
	panic("not implemented")

	// TODO: DB lookup
}
