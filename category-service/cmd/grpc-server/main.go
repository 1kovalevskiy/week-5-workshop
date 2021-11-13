package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/ozonmp/week-5-workshop/category-service/internal/config"
	"github.com/ozonmp/week-5-workshop/category-service/internal/pkg/logger"
	"github.com/ozonmp/week-5-workshop/category-service/internal/server"
	"github.com/ozonmp/week-5-workshop/category-service/internal/service/category"
	cat_repository "github.com/ozonmp/week-5-workshop/category-service/internal/service/category/repository"
	"github.com/ozonmp/week-5-workshop/category-service/internal/service/database"
	"github.com/ozonmp/week-5-workshop/category-service/internal/service/task"
	task_repository "github.com/ozonmp/week-5-workshop/category-service/internal/service/task/repository"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

	initCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.New(initCtx, cfg.Database.DSN)
	if err != nil {
		logger.ErrorKV(ctx, "failed to create client", "err", err)
	}

	categoryRepository := cat_repository.New(db)
	categoryService := category.New(categoryRepository)

	taskRepository := task_repository.New(db)
	taskService := task.New(taskRepository, db)

	if err := server.NewGrpcServer(
		categoryService,
		taskService,
	).Start(&cfg); err != nil {
		logger.ErrorKV(ctx, "Failed creating gRPC server", "err", err)

		return
	}
}

func initLogger(ctx context.Context, cfg config.Config) (syncFn func()) {
	loggingLevel := zap.InfoLevel
	if cfg.Project.Debug {
		loggingLevel = zap.DebugLevel
	}

	consoleCore := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), os.Stderr, zap.NewAtomicLevelAt(loggingLevel))

	notSugaredLogger := zap.New(consoleCore)

	sugaredLogger := notSugaredLogger.Sugar()
	logger.SetLogger(sugaredLogger)

	return func() {
		notSugaredLogger.Sync()
	}
}
