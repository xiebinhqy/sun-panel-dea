package system

import (
	"sun-panel/api/api_v1"

	"github.com/gin-gonic/gin"
)

func InitUpdate(router *gin.RouterGroup) {
	update := api_v1.ApiGroupApp.ApiSystem.UpdateApi
	{
		router.POST("update/check", update.CheckUpdate)
		router.POST("update/perform", update.PerformUpdate)
		router.POST("update/status", update.GetUpdateStatus)
	}
}
