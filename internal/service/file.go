package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	pb "github.com/File-Sharer/file-service/hasher_pbs"
	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/repository"
	"github.com/File-Sharer/file-service/internal/repository/redisrepo"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type FileService struct {
	logger *zap.Logger
	repo *repository.Repository
	hasher pb.HasherClient
	httpClient *http.Client
	userSpaceService UserSpace
	rdb *redis.Client
	folderService Folder
}

func NewFileService(logger *zap.Logger, repo *repository.Repository, hasherClient pb.HasherClient, userSpaceService UserSpace, rdb *redis.Client, folderService Folder) *FileService {
	return &FileService{
		logger: logger,
		repo: repo,
		hasher: hasherClient,
		httpClient: &http.Client{},
		userSpaceService: userSpaceService,
		rdb: rdb,
		folderService: folderService,
	}
}

func (s *FileService) Create(ctx context.Context, fileObj model.File, file multipart.File, fileHeader *multipart.FileHeader) (*model.File, error) {
	// Checking user creating files delay
	delay := s.rdb.Get(ctx, FileCreateDelayPrefix(fileObj.CreatorID))
	if delay.Err() != redis.Nil {
		return nil, errWaitDelay
	}

	space, err := s.userSpaceService.Get(ctx, fileObj.CreatorID)
	if err != nil {
		return nil, err
	}
	
	if fileHeader.Size > levelSpaceSizes[space.Level].maxFileSize {
		return nil, errFileIsTooBig
	}
	
	spaceSize, err := s.userSpaceService.GetSize(ctx, fileObj.CreatorID)
	if err != nil {
		return nil, err
	}
	if spaceSize + fileHeader.Size > levelSpaceSizes[space.Level].maxSpaceSize {
		return nil, errYouDoNotHaveEnoughSpace
	}
	
	fileHashIDResp, err := s.hasher.Hash(ctx, &pb.HashReq{BaseString: fileObj.CreatorID})
	if !fileHashIDResp.GetOk() {
		s.logger.Sugar().Errorf("failed to hash user(%s)'s file ID: %s", fileObj.CreatorID, err.Error())
		return nil, errInternal
	}
	fileObj.ID = fileHashIDResp.GetHash()

	ext := filepath.Ext(fileHeader.Filename)
	fileObj.DateAdded = time.Now()

	// Validating file extension
	downloadFilenameExt := filepath.Ext(*fileObj.DownloadFilename)
	if downloadFilenameExt == "" || downloadFilenameExt != ext {
		*fileObj.DownloadFilename += ext
	}

	// Sending user to timeout
	if err := s.rdb.Set(ctx, FileCreateDelayPrefix(fileObj.CreatorID), 1, time.Minute * 2).Err(); err != nil {
		s.logger.Sugar().Errorf("failed to set user(%s) to timeout in redis: %s", fileObj.CreatorID, err.Error())
		return nil, errInternal
	}

	// Clear cache
	if err := s.rdb.Del(ctx, UserFilesPrefix(fileObj.CreatorID), SpaceSizePrefix(fileObj.CreatorID)).Err(); err != nil {
		s.logger.Sugar().Errorf("failed to clear user(%s) files cache in redis: %s", fileObj.CreatorID, err.Error())
		return nil, errInternal
	}

	path := fileObj.CreatorID
	if fileObj.FolderID != nil {
		folder, err := s.folderService.findByID(ctx, *fileObj.FolderID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}

		fileObj.MainFolderID = new(string)
		if folder.MainFolderID != nil {
			*fileObj.MainFolderID = *folder.MainFolderID
		} else {
			*fileObj.MainFolderID = folder.ID
		}

		fileObj.DownloadFilename = nil
		fileObj.Public = nil

		sep := fmt.Sprintf("%s/files/%s/folders/", viper.GetString("fileStorage.origin"), fileObj.CreatorID)
		path = fmt.Sprintf("%s/folders/%s", fileObj.CreatorID, strings.Split(folder.URL, sep)[1])
	}

	fileSize, fileURL, err := s.saveToFileStorage(path, file, fileHeader)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, errFailedToUploadFileToFileStorage
	}
	fileObj.Size = fileSize
	fileObj.URL = fileURL
	parts := strings.Split(fileURL, "/")
	fileObj.Filename = parts[len(parts)-1]

	if err := s.repo.Postgres.File.Create(ctx, &fileObj); err != nil {
		s.logger.Sugar().Errorf("failed to create file by user(%s) in postgres: %s", fileObj.CreatorID, err.Error())
		return nil, errInternal
	}

	return &fileObj, err
}

