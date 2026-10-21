package pin

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newTestService wires a pinService with fresh mocks. Unexpected mock calls fail
// the test, and all registered expectations are asserted on cleanup.
func newTestService(t *testing.T) (*pinService, *mockPinRepository, *mockImageStore) {
	t.Helper()
	repo := new(mockPinRepository)
	repo.Test(t)
	store := new(mockImageStore)
	store.Test(t)
	t.Cleanup(func() {
		repo.AssertExpectations(t)
		store.AssertExpectations(t)
	})
	return &pinService{repo: repo, validate: validator.New(), imgStorage: store}, repo, store
}

func newTestPin(id, owner uuid.UUID, status string) *PinResponse {
	return &PinResponse{
		Id:           id,
		User_id:      owner,
		Title:        "Sunset",
		Image_url:    "pins/" + owner.String() + "/image",
		Image_status: status,
	}
}

func TestPinService_CreatePin(t *testing.T) {
	userID := uuid.New()
	keyPattern := regexp.MustCompile(`^pins/` + userID.String() + `/[0-9a-f-]{36}$`)
	isImageKey := mock.MatchedBy(func(key string) bool { return keyPattern.MatchString(key) })
	validReq := CreatePinRequest{Title: "Sunset"}
	storageErr := errors.New("storage down")

	tests := []struct {
		name              string
		req               CreatePinRequest
		setup             func(repo *mockPinRepository, store *mockImageStore)
		wantErr           error
		wantValidationErr bool
	}{
		{
			name:              "invalid input",
			req:               CreatePinRequest{Title: ""},
			wantValidationErr: true,
		},
		{
			name: "upload url error",
			req:  validReq,
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				store.On("GenerateUploadURL", mock.Anything, isImageKey, 15*time.Minute).Return("", storageErr)
			},
			wantErr: storageErr,
		},
		{
			name: "success",
			req:  validReq,
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				store.On("GenerateUploadURL", mock.Anything, isImageKey, 15*time.Minute).Return("https://upload.example", nil)
				repo.On("CreatePin", mock.Anything, userID, isImageKey, validReq).
					Return(newTestPin(uuid.New(), userID, "pending"), nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, store := newTestService(t)
			if tt.setup != nil {
				tt.setup(repo, store)
			}

			resp, err := svc.CreatePin(context.Background(), userID, tt.req)

			switch {
			case tt.wantValidationErr:
				var valErr *errValidation
				require.ErrorAs(t, err, &valErr)
				assert.Contains(t, valErr.Details, "Title")
				assert.Nil(t, resp)
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, resp)
			default:
				require.NoError(t, err)
				assert.Equal(t, "https://upload.example", resp.Upload_url)
				assert.Equal(t, userID, resp.Pin.User_id)
				assert.Equal(t, "pending", resp.Pin.Image_status)
			}
		})
	}
}

func TestPinService_GetPinByID(t *testing.T) {
	pinID := uuid.New()
	owner := uuid.New()

	tests := []struct {
		name    string
		setup   func(repo *mockPinRepository, store *mockImageStore)
		wantErr error
		wantURL string
	}{
		{
			name: "not found",
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				repo.On("GetPinByID", mock.Anything, pinID.String()).Return(nil, errPinNotFound)
			},
			wantErr: errPinNotFound,
		},
		{
			// The object may not exist yet, so no download URL is generated
			name: "pending upload has no download url",
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				repo.On("GetPinByID", mock.Anything, pinID.String()).Return(newTestPin(pinID, owner, "pending"), nil)
			},
			wantURL: "",
		},
		{
			name: "confirmed upload has download url",
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				p := newTestPin(pinID, owner, "confirmed")
				repo.On("GetPinByID", mock.Anything, pinID.String()).Return(p, nil)
				store.On("GenerateDownloadURL", mock.Anything, p.Image_url, time.Hour).Return("https://download.example", nil)
			},
			wantURL: "https://download.example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, store := newTestService(t)
			tt.setup(repo, store)

			resp, err := svc.GetPinByID(context.Background(), pinID.String())

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, pinID, resp.Pin.Id)
			assert.Equal(t, tt.wantURL, resp.Download_url)
		})
	}
}

