package services

import (
	"context"
	"github.com/KouT127/attendance-management/domain/models"
	"github.com/KouT127/attendance-management/infrastructure/sqlstore"
	"golang.org/x/xerrors"
)

type UserService interface {
	GetOrCreateUser(params models.GetOrCreateUserParams) (*models.GetOrCreateUserResults, error)
	UpdateUser(user *models.User) error
}

type userService struct {
	store sqlstore.SqlStore
}

func NewUserService(ss sqlstore.SqlStore) *userService {
	return &userService{
		store: ss,
	}
}

func (s *userService) GetOrCreateUser(params models.GetOrCreateUserParams) (*models.GetOrCreateUserResults, error) {
	var (
		user       *models.User
		attendance *models.Attendance
		err        error
	)
	if params.UserID == "" {
		return nil, xerrors.New("user id is empty")
	}

	_, err = s.store.InTransaction(context.Background(), func(ctx context.Context) (interface{}, error) {
		user, err = s.store.GetUser(ctx, params.UserID)
		if err != nil {
			return nil, err
		}

		if user.ID == "" {
			user.ID = params.UserID
			if err = s.store.CreateUser(ctx, user); err != nil {
				return nil, err
			}
		}

		if attendance, err = s.store.GetLatestAttendance(context.Background(), params.UserID); err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	res := models.GetOrCreateUserResults{
		User:             user,
		LatestAttendance: attendance,
	}
	return &res, nil
}

func (s *userService) UpdateUser(user *models.User) error {
	if user == nil {
		return xerrors.New("user pointer is empty")
	}
	_, err := s.store.InTransaction(context.Background(), func(ctx context.Context) (interface{}, error) {
		return nil, s.store.UpdateUser(ctx, user)
	})

	if err != nil {
		return err
	}
	return nil
}