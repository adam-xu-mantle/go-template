package graphql

import (
	"github.com/adam-xu-mantle/go-template/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the resolver for the GraphQL schema.
type Resolver struct {
	greeterService *service.GreeterService
}

// NewResolver creates a new GraphQL resolver
func NewResolver(greeterService *service.GreeterService) *Resolver {
	return &Resolver{
		greeterService: greeterService,
	}
}
