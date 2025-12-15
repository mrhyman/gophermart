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
			CreateUser(gomock.Any(), login, gomock.Any()).
			DoAndReturn(func(ctx context.Context, l, p string) error {
				// Проверяем, что пароль был захеширован
				assert.NotEqual(t, password, p)
				assert.Greater(t, len(p), 50)
				return nil
			}).
			Times(1)

		err := svc.Register(context.Background(), login, password)

		assert.NoError(t, err)
	})

	t.Run("user already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		login := "existinguser"
		password := "password123"

		// Репозиторий возвращает ошибку AlreadyExistsError
		mockRepo.EXPECT().
			CreateUser(gomock.Any(), login, gomock.Any()).
			Return(model.NewAlreadyExistsError("user", login, nil)).
			Times(1)

		err := svc.Register(context.Background(), login, password)

		assert.Error(t, err)
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
			CreateUser(gomock.Any(), login, gomock.Any()).
			Return(assert.AnError).
			Times(1)

		err := svc.Register(context.Background(), login, password)

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
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

		hashedPassword, err := auth.HashPassword(password)
		require.NoError(t, err)

		user := &model.User{
			ID:       uuid.New(),
			Login:    login,
			Password: hashedPassword,
		}

		mockRepo.EXPECT().
			GetUserByLogin(gomock.Any(), login).
			Return(user, nil)

		err = svc.Login(context.Background(), login, password)

		assert.NoError(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		login := "nonexistent"
		password := "password123"

		mockRepo.EXPECT().
			GetUserByLogin(gomock.Any(), login).
			Return(nil, model.ErrInvalidCredentials)

		err := svc.Login(context.Background(), login, password)

		assert.ErrorIs(t, err, model.ErrInvalidCredentials)
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
			GetUserByLogin(gomock.Any(), login).
			Return(user, nil)

		err = svc.Login(context.Background(), login, wrongPassword)

		assert.ErrorIs(t, err, model.ErrInvalidCredentials)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockUserRepository(ctrl)
		svc := NewUserService(mockRepo)

		login := "testuser"
		password := "password123"

		mockRepo.EXPECT().
			GetUserByLogin(gomock.Any(), login).
			Return(nil, assert.AnError)

		err := svc.Login(context.Background(), login, password)

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})
}
