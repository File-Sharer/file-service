package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	pb "github.com/File-Sharer/file-service/hasher_pbs"
	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/repository"
	"github.com/File-Sharer/file-service/internal/repository/redisrepo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type folderService struct {
	logger *zap.Logger
	repo *repository.Repository
	hasher pb.HasherClient
	rdb *redis.Client
	httpClient *http.Client
	userSpaceService UserSpace
}

func newFolderService(logger *zap.Logger, repo *repository.Repository, hasher pb.HasherClient, rdb *redis.Client, userSpaceService UserSpace) Folder {
	return &folderService{
		logger: logger,
		repo: repo,
		hasher: hasher,
		rdb: rdb,
		httpClient: &http.Client{},
		userSpaceService: userSpaceService,
	}
}

func (s *folderService) Create(ctx context.Context, f model.Folder) (*model.Folder, error) {
	resp, err := s.hasher.Hash(ctx, &pb.HashReq{BaseString: f.CreatorID})
	if err != nil || !resp.Ok {
		s.logger.Sugar().Errorf("failed to hash for user(%s)'s new folder: %s", f.CreatorID, err.Error())
		return nil, errInternal
	}
	f.ID = resp.GetHash()

	path := fmt.Sprintf("files/%s/folders/%s", f.CreatorID, f.Name)
	if f.FolderID != nil {
		f.Public = nil

		parentFolder, err := s.repo.Postgres.Folder.FindByID(ctx, *f.FolderID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, nil
			}
			s.logger.Sugar().Errorf("failed to find folder(%s) in postgres: %s", *f.FolderID, err.Error())
			return nil, errInternal
		}
		f.MainFolderID = new(string)
		if parentFolder.MainFolderID == nil {
			*f.MainFolderID = parentFolder.ID
		} else {
			*f.MainFolderID = *parentFolder.MainFolderID
		}
		f.Public = nil

		path = strings.Split(parentFolder.URL, viper.GetString("fileStorage.origin") + "/")[1] + "/" + f.Name
	}

	f.CreatedAt = time.Now()
	f.URL = fmt.Sprintf("%s/%s", viper.GetString("fileStorage.origin"), path)

	if err := s.repo.Postgres.Folder.Create(ctx, f); err != nil {
		s.logger.Sugar().Errorf("failed to create folder for user(%s) in postgres: %s", f.CreatorID, err.Error())
		return nil, errInternal
	}

	if err := s.createFolderInFS(path); err != nil {
		s.logger.Error(err.Error())
		return nil, errInternal
	}

	return &f, nil
}

type createFolderReq struct {
	Path string `json:"path"`
}

