package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ozonmp/week-5-workshop/product-service/internal/pkg/db"
	"github.com/ozonmp/week-5-workshop/product-service/internal/pkg/logger"
	gelf "github.com/snovichkov/zap-gelf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"

	grpc_category_service "github.com/ozonmp/week-5-workshop/category-service/pkg/category-service"

	"github.com/ozonmp/week-5-workshop/product-service/internal/config"
	mwclient "github.com/ozonmp/week-5-workshop/product-service/internal/pkg/mw/client"
	"github.com/ozonmp/week-5-workshop/product-service/internal/server"
	product_service "github.com/ozonmp/week-5-workshop/product-service/internal/service/product"
)

func main() {
	ctx := context.Background()

	if err := config.ReadConfigYML("config.yml"); err != nil {
		logger.FatalKV(ctx, "Failed init configuration", "err", err)
	}
	cfg := config.GetConfigInstance()

	flag.Parse()

	syncLogger := initLogger(ctx, cfg)
	defer syncLogger()

	logger.InfoKV(ctx, fmt.Sprintf("Starting service: %s", cfg.Project.Name),
		"version", cfg.Project.Version,
		"commitHash", cfg.Project.CommitHash,
		"debug", cfg.Project.Debug,
		"environment", cfg.Project.Environment,
	)

	categoryServiceConn, err := grpc.DialContext(
		context.Background(),
		cfg.CategoryServiceAddr,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(mwclient.AddAppInfoUnary),
	)
	if err != nil {
		logger.ErrorKV(ctx, "failed to create client", "err", err)
	}

	conn, err := db.ConnectDB(&cfg.DB)
	if err != nil {
		logger.FatalKV(ctx, "sql.Open() error", "err", err)
	}
	defer conn.Close()

	categoryServiceClient := grpc_category_service.NewCategoryServiceClient(categoryServiceConn)

	productService := product_service.NewService(categoryServiceClient, conn)

	if err := server.NewGrpcServer(productService).Start(&cfg); err != nil {
		logger.ErrorKV(ctx, "Failed creating gRPC server", "err", err)

		return
	}
}

func initLogger(ctx context.Context, cfg config.Config) (syncFn func()) {
	loggingLevel := zap.InfoLevel
	if cfg.Project.Debug {
		loggingLevel = zap.DebugLevel
	}

	consoleCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		os.Stderr,
		zap.NewAtomicLevelAt(loggingLevel),
	)

	gelfCore, err := gelf.NewCore(
		gelf.Addr(cfg.Telemetry.GraylogPath),
		gelf.Level(loggingLevel),
	)
	if err != nil {
		logger.FatalKV(ctx, "sql.Open() error", "err", err)
	}

	notSugaredLogger := zap.New(zapcore.NewTee(consoleCore, gelfCore))

	sugaredLogger := notSugaredLogger.Sugar()
	logger.SetLogger(sugaredLogger.With(
		"service", cfg.Project.ServiceName,
	))

	return func() {
		notSugaredLogger.Sync()
	}
}
