package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/rabbitmq"
	"github.com/File-Sharer/file-service/internal/repository"
	"github.com/File-Sharer/file-service/internal/repository/redisrepo"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type userSpaceService struct {
	logger *zap.Logger
	repo   *repository.Repository
	rabbitmq *rabbitmq.MQConn
	rdb *redis.Client
}

func newUserSpaceService(logger *zap.Logger, repo *repository.Repository, rabbitmq *rabbitmq.MQConn, rdb *redis.Client) UserSpace {
	return &userSpaceService{
		logger: logger,
		repo: repo,
		rabbitmq: rabbitmq,
		rdb: rdb,
	}
}

func (s *userSpaceService) Get(ctx context.Context, userID string) (*model.FullUserSpace, error) {
	spaceCache, err := redisrepo.Get[model.FullUserSpace](s.rdb, ctx, SpacePrefix(userID))
	if err == nil {
		return spaceCache, nil
	}
	if err != redis.Nil {
		s.logger.Sugar().Errorf("failed to get user(%s) space from redis: %s", userID, err.Error())
		return nil, errInternal
	}

	space, err := s.repo.Postgres.UserSpace.GetFull(ctx, userID)
	if err != nil {
		s.logger.Sugar().Errorf("failed to get user(%s) space from postgres: %s", userID, err.Error())
		return nil, errInternal
	}

	if err := redisrepo.SetJSON(s.rdb, ctx, SpacePrefix(userID), space, time.Minute * 30); err != nil {
		s.logger.Sugar().Errorf("failed to set user(%s) space in redis: %s", userID, err.Error())
	}

	return space, nil
}

func (s *userSpaceService) GetSize(ctx context.Context, userID string) (int64, error) {
	spaceSizeCache, err := s.rdb.Get(ctx, SpaceSizePrefix(userID)).Int64()
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

	if err := s.rdb.Set(ctx, SpaceSizePrefix(userID), spaceSize, time.Minute * 5).Err(); err != nil {
		s.logger.Sugar().Errorf("failed to set user(%s) space size in redis: %s", userID, err.Error())
		return 0, errInternal
	}

	return spaceSize, nil
}

type userCreated struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
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

		if err := s.repo.Postgres.UserSpace.Create(ctx, model.UserSpace{UserID: data.UserID, Username: data.Username}); err != nil {
			s.logger.Sugar().Errorf("failed to create user(%s) space in postgres: %s", data.UserID, err.Error())
			msg.Nack(false, true)
			continue
		}

		msg.Ack(false)
	}
}

func (s *userSpaceService) UpdateLevel(ctx context.Context, userID string, newLevel uint8) error {
	return s.repo.Postgres.UserSpace.UpdateLevel(ctx, userID, newLevel)
}

func (s *userSpaceService) getByUsername(ctx context.Context, username string) (*model.UserSpace, error) {
	spaceCache, err := redisrepo.Get[model.UserSpace](s.rdb, ctx, SpaceByUsernamePrefix(username))
	if err == nil {
		return spaceCache, nil
	}
	if err != redis.Nil {
		s.logger.Sugar().Errorf("failed to get user space by username(%s) from redis: %s", username, err.Error())
		return nil, errInternal
	}

	space, err := s.repo.Postgres.UserSpace.GetByUsername(ctx, username)
	if err != nil {
		s.logger.Sugar().Errorf("failed to get user space by username(%s) from postgres: %s", username, err.Error())
		return nil, errInternal
	}

	if err := redisrepo.SetJSON(s.rdb, ctx, SpaceByUsernamePrefix(username), space, time.Minute * 2); err != nil {
		s.logger.Sugar().Errorf("failed to set user(%s) space in redis: %s", username, err.Error())
	}

	return space, nil
}
