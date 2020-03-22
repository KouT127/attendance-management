package v1

import (
	"github.com/KouT127/attendance-management/handlers/middlewares"
	"github.com/KouT127/attendance-management/handlers/v1/attendance"
	. "github.com/gin-gonic/gin"
)

func AttendancesRouter(v1 *RouterGroup) {
	funcs := []HandlerFunc{
		middlewares.AuthRequired(),
	}

	attendances := v1.Group("/attendances", funcs...)
	attendances.GET("", attendance.ListHandler)
	attendances.POST("", attendance.CreateHandler)
	attendances.GET("monthly", attendance.MonthlyHandler)
}
