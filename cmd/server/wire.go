//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/adam-xu-mantle/go-template/internal/biz"
	"github.com/adam-xu-mantle/go-template/internal/conf"
	"github.com/adam-xu-mantle/go-template/internal/data"
	"github.com/adam-xu-mantle/go-template/internal/server"
	"github.com/adam-xu-mantle/go-template/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
