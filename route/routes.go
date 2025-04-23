package route

import (
	controller "example.com/Product_RoadMap/controller"
	"example.com/Product_RoadMap/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {

	// Public routes
	router.POST("/register", controller.RegisterUser)
	router.POST("/login", controller.LoginUser)

	// Protected routes
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		//Users
		protected.POST("/createUser", controller.CreateUsers)
		protected.GET("/getUser/:id", controller.GetUsersByID)
		protected.PUT("/updateUser/:id", controller.UpdateUsers)
		protected.DELETE("/deleteUser/:id", controller.DeleteUsers)
		protected.GET("/getAllUsers", controller.GetAllUsers)

		//Features
		protected.POST("/createFeature", controller.CreateFeatures)
		protected.GET("/getFeature/:id", controller.GetFeatureByID)
		protected.DELETE("/deleteFeature/:id", controller.DeletFeatureById)
		protected.PUT("/updateFeature/:id", controller.UpdateFeatureById)
		protected.GET("/getAllFeatures", controller.GetAllFeatures)

		//Feature Requests
		protected.POST("/createFeatureRequest", controller.CreateFeatureRequest)
		protected.GET("/getFeatureRequest/:id", controller.GetFeatureRequestByID)
		protected.PUT("/updateFeatureRequest/:id", controller.UpdateFeatureRequestByID)
		protected.DELETE("/deleteFeatureRequest/:id", controller.DeleteFeatureRequestByID)
		protected.GET("/getAllFeatureRequests", controller.GetAllFeatureRequests)

		//router.POST("/test", controller.TestProductBoardAPI)
		protected.POST("/productBoard", controller.CreateProductBoardFeature)
	}
}