func (s *FileService) saveToFileStorage(path string, file multipart.File, fileHeader *multipart.FileHeader) (int64, string, error) {
	endpoint := "/files"
	url := viper.GetString("fileStorage.origin") + endpoint

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Writing text fields
	if err := writer.WriteField("path", path); err != nil {
		return 0, "", fmt.Errorf("failed to write 'path' field for file-storage request: %s", err.Error())
	}

	// Writing file
	fileWriter, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create file part for file-storage request: %s", err.Error())
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return 0, "", fmt.Errorf("failed to seek to the start of the file: %s", err.Error())
	}

	if _, err := io.Copy(fileWriter, file); err != nil {
		return 0, "", fmt.Errorf("failed to copy file content for file-storage request: %s", err.Error())
	}

	// End of request body
	if err := writer.Close(); err != nil {
		return 0, "", fmt.Errorf("failed to close writer for file-storage request: %s", err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create file-storage request: %s", err.Error())
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Internal-Token", os.Getenv("X_INTERNAL_TOKEN"))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to do file-storage request: %s", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read response body from file-storage: %s", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		var bodyJSON map[string]interface{}
        if err := json.Unmarshal(body, &bodyJSON); err != nil {
            return 0, "", fmt.Errorf("failed to decode error response from file-storage: %s", err.Error())
        }
		return 0, "", fmt.Errorf("ERROR from file-storage endpoint(%s), code(%d), details: %s", endpoint, resp.StatusCode, bodyJSON["details"])
	}

	var response uploadResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, "", fmt.Errorf("failed to unmarshal json response from file-storage: %s", err.Error())
	}

	return response.FileSize, response.URL, nil
}

type uploadResponse struct {
	Ok       bool   `json:"ok"`
	URL      string `json:"url"`
	FileSize int64  `json:"file_size"`
}

func (s *FileService) ProtectedFindByID(ctx context.Context, fileID string, user model.User) (*model.File, error) {
	file, err := s.FindByID(ctx, fileID)
	if err != nil {
		return nil, errInternal
	}

	if file.MainFolderID != nil {
		return nil, nil
	}

	if file.CreatorID == user.ID || user.Role == "ADMIN" {
		return file, nil
	}

	if *file.Public {
		return file, nil
	}

	permission, err := s.HasPermission(ctx, fileID, user.ID)
	if err != nil {
		return nil, err
	}

	if permission {
		return file, nil
	}

	return nil, errNoAccess
}

func (s *FileService) FindByID(ctx context.Context, id string) (*model.File, error) {
	fileCache, err := redisrepo.Get[model.File](s.rdb, ctx, FilePrefix(id))
	if err == nil {
		return fileCache, nil
	}
	if err != redis.Nil {
		return nil, err
	}

	file, err := s.repo.Postgres.File.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Caching result
	if err := redisrepo.SetJSON(s.rdb, ctx, FilePrefix(file.ID), file,  time.Hour); err != nil {
		return nil, err
	}

	return file, nil
}

func (s *FileService) FindUserFiles(ctx context.Context, userID string) ([]*model.File, error) {
	filesCache, err := redisrepo.GetMany[model.File](s.rdb, ctx, UserFilesPrefix(userID))
	if err == nil {
		return filesCache, nil
	}
	if err != redis.Nil {
		return nil, err
	}

	files, err := s.repo.Postgres.File.FindUserFiles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Caching result
	if err := redisrepo.SetJSON(s.rdb, ctx, UserFilesPrefix(userID), files, time.Minute * 5); err != nil {
		return nil, err
	}

	return files, nil
}

func (s *FileService) HasPermission(ctx context.Context, fileID string, userID string) (bool, error) {
	permissionCache, err := s.rdb.Get(ctx, FilePermissionPrefix(fileID, userID)).Bool()
	if err == nil {
		return permissionCache, nil
	}
	if err != redis.Nil {
		return false, err
	}

	hasPermission, err := s.repo.Postgres.File.HasPermission(ctx, fileID, userID)
	if err != nil {
		return false, err
	}

	if err := s.rdb.Set(ctx, FilePermissionPrefix(fileID, userID), hasPermission, time.Minute).Err(); err != nil {
		return false, err
	}

	return hasPermission, nil
}

