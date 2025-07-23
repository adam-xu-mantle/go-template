package main

import (
	"fmt"
	"go-template/internal/conf"
	"go-template/internal/log"
	"go-template/internal/server"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/spf13/cobra"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "go-template"
	// Version is the version of the compiled software.
	Version = "dev"
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func newApp(logger klog.Logger, gs *server.GRPCServer, hs *server.HTTPServer) *kratos.App {
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
	Short: "Start the go-template server",
	Long:  `Start the HTTP and gRPC servers with the specified configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version information of go-template.`,
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
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
			// env.NewSource(""), Load env variables from the environment
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

	if bc.Metrics != nil && !bc.Metrics.Disable {
		go func() {
			addr := ":8080"
			if bc.Metrics.Addr != "" {
				addr = bc.Metrics.Addr
			}
			http.Handle("/metrics", promhttp.Handler())

			server := &http.Server{
				Addr:         addr,
				Handler:      nil, // uses DefaultServeMux
				ReadTimeout:  60 * time.Second,
				WriteTimeout: 60 * time.Second,
				IdleTimeout:  60 * time.Second,
			}

			if err := server.ListenAndServe(); err != nil {
				panic(err)
			}
		}()
	}

	logger := log.NewLogger(bc.Log)
	_ = logger.Log(klog.LevelInfo, "msg", "starting logger")
	logger = klog.With(logger,
		"level", klog.LevelInfo,
		"ts", klog.DefaultTimestamp,
		"caller", klog.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
	)
	klog.SetLogger(logger)

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
			// env.NewSource(""), Load env variables from the environment
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

	if bc.Log != nil {
		fmt.Printf("Log configuration:\n")
		fmt.Printf("  Level: %s\n", bc.Log.Level)
		fmt.Printf("  Format: %s\n", bc.Log.Format)
	}

	if bc.Metrics != nil {
		fmt.Printf("Metrics configuration:\n")
		fmt.Printf("  Addr: %s\n", bc.Metrics.Addr)
		fmt.Printf("  Disable: %t\n", bc.Metrics.Disable)
	}
}

func runMigrations() {
	fmt.Println("Running database migrations...")

	// Load configuration to get database connection
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
			// env.NewSource(""), Load env variables from the environment
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
	fmt.Println("Starting go-template...") // 添加这行用于调试
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Program completed successfully") // 添加这行用于调试
}
