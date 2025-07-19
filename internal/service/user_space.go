package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/rabbitmq"
	"github.com/File-Sharer/file-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type userSpaceService struct {
	logger *zap.Logger
	repo   *repository.Repository
	rabbitmq *rabbitmq.MQConn
}

func newUserSpaceService(logger *zap.Logger, repo *repository.Repository, rabbitmq *rabbitmq.MQConn) UserSpace {
	return &userSpaceService{
		logger: logger,
		repo: repo,
		rabbitmq: rabbitmq,
	}
}

func (s *userSpaceService) GetSize(ctx context.Context, userID string) (int64, error) {
	spaceSizeCache, err := s.repo.Redis.Default.Get(ctx, SpaceSizePrefix(userID)).Int64()
	if err == nil {
		return spaceSizeCache, nil
	}
	if err != redis.Nil {
		s.logger.Sugar().Errorf("failed to get user(%s) space size from redis: %s", userID, err.Error())
		return 0, errInternal
	}

	spaceSize, err := s.repo.Postgres.UserSpace.GetSize(ctx, userID)
	if err != nil {
		s.logger.Sugar().Errorf("failed to get user(%s) space size from postgres: %s", userID, err.Error())
		return 0, errInternal
	}

	if err := s.repo.Redis.Default.Set(ctx, SpaceSizePrefix(userID), spaceSize, time.Minute * 5); err != nil {
		s.logger.Sugar().Errorf("failed to set user(%s) space size in redis: %s", userID, err.Error())
		return 0, errInternal
	}

	return spaceSize, nil
}

type userCreated struct {
	UserID string `json:"userId"`
}

func (s *userSpaceService) StartCreatingUsersSpaces(ctx context.Context) {
	msgs, err := s.rabbitmq.ConsumeExchange(rabbitmq.USERS_CREATE_EXCHANGE)
	if err != nil {
		panic(err)
	}

	for msg := range msgs {
		var data userCreated
		if err := json.Unmarshal(msg.Body, &data); err != nil {
			s.logger.Sugar().Errorf("failed to unmarshal json: %s", err.Error())
			msg.Ack(false)
			continue
		}

		if err := s.repo.Postgres.UserSpace.Create(ctx, model.UserSpace{UserID: data.UserID}); err != nil {
			s.logger.Sugar().Errorf("failed to create user(%s) space in postgres: %s", data.UserID, err.Error())
			msg.Nack(false, true)
			continue
		}

		msg.Ack(false)
	}
}
