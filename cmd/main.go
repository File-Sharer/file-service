package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/File-Sharer/file-service/hasher_pbs"
	"github.com/File-Sharer/file-service/internal/config"
	"github.com/File-Sharer/file-service/internal/handler"
	"github.com/File-Sharer/file-service/internal/rabbitmq"
	"github.com/File-Sharer/file-service/internal/repository"
	"github.com/File-Sharer/file-service/internal/repository/postgres"
	"github.com/File-Sharer/file-service/internal/server"
	"github.com/File-Sharer/file-service/internal/service"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger, _ := zap.NewProduction()

	if err := initConfig(); err != nil {
		logger.Sugar().Fatalf("error initializing config: %s", err.Error())
	}

	if err := initEnv(); err != nil {
		logger.Sugar().Fatalf("error initializing env: %s", err.Error())
	}

	hasherServiceConn, err := grpc.NewClient(viper.GetString("hasherService.host"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Sugar().Fatalf("error connecting to grpc hasher service: %s", err.Error())
	}
	defer func ()  {
		if err := hasherServiceConn.Close(); err != nil {
			logger.Sugar().Fatalf("error occured on grpc hasher service connection close: %s", err.Error())
		}
	}()

	hasherClient := pb.NewHasherClient(hasherServiceConn)

	dbConfig := &config.DBConfig{
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		Host: os.Getenv("DB_HOST"),
		Port: os.Getenv("DB_PORT"),
		DBName: os.Getenv("DB_NAME"),
		SSLMode: os.Getenv("DB_SSLMODE"),
	}
	db, err := postgres.NewPgPool(context.Background(), dbConfig)
	if err != nil {
		logger.Sugar().Fatalf("error connecting to postgresql: %s", err.Error())
	}
	defer func ()  {
		db.Close()
	}()

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	defer func ()  {
		if err := rdb.Close(); err != nil {
			logger.Sugar().Fatalf("error occured on redis db connection close: %s", err.Error())
		}
	}()

	rabbitmq, err := rabbitmq.New(os.Getenv("RABBITMQ_URI"))
	if err != nil {
		logger.Sugar().Fatalf("error connection to rabbitmq: %s", err.Error())
	}

	repo := repository.New(db)
	services := service.New(logger, repo, rabbitmq, hasherClient, rdb)
	handlers := handler.New(services, hasherClient)

	services.StartAllWorkers(context.Background())

	srv := server.New()
	serverConfig := &config.ServerConfig{
		Port: viper.GetString("app.port"),
		Handler: handlers.InitRoutes(),
		MaxHeaderBytes: 1 << 20,
		ReadTimeout: time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	go func ()  {
		if err := srv.Run(serverConfig); err != nil {
			logger.Sugar().Fatalf("error running server: %s", err.Error())
		}
	}()

	logger.Info("File Server Started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info("File Server Shutting Down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Sugar().Fatalf("error shutting down server: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func initEnv() error {
	return godotenv.Load()
}
