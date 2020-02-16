package repositories

import (
	. "github.com/KouT127/attendance-management/database"
	"github.com/KouT127/attendance-management/models"
	"github.com/go-xorm/xorm"
	"time"
)

type AttendanceKind int

const (
	AttendanceKindNone AttendanceKind = iota
	AttendanceKindClockIn
	AttendanceKindClockOut
)

func (k AttendanceKind) String() string {
	switch k {
	case AttendanceKindClockIn:
		return "出勤"
	case AttendanceKindClockOut:
		return "退勤"
	}
	return "不明"
}

type Attendance struct {
	Id        int64
	UserId    string
	CreatedAt time.Time `xorm:"created_at"`
	UpdatedAt time.Time `xorm:"updated_at"`
}

func (a *Attendance) build(cit *AttendanceTime, cot *AttendanceTime) models.Attendance {
	return models.Attendance{
		Id:         a.Id,
		UserId:     a.UserId,
		ClockedIn:  cit.build(),
		ClockedOut: cot.build(),
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}

type AttendanceTime struct {
	Id               int64
	PushedAt         time.Time
	Remark           string
	AttendanceId     int64
	AttendanceKindId int64
	CreatedAt        time.Time `xorm:"created_at"`
	UpdatedAt        time.Time `xorm:"updated_at"`
}

func NewTime(at *models.AttendanceTime) *AttendanceTime {
	t := new(AttendanceTime)
	t.Id = at.Id
	t.Remark = at.Remark
	t.AttendanceId = at.AttendanceId
	t.AttendanceKindId = at.AttendanceKindId
	t.CreatedAt = at.CreatedAt
	t.UpdatedAt = at.UpdatedAt
	t.PushedAt = at.PushedAt
	return t
}

func (t AttendanceTime) build() *models.AttendanceTime {
	return &models.AttendanceTime{
		Id:        t.Id,
		Remark:    t.Remark,
		PushedAt:  t.PushedAt,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

type AttendanceDetail struct {
	Attendance     `xorm:"extends"`
	ClockedInTime  *AttendanceTime `xorm:"extends"`
	ClockedOutTime *AttendanceTime `xorm:"extends"`
}

func (d AttendanceDetail) build() *models.Attendance {
	var (
		in  *models.AttendanceTime
		out *models.AttendanceTime
	)
	a := d.Attendance
	if d.ClockedInTime.Id != 0 {
		in = d.ClockedInTime.build()
	}
	if d.ClockedOutTime.Id != 0 {
		out = d.ClockedOutTime.build()
	}

	attendance := &models.Attendance{
		Id:         a.Id,
		UserId:     a.UserId,
		ClockedIn:  in,
		ClockedOut: out,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
	return attendance
}

func NewAttendanceRepository() *attendanceRepository {
	return &attendanceRepository{}
}

type AttendanceRepository interface {
	FetchAttendancesCount(eng *xorm.Engine, a *models.Attendance) (int64, error)
	FetchAttendances(eng *xorm.Engine, a *models.Attendance, p *Paginator) ([]*models.Attendance, error)
	FetchLatestAttendance(eng *xorm.Engine, query *models.Attendance) (*models.Attendance, error)
	CreateAttendance(sess *xorm.Session, a *models.Attendance) error
	CreateAttendanceTime(sess *xorm.Session, t *models.AttendanceTime) error
	Transaction
}

type attendanceRepository struct {
	transaction
}

func (r *attendanceRepository) FetchAttendancesCount(eng *xorm.Engine, a *models.Attendance) (int64, error) {
	attendance := &Attendance{}
	attendance.Id = a.Id
	return eng.Table(AttendanceTable).Count(attendance)
}

func (r *attendanceRepository) FetchLatestAttendance(eng *xorm.Engine, a *models.Attendance) (*models.Attendance, error) {
	attendance := &AttendanceDetail{}
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 59, time.Local)

	has, err := eng.Select("attendances.*, clocked_in_time.*, clocked_out_time.*").
		Table(AttendanceTable).
		Join("left", "attendances_time clocked_in_time", "attendances.id = clocked_in_time.attendance_id").
		Join("left", "attendances_time clocked_out_time", "attendances.id = clocked_out_time.attendance_id").
		Where("attendances.user_id = ?", a.UserId).
		Where("attendances.created_at Between ? and ? ", start, end).
		Limit(1).
		OrderBy("-attendances.id").
		Get(attendance)

	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	return attendance.build(), nil
}

func (r *attendanceRepository) FetchAttendances(eng *xorm.Engine, a *models.Attendance, p *Paginator) ([]*models.Attendance, error) {
	attendances := make([]*models.Attendance, 0)
	page := p.CalculatePage()
	err := eng.
		Select("attendances.*, clocked_in_time.*, clocked_out_time.*").
		Table(AttendanceTable).
		Join("left", "attendances_time clocked_in_time", "attendances.id = clocked_in_time.attendance_id").
		Join("left", "attendances_time clocked_out_time", "attendances.id = clocked_out_time.attendance_id").
		Where("clocked_in_time.attendance_kind_id = ?", AttendanceKindClockIn).
		Where("clocked_out_time.attendance_kind_id = ?", AttendanceKindClockOut).
		Limit(int(p.Limit), int(page)).
		OrderBy("-attendances.id").
		Iterate(&AttendanceDetail{}, func(idx int, bean interface{}) error {
			d := bean.(*AttendanceDetail)
			a := d.build()
			attendances = append(attendances, a)
			return nil
		})
	return attendances, err
}

func (r *attendanceRepository) CreateAttendance(sess *xorm.Session, a *models.Attendance) error {
	attendance := Attendance{}
	attendance.UserId = a.UserId
	attendance.CreatedAt = time.Now()
	attendance.UpdatedAt = time.Now()
	if _, err := sess.Table(AttendanceTable).Insert(&attendance); err != nil {
		return err
	}
	a.Id = attendance.Id
	return nil
}

func (r *attendanceRepository) CreateAttendanceTime(sess *xorm.Session, t *models.AttendanceTime) error {
	at := NewTime(t)
	if _, err := sess.Table(AttendanceTimeTable).Insert(at); err != nil {
		return err
	}
	t.Id = at.Id
	return nil
}