func (s *folderService) createFolderInFS(path string) error {
	endpoint := "/folders"
	url := viper.GetString("fileStorage.origin") + endpoint

	bodyJSON, err := json.Marshal(createFolderReq{Path: path})
	if err != nil {
		return fmt.Errorf("failed to marshal JSON request body: %s", err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("failed to create request for file-storage: %s", err.Error())
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

func (s *folderService) findByID(ctx context.Context, id string) (*model.Folder, error) {
	folderCache, err := redisrepo.Get[model.Folder](s.rdb, ctx, FolderPrefix(id))
	if err == nil {
		return folderCache, nil
	}
	if err != redis.Nil {
		s.logger.Sugar().Errorf("failed to get folder(%s) from redis: %s", id, err.Error())
		return nil, errInternal
	}

	folder, err := s.repo.Postgres.Folder.FindByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		s.logger.Sugar().Errorf("failed to find folder(%s) in postgres: %s", id, err.Error())
		return nil, errInternal
	}

	if err := redisrepo.SetJSON(s.rdb, ctx, FolderPrefix(id), folder, time.Minute); err != nil {
		s.logger.Sugar().Errorf("failed to set folder(%s) in redis: %s", id, err.Error())
	}

	return folder, nil
}

func (s *folderService) ProtectedFindByID(ctx context.Context, id, userRole string, userSpace model.UserSpace) (*model.Folder, error) {
	folder, err := s.findByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if folder.MainFolderID != nil {
		return nil, nil
	}

	if folder.CreatorID == userSpace.UserID || userRole == "ADMIN" {
		return folder, nil
	}

	if folder.Public != nil && *folder.Public {
		return folder, nil
	}

	permission, err := s.hasPermission(ctx, id, userSpace.Username)
	if err != nil {
		return nil, err
	}

	if permission {
		return folder, nil
	}

	return nil, errNoAccess
}

func (s *folderService) hasPermission(ctx context.Context, id, username string) (bool, error) {
	permissionCache, err := s.rdb.Get(ctx, FolderPermissionPrefix(id, username)).Bool()
	if err == nil {
		return permissionCache, nil
	}
	if err != redis.Nil {
		s.logger.Sugar().Errorf("failed to get folder(%s) permission for user(%s) from redis: %s", id, username, err.Error())
		return false, errInternal
	}

	permission, err := s.repo.Postgres.Folder.HasPermission(ctx, id, username)
	if err != nil {
		s.logger.Sugar().Errorf("failed to find folder(%s) permisson for user(%s) in postgres: %s", id, username, err.Error())
		return false, errInternal
	}

	if err := s.rdb.Set(ctx, FolderPermissionPrefix(id, username), permission, time.Minute).Err(); err != nil {
		s.logger.Sugar().Errorf("failed to set folder(%s) permission for user(%s) in redis: %s", id, username, err.Error())
	}

	return permission, nil
}

func (s *folderService) Rename(ctx context.Context, id, userID, newName string) error {
	if err := s.repo.Postgres.Folder.Update(ctx, id, map[string]interface{}{"name": newName}); err != nil {
		s.logger.Sugar().Errorf("failed to rename folder(%s) in postgres: %s", id, err.Error())
		return errInternal
	}

	if err := s.rdb.Del(ctx, FolderPrefix(id)).Err(); err != nil {
		s.logger.Sugar().Errorf("failed to delete cached folder(%s) data from redis: %s", id, err.Error())
	}

	return nil
}

func (s *folderService) GetFolderContents(ctx context.Context, id, userRole string, userSpace model.UserSpace) (*model.FolderContents, error) {
	folder, err := s.findByID(ctx, id)
	if err != nil {
		return nil, err
	}

	mainFolderID := id
	if folder.MainFolderID != nil {
		mainFolderID = *folder.MainFolderID
	}

	mainFolder, err := s.findByID(ctx, mainFolderID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := s.hasPermission(ctx, mainFolderID, userSpace.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errNoAccess
		}
		s.logger.Sugar().Errorf("failed to get permission for folder(%s) to user(%s) from postgres: %s", id, userSpace.UserID ,err.Error())
		return nil, errInternal
	}
	if !hasPermission && mainFolder.CreatorID != userSpace.UserID && userRole != "ADMIN" {
		return nil, errNoAccess
	}

	contentsCache, err := redisrepo.Get[model.FolderContents](s.rdb, ctx, FolderContentsPrefix(id))
	if err == nil {
		return contentsCache, nil
	}
	if err != redis.Nil {
		s.logger.Sugar().Errorf("failed to get folder(%s) contents from redis: %s", id, err.Error())
		return nil, errInternal
	}

	files, folders, err := s.repo.Postgres.Folder.GetFolderContents(ctx, id)
	if err != nil {
		s.logger.Sugar().Errorf("failed to get folder(%s) contents from postgres: %s", id, err.Error())
		return nil, errInternal
	}

	contents := model.FolderContents{
		Files: files,
		Folders: folders,
	}

	if err := redisrepo.SetJSON(s.rdb, ctx, FolderContentsPrefix(id), contents, time.Minute * 5); err != nil {
		s.logger.Sugar().Errorf("failed to set folder(%s) contents in redis: %s", id, err.Error())
	}

	return &contents, nil
}

func (s *folderService) GetUserFolders(ctx context.Context, userID string) ([]*model.Folder, error) {
	foldersCache, err := redisrepo.GetMany[model.Folder](s.rdb, ctx, UserFoldersPrefix(userID))
	if err == nil {
		return foldersCache, nil
	}
	if err != redis.Nil {
		s.logger.Sugar().Errorf("failed to get user(%s) folders from redis: %s", userID, err.Error())
		return nil, errInternal
	}

	folders, err := s.repo.Postgres.Folder.GetUserFolders(ctx, userID)
	if err != nil {
		s.logger.Sugar().Errorf("failed to get user(%s) folders from postgres: %s", userID, err.Error())
		return nil, errInternal	
	}

	if err := redisrepo.SetJSON(s.rdb, ctx, UserFoldersPrefix(userID), folders, time.Minute * 2); err != nil {
		s.logger.Sugar().Errorf("failed to set user(%s) folders in redis: %s", userID, err.Error())
	}

	return folders, nil
}

func (s *folderService) AddPermission(ctx context.Context, d AddPermissionData) error {
	folder, err := s.findByID(ctx, d.ResourceID)
	if err != nil {
		return err
	}

	// Skip if the folder is nested
	if folder.MainFolderID != nil {
		return nil
	}

	if d.UserSpace.UserID != folder.CreatorID && d.UserRole != "ADMIN" {
		return errNoAccess
	}

	if d.UserToAddName == d.UserSpace.Username {
		return errCantAddPermissionForYourself
	}

	// Clear cache
	if err := s.rdb.Del(ctx, FolderPermissionPrefix(d.ResourceID, d.UserToAddName), FolderPermissionsPrefix(d.ResourceID)).Err(); err != nil {
		return err
	}

	if err := s.repo.Postgres.Folder.AddPermission(ctx, d.ResourceID, d.UserToAddName); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23503" {
				return errUserNotFound
			}
		}
		
		s.logger.Sugar().Errorf("failed to add permisson to folder(%s) to user(%s) in postgres: %s", d.ResourceID, d.UserToAddName, err.Error())
		return errInternal
	}

	return nil
}

