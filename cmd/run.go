package cmd

import (
	"log"
	"net"
	"sync"

	"github.com/cvetkovski98/zvax-auth/internal/config"
	"github.com/cvetkovski98/zvax-auth/internal/delivery"
	"github.com/cvetkovski98/zvax-auth/internal/model/migrations"
	"github.com/cvetkovski98/zvax-auth/internal/repository"
	"github.com/cvetkovski98/zvax-auth/internal/service"
	"github.com/cvetkovski98/zvax-common/gen/pbauth"
	"github.com/cvetkovski98/zvax-common/pkg/healthz"
	"github.com/cvetkovski98/zvax-common/pkg/postgresql"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	runCommand = &cobra.Command{
		Use:   "run",
		Short: "Run auth microservice",
		Long:  `Run auth microservice`,
		Run:   run,
	}
	network string
	address string
)

func init() {
	runCommand.Flags().StringVarP(&network, "network", "n", "tcp", "network to listen on")
	runCommand.Flags().StringVarP(&address, "address", "a", ":50052", "address to listen on")
}

func run(cmd *cobra.Command, args []string) {
	cfg := config.GetConfig()
	db, err := postgresql.NewPgDb(&cfg.PostgreSQL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := repository.RegisterModels(cmd.Context(), db); err != nil {
		log.Fatalf("failed to register models: %v", err)
	}
	if err := postgresql.Migrate(cmd.Context(), db, migrations.Migrations); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	authRepository := repository.NewPgAuthRepository(db)
	authService := service.NewAuthServiceImpl(authRepository)
	authGrpc := delivery.NewAuthServer(authService)
	server := grpc.NewServer()
	pbauth.RegisterAuthServer(server, authGrpc)

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		healtzSrv := healthz.CreateServer(80)
		if err := healtzSrv.ListenAndServe(); err != nil {
			log.Printf("error running healthz server: %v", err)
		}
		log.Println("Running healthz...")
		wg.Done()
	}()

	go func() {
		lis, err := net.Listen(network, address)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		log.Printf("Listening on %s://%s...", network, address)
		wg.Done()
	}()
	wg.Wait()
}
