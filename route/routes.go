package route

import (
	controller "example.com/ringover_kb/controller"
	"example.com/ringover_kb/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {

	router.Use(middleware.CORSMiddleware())

	// Public routes
	router.GET("/poc", controller.GetCleanDataFromTranscriptDump)

}
