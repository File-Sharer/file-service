package service

import (
	"context"
	"mime/multipart"

	pb "github.com/File-Sharer/file-service/hasher_pbs"
	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/mq"
	"github.com/File-Sharer/file-service/internal/repository"
)

type File interface {
	Create(ctx context.Context, fileObj *model.File, file *multipart.FileHeader) (*model.File, error)
	ProtectedFindByID(ctx context.Context, fileID string, userID string) (*model.File, error)
	FindByID(ctx context.Context, id string) (*model.File, error)
	FindUserFiles(ctx context.Context, userID string) ([]*model.File, error)
	AddPermission(ctx context.Context, data *AddPermissionData) error
	Delete(ctx context.Context, fileID string, user *model.User) error
	DeletePermission(ctx context.Context, data *DeletePermissionData) error
	FindPermissionsToFile(ctx context.Context, fileID string) ([]*model.Permission, error)
	FilesDeleteConsumer()
}

type Service struct {
	File
}

func New(repo *repository.Repository, rabbitMQ *mq.Conn, hasherClient pb.HasherClient) *Service {
	return &Service{
		File: NewFileService(repo, rabbitMQ, hasherClient),
	}
}
