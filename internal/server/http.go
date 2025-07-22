package server

import (
	"context"
	"net"
	"net/http"
	"time"

	//"moho-router/api/helloworld/graphql/generated"
	v1 "moho-router/api/helloworld/v1"
	"moho-router/internal/conf"
	"moho-router/internal/server/graphql"
	"moho-router/internal/server/graphql/generated"
	"moho-router/internal/service"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

// HTTPServer wraps gin.Engine to implement kratos transport interface
type HTTPServer struct {
	*gin.Engine
	server  *http.Server
	network string
	address string
	timeout time.Duration
	logger  *log.Helper
}

// NewHTTPServer creates a new Gin HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, logger log.Logger) *HTTPServer {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Set default values
	network := "tcp"
	address := ":8000"
	timeout := 30 * time.Second

	if c.Http != nil {
		if c.Http.Network != "" {
			network = c.Http.Network
		}
		if c.Http.Addr != "" {
			address = c.Http.Addr
		}
		if c.Http.Timeout != nil {
			timeout = c.Http.Timeout.AsDuration()
		}
	}

	srv := &HTTPServer{
		Engine:  r,
		network: network,
		address: address,
		timeout: timeout,
		logger:  log.NewHelper(logger),
	}

	srv.server = &http.Server{
		Addr:         address,
		Handler:      r,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	// Register routes
	srv.registerRoutes(greeter)

	return srv
}

// registerRoutes sets up the API routes
func (s *HTTPServer) registerRoutes(greeter *service.GreeterService) {
	// Register the greeter route: GET /helloworld/{name}
	s.GET("/helloworld/:name", func(c *gin.Context) {
		name := c.Param("name")

		req := &v1.HelloRequest{
			Name: name,
		}

		resp, err := greeter.SayHello(c.Request.Context(), req)
		if err != nil {
			s.logger.Errorf("Failed to process SayHello request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	// GraphQL setup
	resolver := graphql.NewResolver(greeter)
	// NewDefaultServer is a demonstration only. Not for prod.
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
	}))

	// GraphQL endpoint
	s.POST("/graphql", func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})

	// Health check endpoint
	s.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
}

// Start implements the transport.Server interface
func (s *HTTPServer) Start(ctx context.Context) error {
	s.logger.Infof("[HTTP] server listening on: %s", s.address)

	listener, err := net.Listen(s.network, s.address)
	if err != nil {
		return err
	}

	go func() {
		s.server.Serve(listener)
	}()

	return nil
}

// Stop implements the transport.Server interface
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("[HTTP] server stopping")
	return s.server.Shutdown(ctx)
}

// Ensure GinServer implements transport.Server interface
var _ transport.Server = (*HTTPServer)(nil)
