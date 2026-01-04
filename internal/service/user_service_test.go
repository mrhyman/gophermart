package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mrhyman/gophermart/internal/auth"
	"github.com/mrhyman/gophermart/internal/model"
	"github.com/mrhyman/gophermart/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserService_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		login := "newuser"
		password := "password123"

		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, user model.User) error {
				// Проверяем, что пароль был захеширован
				assert.NotEqual(t, password, user.Password)
				assert.Greater(t, len(user.Password), 50)
				assert.Equal(t, login, user.Login)
				assert.NotEqual(t, uuid.Nil, user.ID)
				return nil
			}).
			Times(1)

		userID, err := svc.Register(context.Background(), login, password)

		require.NoError(t, err)
		assert.NotEmpty(t, userID)

		// Проверяем, что вернулся валидный UUID
		_, err = uuid.Parse(userID)
		assert.NoError(t, err)
	})

	t.Run("user already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		login := "existinguser"
		password := "password123"

		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(model.NewAlreadyExistsError("user", login, nil)).
			Times(1)

		userID, err := svc.Register(context.Background(), login, password)

		assert.Error(t, err)
		assert.Empty(t, userID)
		var alreadyExistsErr *model.AlreadyExistsError
		assert.ErrorAs(t, err, &alreadyExistsErr)
	})

	t.Run("repository error on create", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		login := "testuser"
		password := "password123"

		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(assert.AnError).
			Times(1)

		userID, err := svc.Register(context.Background(), login, password)

		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("different users get different IDs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		var capturedIDs []uuid.UUID

		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, user model.User) error {
				capturedIDs = append(capturedIDs, user.ID)
				return nil
			}).
			Times(2)

		userID1, err1 := svc.Register(context.Background(), "user1", "pass1")
		userID2, err2 := svc.Register(context.Background(), "user2", "pass2")

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, userID1, userID2)
		assert.Len(t, capturedIDs, 2)
		assert.NotEqual(t, capturedIDs[0], capturedIDs[1])
	})
}

func TestUserService_Login(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		login := "testuser"
		password := "password123"
		expectedUserID := uuid.New()

		hashedPassword, err := auth.HashPassword(password)
		require.NoError(t, err)

		user := &model.User{
			ID:       expectedUserID,
			Login:    login,
			Password: hashedPassword,
		}

		mockRepo.EXPECT().
			GetByLogin(gomock.Any(), login).
			Return(user, nil).
			Times(1)

		userID, err := svc.Login(context.Background(), login, password)

		require.NoError(t, err)
		assert.Equal(t, expectedUserID.String(), userID)
	})

	t.Run("user not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		login := "nonexistent"
		password := "password123"

		mockRepo.EXPECT().
			GetByLogin(gomock.Any(), login).
			Return(nil, model.ErrNotFound).
			Times(1)

		userID, err := svc.Login(context.Background(), login, password)

		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Equal(t, model.ErrNotFound, err)
	})

	t.Run("wrong password", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		login := "testuser"
		correctPassword := "correctpass"
		wrongPassword := "wrongpass"

		hashedPassword, err := auth.HashPassword(correctPassword)
		require.NoError(t, err)

		user := &model.User{
			ID:       uuid.New(),
			Login:    login,
			Password: hashedPassword,
		}

		mockRepo.EXPECT().
			GetByLogin(gomock.Any(), login).
			Return(user, nil).
			Times(1)

		userID, err := svc.Login(context.Background(), login, wrongPassword)

		assert.ErrorIs(t, err, model.ErrInvalidCredentials)
		assert.Empty(t, userID)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		login := "testuser"
		password := "password123"

		mockRepo.EXPECT().
			GetByLogin(gomock.Any(), login).
			Return(nil, assert.AnError).
			Times(1)

		userID, err := svc.Login(context.Background(), login, password)

		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("empty password in database", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		login := "testuser"
		password := "password123"

		user := &model.User{
			ID:       uuid.New(),
			Login:    login,
			Password: "", // Пустой пароль
		}

		mockRepo.EXPECT().
			GetByLogin(gomock.Any(), login).
			Return(user, nil).
			Times(1)

		userID, err := svc.Login(context.Background(), login, password)

		assert.ErrorIs(t, err, model.ErrInvalidCredentials)
		assert.Empty(t, userID)
	})
}
