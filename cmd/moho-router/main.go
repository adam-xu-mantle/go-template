package main

import (
	"fmt"
	"os"

	"moho-router/internal/conf"
	"moho-router/internal/server"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/spf13/cobra"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "moho-router"
	// Version is the version of the compiled software.
	Version = "dev"
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func newApp(logger log.Logger, gs *server.GRPCServer, hs *server.HTTPServer) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the moho-router server",
	Long:  `Start the HTTP and gRPC servers with the specified configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version information of moho-router.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s\n", Name, Version)
	},
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Validate configuration file",
	Long:  `Validate the configuration file and print parsed settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		validateConfig()
	},
}

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run database migrations to set up or update the database schema.`,
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations()
	},
}

func init() {
	// Add persistent flags to root command
	rootCmd.PersistentFlags().StringVarP(&flagconf, "conf", "c", "./configs", "config path, eg: -conf config.yaml")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(migrateCmd)
}

func runServer() {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func validateConfig() {
	fmt.Printf("Validating configuration from: %s\n", flagconf)

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration is valid!")

	// Print configuration summary
	if bc.Server != nil {
		fmt.Printf("Server configuration:\n")
		if bc.Server.Http != nil {
			fmt.Printf("  HTTP: %s\n", bc.Server.Http.Addr)
		}
		if bc.Server.Grpc != nil {
			fmt.Printf("  gRPC: %s\n", bc.Server.Grpc.Addr)
		}
	}

	if bc.Data != nil {
		fmt.Printf("Data configuration:\n")
		if bc.Data.Database != nil {
			fmt.Printf("  Database driver: %s\n", bc.Data.Database.Driver)
		}
		if bc.Data.Redis != nil {
			fmt.Printf("  Redis: %s\n", bc.Data.Redis.Addr)
		}
	}
}

func runMigrations() {
	fmt.Println("Running database migrations...")

	// Load configuration to get database connection
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %v\n", err)
		os.Exit(1)
	}

	if bc.Data == nil || bc.Data.Database == nil {
		fmt.Fprintf(os.Stderr, "No database configuration found\n")
		os.Exit(1)
	}

	fmt.Printf("Database driver: %s\n", bc.Data.Database.Driver)
	fmt.Printf("Migration files location: ./migrations/\n")

	// TODO: Implement actual migration logic here
	// For now, just print information about what would be done
	fmt.Println("Found migration files:")
	fmt.Println("  - migrations/README.md")

	fmt.Println("Migration implementation is pending.")
	fmt.Println("You can implement actual database migration logic in the runMigrations() function.")
	fmt.Println("Migration command is ready to be extended!")
}

func main() {
	fmt.Println("Starting moho-router...") // 添加这行用于调试
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Program completed successfully") // 添加这行用于调试
}
