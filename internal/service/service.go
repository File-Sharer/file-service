package service

import (
	"context"
	"mime/multipart"

	pb "github.com/File-Sharer/file-service/hasher_pbs"
	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/rabbitmq"
	"github.com/File-Sharer/file-service/internal/repository"
	"go.uber.org/zap"
)

type UserSpace interface {
	FindUserSpace(ctx context.Context, userID string) (int64, error)
	StartCreatingUsersSpaces(ctx context.Context)
}

type File interface {
	Create(ctx context.Context, fileObj *model.File, file multipart.File, fileHeader *multipart.FileHeader) (*model.File, error)
	ProtectedFindByID(ctx context.Context, fileID string, userID string) (*model.File, error)
	FindByID(ctx context.Context, id string) (*model.File, error)
	FindUserFiles(ctx context.Context, userID string) ([]*model.File, error)
	AddPermission(ctx context.Context, data *AddPermissionData) error
	Delete(ctx context.Context, fileID string, user *model.User) error
	DeletePermission(ctx context.Context, data *DeletePermissionData) error
	FindPermissionsToFile(ctx context.Context, fileID, creatorID string) ([]*model.Permission, error)
	TogglePublic(ctx context.Context, id, creatorID string) error
}

type Service struct {
	logger *zap.Logger
	UserSpace
	File
}

func New(logger *zap.Logger, repo *repository.Repository, rabbitmq *rabbitmq.MQConn, hasherClient pb.HasherClient) *Service {
	return &Service{
		logger: logger,
		UserSpace: newUserSpaceService(logger, repo, rabbitmq),
		File: NewFileService(logger, repo, hasherClient),
	}
}

func (s *Service) StartAllWorkers(ctx context.Context) {
	go s.UserSpace.StartCreatingUsersSpaces(ctx)
	s.logger.Info("Started all workers")
}