func TestPinService_UpdatePin(t *testing.T) {
	pinID := uuid.New()
	owner := uuid.New()
	title := "New title"
	validReq := UpdatePinRequest{Title: &title}
	tooLong := strings.Repeat("a", 256)

	tests := []struct {
		name              string
		userID            uuid.UUID
		req               UpdatePinRequest
		setup             func(repo *mockPinRepository)
		wantErr           error
		wantValidationErr bool
	}{
		{
			name:              "invalid input",
			userID:            owner,
			req:               UpdatePinRequest{Title: &tooLong},
			wantValidationErr: true,
		},
		{
			name:   "not the owner",
			userID: uuid.New(),
			req:    validReq,
			setup: func(repo *mockPinRepository) {
				repo.On("GetPinByID", mock.Anything, pinID.String()).Return(newTestPin(pinID, owner, "confirmed"), nil)
			},
			wantErr: errForbidden,
		},
		{
			name:   "success",
			userID: owner,
			req:    validReq,
			setup: func(repo *mockPinRepository) {
				repo.On("GetPinByID", mock.Anything, pinID.String()).Return(newTestPin(pinID, owner, "confirmed"), nil)
				repo.On("UpdatePin", mock.Anything, UpdatePinRequest{Id: pinID, Title: &title}).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _ := newTestService(t)
			if tt.setup != nil {
				tt.setup(repo)
			}

			err := svc.UpdatePin(context.Background(), pinID.String(), tt.userID, tt.req)

			switch {
			case tt.wantValidationErr:
				var valErr *errValidation
				require.ErrorAs(t, err, &valErr)
				assert.Contains(t, valErr.Details, "Title")
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestPinService_DeletePin(t *testing.T) {
	pinID := uuid.New()
	owner := uuid.New()

	tests := []struct {
		name    string
		userID  uuid.UUID
		setup   func(repo *mockPinRepository, store *mockImageStore)
		wantErr error
	}{
		{
			name:   "not the owner",
			userID: uuid.New(),
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				repo.On("GetPinByID", mock.Anything, pinID.String()).Return(newTestPin(pinID, owner, "confirmed"), nil)
			},
			wantErr: errForbidden,
		},
		{
			// The DB row is the source of truth - a storage failure is only logged
			name:   "storage cleanup failure is ignored",
			userID: owner,
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				p := newTestPin(pinID, owner, "confirmed")
				repo.On("GetPinByID", mock.Anything, pinID.String()).Return(p, nil)
				repo.On("DeletePin", mock.Anything, pinID.String()).Return(nil)
				store.On("RemoveObject", mock.Anything, p.Image_url).Return(errors.New("storage down"))
			},
		},
		{
			name:   "success",
			userID: owner,
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				p := newTestPin(pinID, owner, "confirmed")
				repo.On("GetPinByID", mock.Anything, pinID.String()).Return(p, nil)
				repo.On("DeletePin", mock.Anything, pinID.String()).Return(nil)
				store.On("RemoveObject", mock.Anything, p.Image_url).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, store := newTestService(t)
			tt.setup(repo, store)

			err := svc.DeletePin(context.Background(), pinID.String(), tt.userID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPinService_ConfirmUpload(t *testing.T) {
	pinID := uuid.New()
	owner := uuid.New()
	id := pinID.String()

	// pendingPin registers the ownership lookup shared by most cases
	pendingPin := func(repo *mockPinRepository) *PinResponse {
		p := newTestPin(pinID, owner, "pending")
		repo.On("GetPinByID", mock.Anything, id).Return(p, nil)
		return p
	}

	tests := []struct {
		name    string
		userID  uuid.UUID
		setup   func(repo *mockPinRepository, store *mockImageStore)
		wantErr error
	}{
		{
			name:   "not the owner",
			userID: uuid.New(),
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				pendingPin(repo)
			},
			wantErr: errForbidden,
		},
		{
			name:   "already confirmed is idempotent",
			userID: owner,
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				repo.On("GetPinByID", mock.Anything, id).Return(newTestPin(pinID, owner, "confirmed"), nil)
			},
		},
		{
			name:   "object never uploaded deletes pin",
			userID: owner,
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				p := pendingPin(repo)
				store.On("StatObject", mock.Anything, p.Image_url).
					Return(minio.ObjectInfo{}, minio.ErrorResponse{Code: "NoSuchKey"})
				repo.On("DeletePin", mock.Anything, id).Return(nil)
			},
			wantErr: errImageInvalid,
		},
		{
			name:   "storage unavailable keeps pin",
			userID: owner,
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				p := pendingPin(repo)
				store.On("StatObject", mock.Anything, p.Image_url).
					Return(minio.ObjectInfo{}, errors.New("connection refused"))
			},
			wantErr: errStorageUnavailable,
		},
		{
			name:   "disallowed content type is rejected",
			userID: owner,
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				p := pendingPin(repo)
				store.On("StatObject", mock.Anything, p.Image_url).
					Return(minio.ObjectInfo{ContentType: "text/html", Size: 1024}, nil)
				store.On("RemoveObject", mock.Anything, p.Image_url).Return(nil)
				repo.On("DeletePin", mock.Anything, id).Return(nil)
			},
			wantErr: errImageInvalid,
		},
		{
			name:   "oversized image is rejected",
			userID: owner,
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				p := pendingPin(repo)
				store.On("StatObject", mock.Anything, p.Image_url).
					Return(minio.ObjectInfo{ContentType: "image/png", Size: maxImageSizeBytes + 1}, nil)
				store.On("RemoveObject", mock.Anything, p.Image_url).Return(nil)
				repo.On("DeletePin", mock.Anything, id).Return(nil)
			},
			wantErr: errImageInvalid,
		},
		{
			name:   "valid upload is confirmed",
			userID: owner,
			setup: func(repo *mockPinRepository, store *mockImageStore) {
				p := pendingPin(repo)
				store.On("StatObject", mock.Anything, p.Image_url).
					Return(minio.ObjectInfo{ContentType: "image/jpeg", Size: 2 << 20}, nil)
				repo.On("UpdateImageStatus", mock.Anything, id, "confirmed").Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, store := newTestService(t)
			tt.setup(repo, store)

			resp, err := svc.ConfirmUpload(context.Background(), id, tt.userID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "confirmed", resp.Image_status)
		})
	}
}
