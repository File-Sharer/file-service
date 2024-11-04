package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	pb "github.com/File-Sharer/file-service/hasher_pbs"
	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/mq"
	"github.com/File-Sharer/file-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type FileService struct {
	repo *repository.Repository
	rabbitMQ *mq.Conn
	hasher pb.HasherClient
}

func NewFileService(repo *repository.Repository, rabbitMQ *mq.Conn, hasherClient pb.HasherClient) *FileService {
	return &FileService{
		repo: repo,
		rabbitMQ: rabbitMQ,
		hasher: hasherClient,
	}
}

func (s *FileService) Create(ctx context.Context, fileObj *model.File, file *multipart.FileHeader) (*model.File, error) {
	// Checking user creating files delay
	delay := s.repo.Redis.Default.Get(ctx, FileCreateDelayPrefix(fileObj.CreatorID))
	if delay.Err() != redis.Nil {
		return nil, errWaitDelay
	}

	userFilesDir := "files/" + fileObj.CreatorID
	if err := os.MkdirAll(userFilesDir, os.ModePerm); err != nil {
		return nil, err
	}
	
	if file.Size > MAX_FILE_SIZE {
		return nil, errFileIsTooBig
	}

	// Checking user uploads size limit
	userFilesDirSize, err := getDirSize(userFilesDir)
	if err != nil {
		return nil, err
	}
	if userFilesDirSize + file.Size >= MAX_USER_FILES_DIR_SIZE {
		return nil, errMaxUploadsReached
	}
	
	fileHashIDResp, err := s.hasher.Hash(ctx, &pb.HashReq{BaseString: fileObj.CreatorID})
	if !fileHashIDResp.GetOk() {
		return nil, err
	}
	fileObj.ID = fileHashIDResp.GetHash()

	ext := filepath.Ext(file.Filename)
	filename := fileObj.ID + ext
	fileObj.Filename = filename
	fileObj.DateAdded = time.Now()

	// Validating file extension
	downloadFilenameExt := filepath.Ext(fileObj.DownloadFilename)
	if downloadFilenameExt == "" || downloadFilenameExt != ext {
		fileObj.DownloadFilename += ext
	}

	// Sending user to timeout
	if err := s.repo.Redis.Default.Set(ctx, FileCreateDelayPrefix(fileObj.CreatorID), 1, time.Minute * 2); err != nil {
		return nil, err
	}

	// Clear cache
	if err := s.repo.Redis.File.Delete(ctx, UserFilesPrefix(fileObj.CreatorID)); err != nil {
		return nil, err
	}

	if err := s.repo.Postgres.File.Create(ctx, fileObj); err != nil {
		return nil, err
	}

	file.Filename = filename
	err = saveFile(file, userFilesDir)
	return fileObj, err
}

func getDirSize(dir string) (int64, error) {
	var size int64
	err := filepath.Walk(dir, func (path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func saveFile(file *multipart.FileHeader, dist string) error {
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

func (s *FileService) ProtectedFindByID(ctx context.Context, fileID string, userID string) (*model.File, error) {
	file, err := s.FindByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	if file.CreatorID == userID {
		return file, nil
	}

	if file.IsPublic {
		return file, nil
	}

	permission, err := s.HasPermission(ctx, fileID, userID)
	if err != nil {
		return nil, err
	}

	if permission {
		return file, nil
	}

	return nil, errNoAccess
}

func (s *FileService) FindByID(ctx context.Context, id string) (*model.File, error) {
	file, err := s.repo.Redis.File.Find(ctx, FilePrefix(id))
	if err == nil {
		return file, nil
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

	if err := s.repo.Redis.File.Create(ctx, FilePrefix(fileDB.ID), fileJSON,  time.Hour * 12); err != nil {
		return nil, err
	}

	return fileDB, nil
}

func (s *FileService) FindUserFiles(ctx context.Context, userID string) ([]*model.File, error) {
	files, err := s.repo.Redis.File.FindMany(ctx, UserFilesPrefix(userID))
	if err == nil {
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

	hasPermissionDB, err := s.repo.Postgres.HasPermission(ctx, fileID, userID)
	if err != nil {
		return false, err
	}

	if err := s.repo.Redis.File.Create(ctx, PermissionPrefix(fileID, userID), []byte(strconv.FormatBool(hasPermissionDB)), time.Hour * 24); err != nil {
		return false, err
	}

	return hasPermissionDB, nil
}

func (s *FileService) AddPermission(ctx context.Context, data *AddPermissionData) error {
	file, err := s.FindByID(ctx, data.FileID)
	if err != nil {
		return err
	}

	if file.CreatorID != data.UserID {
		return errNoAccess
	}

	if data.UserToAddID == file.CreatorID {
		return errCantAddPermissionForYourself
	}

	if err := checkUserExistence(data.UserToken, data.UserToAddID); err != nil {
		return err
	}

	if err := s.repo.Redis.File.Delete(ctx, PermissionPrefix(data.FileID, data.UserToAddID)); err != nil {
		return err
	}

	err = s.repo.Postgres.File.AddPermission(ctx, data.FileID, data.UserToAddID)
	return err
}

func checkUserExistence(token string, userID string) error {
	host := viper.GetString("userService.host")
	endpoint := "/api/user/" + userID

	client := &http.Client{}

	req := &http.Request{
		Proto: "HTTP/1.1",
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host: host,
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

func (s *FileService) Delete(ctx context.Context, fileID string, user *model.User) error {
	file, err := s.FindByID(ctx, fileID)
	if err != nil {
		return err
	}

	if file.CreatorID != user.ID && user.Role != "ADMIN" {
		return errNoAccess
	}

	if err := s.repo.Redis.File.Delete(ctx, FilePrefix(fileID), UserFilesPrefix(user.ID)); err != nil {
		return err
	}

	msg, err := json.Marshal(model.DeleteFileReq{
		FileID: file.ID,
		Path: fmt.Sprintf("files/%s/%s", user.ID, file.Filename),
	})
	if err != nil {
		return err
	}
	err = s.rabbitMQ.Publish(mqFilesDelete, msg)
	return err
}

func (s *FileService) DeletePermission(ctx context.Context, data *DeletePermissionData) error {
	file, err := s.FindByID(ctx, data.FileID)
	if err != nil {
		return err
	}

	if data.UserID != file.CreatorID {
		return errNoAccess
	}

	// Clear cache
	if err := s.repo.Redis.File.Delete(ctx, PermissionPrefix(file.ID, data.UserToDeleteID)); err != nil {
		return err
	}

	err = s.repo.Postgres.File.DeletePermission(ctx, data.FileID, data.UserToDeleteID)
	return err
}

func (s *FileService) FilesDeleteConsumer() {
	msgs, err := s.rabbitMQ.Consume(mqFilesDelete)
	if err != nil {
		logrus.Fatalf("failed to start consumer: %s", err.Error())
	}

	go func ()  {
		for msg := range msgs {
			var message model.DeleteFileReq
			if err := json.Unmarshal(msg.Body, &message); err != nil {
				logrus.Errorf("failed unmarshal message: %s", err.Error())
				msg.Nack(false, true)
				continue
			}

			if err := s.repo.Postgres.File.Delete(context.Background(), message.FileID); err != nil {
				logrus.Errorf("failed delete file from database: %s", err.Error())
				msg.Nack(false, true)
				continue
			}

			if err := deleteFile(message.Path); err != nil {
				logrus.Errorf("failed delete file by path(%s): %s", message.Path, err.Error())
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
			logrus.Print("file deleted successfully!")
		}
	}()
}

func deleteFile(path string) error {
	return os.Remove(path)
}
