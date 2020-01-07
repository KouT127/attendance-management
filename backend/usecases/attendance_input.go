package usecases

import (
	. "github.com/KouT127/attendance-management/backend/models"
	validation "github.com/go-ozzo/ozzo-validation/v3"
	"time"
)

type AttendanceInput struct {
	Remark string
}

func (i AttendanceInput) Validate() error {
	return validation.ValidateStruct(&i,
		validation.Field(&i.Remark, validation.Length(0, 1000)),
	)
}

func (i AttendanceInput) BuildAttendanceTime() *AttendanceTime {
	t := new(AttendanceTime)
	t.Remark = i.Remark
	t.PushedAt = time.Now()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return t
}