func (s *folderService) DeletePermission(ctx context.Context, d DeletePermissionData) error {
	folder, err := s.findByID(ctx, d.ResourceID)
	if err != nil {
		return err
	}

	if d.UserID != folder.CreatorID && d.UserRole != "ADMIN" {
		return errNoAccess
	}

	// Clear cache
	if err := s.rdb.Del(ctx, FolderPermissionPrefix(folder.ID, d.UserToDeleteName), FolderPermissionsPrefix(folder.ID)).Err(); err != nil {
		return err
	}

	return s.repo.Postgres.Folder.DeletePermission(ctx, folder.ID, d.UserToDeleteName)
}

func (s *folderService) GetPermissions(ctx context.Context, folderID, userID string) ([]*string, error) {
	permissionsCache, err := redisrepo.GetMany[string](s.rdb, ctx, FolderPermissionsPrefix(folderID))
	if err == nil {
		return permissionsCache, nil
	}
	if err != redis.Nil {
		s.logger.Sugar().Errorf("failed to get folder(%s) permissions from redis: %s", folderID, err.Error())
		return nil, errInternal
	}

	permissions, err := s.repo.Postgres.Folder.GetPermissions(ctx, folderID, userID)
	if err != nil {
		s.logger.Sugar().Errorf("faield to get folder(%s) permissions from postgres: %s", folderID, err.Error())
		return nil, errInternal
	}

	if err := redisrepo.SetJSON(s.rdb, ctx, FolderPermissionsPrefix(folderID), permissions, time.Minute * 3); err != nil {
		s.logger.Sugar().Errorf("failed to set folder(%s) permissions in redis: %s", folderID, err.Error())
	}

	return permissions, nil
}
