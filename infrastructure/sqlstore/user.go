package sqlstore

import (
	"context"
	"github.com/KouT127/attendance-management/domain/models"
	"github.com/KouT127/attendance-management/utilities/logger"
	"golang.org/x/xerrors"
)

type User interface {
	GetUser(ctx context.Context, userID string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
}

func (sqlStore) GetUser(ctx context.Context, userID string) (*models.User, error) {
	sess, err := getDBSession(ctx)
	if err != nil {
		return nil, err
	}

	user := &models.User{}
	_, err = sess.Where("id = ?", userID).Get(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (sqlStore) CreateUser(ctx context.Context, user *models.User) error {
	sess, err := getDBSession(ctx)
	if err != nil {
		return err
	}

	if _, err := sess.Insert(user); err != nil {
		return err
	}
	return nil
}

func (sqlStore) UpdateUser(ctx context.Context, user *models.User) error {
	sess, err := getDBSession(ctx)
	if err != nil {
		return err
	}

	has, err := sess.Where("id = ?", user.ID).Exist(&models.User{})
	if err != nil {
		return err
	}
	if !has {
		return xerrors.New("user is not exists")
	}

	if _, err := sess.Where("id = ?", user.ID).Update(user); err != nil {
		return err
	}
	logger.NewInfo("updated user_id: " + user.ID)
	return nil
}
