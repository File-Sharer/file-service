package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	pb "github.com/File-Sharer/file-service/hasher_pbs"
	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type FileService struct {
	repo *repository.Repository
	hasher pb.HasherClient
}

func NewFileService(repo *repository.Repository, hasherClient pb.HasherClient) *FileService {
	return &FileService{
		repo: repo,
		hasher: hasherClient,
	}
}

func (s *FileService) Create(ctx context.Context, fileObj *model.File, file *multipart.FileHeader) (*model.File, error) {
	if file.Size > MAX_FILE_SIZE {
		return nil, errFileIsTooBig
	}
	
	hash, err := s.hasher.Hash(ctx, &pb.HashReq{})
	if !hash.GetOk() {
		return nil, err
	}
	fileObj.ID = hash.GetHash()

	ext := filepath.Ext(file.Filename)
	filename := fileObj.ID + ext
	fileObj.Filename = filename
	fileObj.DateAdded = time.Now()

	if err := s.repo.Redis.File.Delete(ctx, userFilesPrefix + fileObj.CreatorID); err != nil {
		return nil, err
	}

	if err := s.repo.Postgres.File.Create(ctx, fileObj); err != nil {
		return nil, err
	}

	file.Filename = filename
	err = saveFile(file, "files")
	return fileObj, err
}

func saveFile(file *multipart.FileHeader, dist string) error {
	if err := os.MkdirAll(dist, os.ModePerm); err != nil  {
		return err
	}

	filePath := filepath.Join(dist, file.Filename)
	createdFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer createdFile.Close()

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = io.Copy(createdFile, src)
	return err
}

func (s *FileService) FindByID(ctx context.Context, id string, userID string) (*model.File, error) {
	file, err := s.repo.Redis.File.Find(ctx, filePrefix + id)
	if err == nil {
		if file.CreatorID == userID {
			return file, nil
		}

		if file.IsPublic {
			return file, nil
		}

		permission, err := s.HasPermission(ctx, id, userID)
		if err != nil {
			return nil, err
		}

		if permission {
			return file, nil
		}

		return nil, errNoAccess
	}

	if err != redis.Nil {
		return nil, err
	}

	fileDB, err := s.repo.Postgres.File.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	fileJSON, err := json.Marshal(fileDB)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Redis.File.Create(ctx, filePrefix + id, fileJSON, time.Hour * 12); err != nil {
		return nil, err
	}

	if fileDB.CreatorID == userID {
		return fileDB, nil
	}

	if fileDB.IsPublic {
		return fileDB, nil
	}

	permission, err := s.HasPermission(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if permission {
		return fileDB, nil
	}

	return nil, errNoAccess
}

func (s *FileService) FindUserFiles(ctx context.Context, userID string) ([]*model.File, error) {
	files, err := s.repo.Redis.File.FindMany(ctx, userFilesPrefix + userID)
	if err == nil {
		fmt.Println("HELLO USER FILES FROM REDIS")
		return files, nil
	}

	if err != redis.Nil {
		return nil, err
	}

	filesDB, err := s.repo.Postgres.File.FindUserFiles(ctx, userID)
	if err != nil {
		return nil, err
	}

	filesJSON, err := json.Marshal(filesDB)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Redis.File.Create(ctx, userFilesPrefix + userID, filesJSON, time.Hour * 12); err != nil {
		return nil, err
	}

	fmt.Println("HELLO USER FILES FROM POSTGRES")
	return filesDB, nil
}

func (s *FileService) HasPermission(ctx context.Context, fileID string, userID string) (bool, error) {
	permission, err := s.repo.Redis.File.HasPermission(ctx, PermissionPrefix(fileID, userID))
	if err == nil {
		return permission, nil
	}

	if err != redis.Nil {
		return false, err
	}

	permissionDB, err := s.repo.Postgres.HasPermission(ctx, fileID, userID)
	if err != nil {
		return false, err
	}

	if err := s.repo.Redis.File.Create(ctx, PermissionPrefix(fileID, userID), []byte(strconv.FormatBool(permissionDB)), time.Hour * 24); err != nil {
		return false, err
	}

	return permissionDB, nil
}

func (s *FileService) AddPermission(ctx context.Context, data *AddPermissionData) error {
	file, err := s.FindByID(ctx, data.FileID, data.UserID)
	if err != nil {
		return err
	}

	if file.CreatorID != data.UserID {
		return errNoAccess
	}

	err = doReq(data.UserToken, data.UserToAdd)
	if err != nil {
		return err
	}

	if err := s.repo.Redis.File.Delete(ctx, PermissionPrefix(data.FileID, data.UserToAdd)); err != nil {
		return err
	}

	err = s.repo.Postgres.File.AddPermission(ctx, data.FileID, data.UserToAdd)
	return err
}

func doReq(token string, userToAdd string) error {
	target := viper.GetString("userService.target")
	endpoint := "/api/user/" + userToAdd

	client := &http.Client{}

	req := &http.Request{
		Proto: "HTTP/1.1",
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host: target,
			Path: endpoint,
		},
		Header: map[string][]string{
			"Authorization": {"Bearer " + token},
		},
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errUserNotFound
	}

	var userRes model.UserRes
	if err := json.NewDecoder(res.Body).Decode(&userRes); err != nil {
		return err
	}

	if !userRes.Ok {
		return errUserNotFound
	}

	return nil
}