func (s *FileService) AddPermission(ctx context.Context, data AddPermissionData) error {
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

	// Clear cache
	if err := s.rdb.Del(ctx, FilePermissionPrefix(data.FileID, data.UserToAddID), FilePermissionsPrefix(data.FileID)).Err(); err != nil {
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

func (s *FileService) Delete(ctx context.Context, fileID string, user model.User) error {
	file, err := s.ProtectedFindByID(ctx, fileID, user)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errFileNotFound
		}
		s.logger.Sugar().Errorf("failed to find file(%s) in postgres: %s", fileID, err.Error())
		return errInternal
	}

	if file.CreatorID != user.ID && user.Role != "ADMIN" {
		return errNoAccess
	}
	
	if err := s.deleteFiles([]string{fmt.Sprintf("%s/%s", user.ID, file.Filename)}); err != nil {
		s.logger.Error(err.Error())
		return errInternal
	}

	if err := s.repo.Postgres.File.Delete(ctx, fileID); err != nil {
		s.logger.Sugar().Errorf("failed to delete file(%s) from postgres: %s", fileID, err.Error())
		return errInternal
	}

	if err := s.rdb.Del(ctx, FilePrefix(fileID), UserFilesPrefix(user.ID), SpaceSizePrefix(user.ID)).Err(); err != nil {
		s.logger.Sugar().Errorf("failed to delete file(%s) from redis: %s", fileID, err.Error())
		return errInternal
	}

	return nil
}

func (s *FileService) deleteFiles(paths []string) error {
	endpoint := "/files"
	url := viper.GetString("fileStorage.origin") + endpoint

	jsonBody, err := json.Marshal(paths)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON request body: %s", err.Error())
	}
	
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create new HTTP request for file-storage: %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", os.Getenv("X_INTERNAL_TOKEN"))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do file-storage request: %s", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body from file-storage: %s", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		var bodyJSON map[string]interface{}
        if err := json.Unmarshal(body, &bodyJSON); err != nil {
            return fmt.Errorf("failed to decode error response from file-storage: %s", err.Error())
        }
		return fmt.Errorf("ERROR from file-storage endpoint(%s), code(%d), details: %s", endpoint, resp.StatusCode, bodyJSON["details"])
	}

	return nil
}

func (s *FileService) DeletePermission(ctx context.Context, data DeletePermissionData) error {
	file, err := s.FindByID(ctx, data.FileID)
	if err != nil {
		return err
	}

	if data.UserID != file.CreatorID {
		return errNoAccess
	}

	// Clear cache
	if err := s.rdb.Del(ctx, FilePermissionPrefix(file.ID, data.UserToDeleteID), FilePermissionsPrefix(file.ID)).Err(); err != nil {
		return err
	}

	err = s.repo.Postgres.File.DeletePermission(ctx, data.FileID, data.UserToDeleteID)
	return err
}

func (s *FileService) FindPermissionsToFile(ctx context.Context, fileID, creatorID string) ([]*model.Permission, error) {
	permissionsCache, err := redisrepo.GetMany[model.Permission](s.rdb, ctx, FilePermissionsPrefix(fileID))
	if err == nil {
		return permissionsCache, nil
	}
	if err != redis.Nil {
		return nil, err
	}

	permissions, err := s.repo.Postgres.File.FindPermissionsToFile(ctx, fileID, creatorID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	// Caching result
	if err := redisrepo.SetJSON(s.rdb, ctx, FilePermissionsPrefix(fileID), permissions, time.Hour); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (s *FileService) TogglePublic(ctx context.Context, id, creatorID string) error {
	file, err := s.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Postgres.File.TogglePublic(ctx, id, creatorID); err != nil {
		s.logger.Sugar().Errorf("failed to toggle file(%s) public field value in postgres: %s", id, err.Error())
		return errInternal
	}

	if !*file.Public {
		if err := s.repo.Postgres.File.ClearPermissions(ctx, id, creatorID); err != nil {
			s.logger.Sugar().Errorf("failed to clear file(%s) permissions in postgres: %s", id, err.Error())
			return errInternal
		}
	}

	if err := s.rdb.Del(ctx, FilePrefix(id), UserFilesPrefix(creatorID), FilePermissionsPrefix(id)).Err(); err != nil {
		s.logger.Sugar().Errorf("failed to clear redis cache(file: %s): %s", id, err.Error())
	}

	return nil
}
