package services

import (
	"context"
	"github.com/KouT127/attendance-management/domain/models"
	"github.com/KouT127/attendance-management/infrastructure/sqlstore"
	"github.com/Songmu/flextime"
	"golang.org/x/xerrors"
)

type AttendanceService interface {
	GetAttendances(params models.GetAttendancesParameters) (*models.GetAttendancesResults, error)
	CreateOrUpdateAttendance(ctx context.Context, attendanceTime *models.AttendanceTime, userID string) (*models.Attendance, error)
}

type attendanceService struct {
	store sqlstore.SqlStore
}

func NewAttendanceService(ss sqlstore.SqlStore) AttendanceService {
	return &attendanceService{
		store: ss,
	}
}

func (s *attendanceService) GetAttendances(params models.GetAttendancesParameters) (*models.GetAttendancesResults, error) {
	ctx := context.Background()
	maxCnt, err := s.store.GetAttendancesCount(ctx, params.UserID)
	if err != nil {
		return nil, err
	}
	attendances, err := s.store.GetAttendances(ctx, &params)
	if err != nil {
		return nil, err
	}

	res := models.GetAttendancesResults{
		MaxCnt:      maxCnt,
		Attendances: attendances,
	}
	return &res, nil
}

func (s *attendanceService) CreateOrUpdateAttendance(ctx context.Context, attendanceTime *models.AttendanceTime, userID string) (*models.Attendance, error) {
	if userID == "" {
		return nil, xerrors.New("userID is empty")
	}
	if attendanceTime == nil {
		return nil, xerrors.New("attendance time is empty")
	}

	ctx, err := s.store.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer s.store.Close(ctx)

	attendance, err := s.store.GetLatestAttendance(ctx, userID)
	if err != nil {
		return nil, err
	}

	if attendance == nil {
		attendance = &models.Attendance{}
		attendance.UserID = userID
		attendance.ClockedIn = attendanceTime
		attendance.CreatedAt = flextime.Now()
		attendance.UpdatedAt = flextime.Now()
		if err = s.store.CreateAttendance(ctx, attendance); err != nil {
			return nil, err
		}
		attendanceTime.AttendanceKindID = uint8(models.AttendanceKindClockIn)
	} else {
		if err = s.store.UpdateOldAttendanceTime(ctx, attendance.ID, uint8(models.AttendanceKindClockOut)); err != nil {
			return nil, err
		}
		attendance.ClockedOut = attendanceTime
		attendanceTime.AttendanceKindID = uint8(models.AttendanceKindClockOut)
	}
	attendanceTime.PushedAt = flextime.Now()
	attendanceTime.CreatedAt = flextime.Now()
	attendanceTime.UpdatedAt = flextime.Now()
	attendanceTime.AttendanceID = attendance.ID

	if err = s.store.CreateAttendanceTime(ctx, attendanceTime); err != nil {
		return nil, err
	}

	if err = s.store.Commit(ctx); err != nil {
		return nil, err
	}
	return attendance, nil
}