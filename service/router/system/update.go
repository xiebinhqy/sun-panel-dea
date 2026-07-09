package system

import (
	"sun-panel/api/api_v1"

	"github.com/gin-gonic/gin"
)

func InitUpdateRouter(router *gin.RouterGroup) {
	api := api_v1.ApiGroupApp.ApiSystem.UpdateApi
	{
		router.POST("system/checkUpdate", api.CheckUpdate)
		router.POST("system/performUpdate", api.PerformUpdate)
	}
}
