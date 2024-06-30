package service

import (
	"context"
	"mime/multipart"

	pb "github.com/File-Sharer/file-service/hasher_pbs"
	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/repository"
)

type File interface {
	Create(ctx context.Context, fileObj *model.File, file *multipart.FileHeader) error
	FindByID(ctx context.Context, id string, userID string) (*model.File, error)
	FindUserFiles(ctx context.Context, userID string) ([]*model.File, error)
	AddPermission(ctx context.Context, fileID string, userID string, userToAdd string) error
}

type Service struct {
	File
}

func New(repo *repository.Repository, hasherClient pb.HasherClient) *Service {
	return &Service{
		File: NewFileService(repo, hasherClient),
	}
}
