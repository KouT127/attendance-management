package usecases

import (
	"github.com/KouT127/attendance-management/backend/database"
	. "github.com/KouT127/attendance-management/backend/models"
	. "github.com/KouT127/attendance-management/backend/repositories"
	. "github.com/KouT127/attendance-management/backend/serializers"
	. "github.com/KouT127/attendance-management/backend/validators"
)

func NewAttendanceInteractor(ar AttendanceRepository) *attendanceInteractor {
	return &attendanceInteractor{
		ar: ar,
	}
}

type AttendanceInteractor interface {
	ViewAttendances(pagination *Pagination, attendance *Attendance) (*AttendancesResponse, error)
	CreateAttendance(query *Attendance, time *AttendanceTime) (*AttendanceResponse, error)
}

type attendanceInteractor struct {
	ar AttendanceRepository
}

func (i *attendanceInteractor) ViewAttendances(pagination *Pagination, attendance *Attendance) (*AttendancesResponse, error) {
	eng := database.NewDB()
	maxCnt, err := i.ar.FetchAttendancesCount(eng, attendance)
	if err != nil {
		return nil, err
	}

	attendances := make([]*Attendance, 0)
	attendances, err = i.ar.FetchAttendances(eng, attendance, pagination)
	if err != nil {
		return nil, err
	}

	responses := make([]*AttendanceResponse, 0)

	for _, attendance := range attendances {
		res := &AttendanceResponse{}
		res.Build(attendance)
		responses = append(responses, res)
	}

	res := new(AttendancesResponse)
	res.HasNext = pagination.HasNext(maxCnt)
	res.IsSuccessful = true
	res.Attendances = responses
	return res, nil
}

func (i *attendanceInteractor) CreateAttendance(query *Attendance, time *AttendanceTime) (*AttendanceResponse, error) {
	sess := i.ar.NewSession(database.NewDB())
	defer i.ar.Close(sess)
	if err := i.ar.Begin(sess); err != nil {
		return nil, err
	}

	attendance, err := i.ar.FetchLatestAttendance(sess, query)
	if err != nil {
		return nil, err
	}

	if err := i.ar.CreateAttendanceTime(sess, time); err != nil {
		return nil, err
	}

	if attendance == nil {
		attendance = &Attendance{
			UserId:    query.UserId,
			ClockedIn: *time,
		}
		if _, err := i.ar.CreateAttendance(sess, attendance); err != nil {
			return nil, err
		}
	} else {
		attendance = &Attendance{
			Id:         attendance.Id,
			UserId:     attendance.UserId,
			ClockedIn:  attendance.ClockedIn,
			ClockedOut: *time,
		}
		if _, err := i.ar.UpdateAttendance(sess, attendance); err != nil {
			return nil, err
		}
	}

	res := new(AttendanceResponse)
	res.Build(attendance)

	if err := i.ar.Commit(sess); err != nil {
		return nil, err
	}
	return res, nil
}
