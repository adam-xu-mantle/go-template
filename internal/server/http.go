package server

import (
	"context"
	"net"
	"net/http"
	"time"

	//"go-template/api/helloworld/graphql/generated"
	v1 "github.com/adam-xu-mantle/go-template/api/helloworld/v1"

	"github.com/adam-xu-mantle/go-template/internal/conf"
	"github.com/adam-xu-mantle/go-template/internal/metrics"
	"github.com/adam-xu-mantle/go-template/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

// HTTPServer wraps gin.Engine to implement kratos transport interface
type HTTPServer struct {
	*gin.Engine
	server  *http.Server
	logger  *log.Helper
	network string
	address string
	timeout time.Duration
}

// customMiddleware is a middleware that logs the request and response
func customMiddleware(logger *log.Helper, metricer metrics.Metricer) gin.HandlerFunc {
	return func(c *gin.Context) {
		done := metricer.RecordMetricRequests(c.Request.URL.Path, c.Request.Method)
		defer done()

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build log message
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Infow("clientIP", clientIP, "method", method, "path", path, "statusCode", statusCode, "latency", latency)

	}
}

// NewHTTPServer creates a new Gin HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, logger log.Logger) *HTTPServer {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	metricer := metrics.NewMetricer("", "")
	logHelper := log.NewHelper(logger)
	r.Use(customMiddleware(logHelper, metricer), gin.Recovery())

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
		logger:  logHelper,
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
	// resolver := graphql.NewResolver(greeter)
	// // NewDefaultServer is a demonstration only. Not for prod.
	// srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
	// 	Resolvers: resolver,
	// }))

	// GraphQL endpoint
	// s.POST("/graphql", func(c *gin.Context) {
	// 	srv.ServeHTTP(c.Writer, c.Request)
	// })

	// Health check endpoint
	// s.GET("/health", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"status": "ok",
	// 	})
	// })
}

// Start implements the transport.Server interface
func (s *HTTPServer) Start(ctx context.Context) error {
	s.logger.Infof("[HTTP] server listening on: %s", s.address)

	listener, err := net.Listen(s.network, s.address)
	if err != nil {
		return err
	}

	go func() {
		_ = s.server.Serve(listener)
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
