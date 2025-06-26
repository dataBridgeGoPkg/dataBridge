package route

import (
	controller "example.com/Product_RoadMap/controller"
	"example.com/Product_RoadMap/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {

	router.Use(middleware.CORSMiddleware())

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

		protected.GET("/getAllfeaturesWithName", controller.GetAllFeaturesWithAssginness) //
		//Feature Assignees
		protected.POST("/createFeatureAssignees", controller.CreateFeatureAssignee)
		protected.POST("/AddAssigneeToFeature", controller.AddUserToAFeature)
		protected.DELETE("/DeleteAssigneeToFeature", controller.DeleteAssigneeFromAFeature)

		//Feature Requests
		protected.POST("/createFeatureRequest", controller.CreateFeatureRequest)
		protected.GET("/getFeatureRequest/:id", controller.GetFeatureRequestByID)
		protected.PUT("/updateFeatureRequest/:id", controller.UpdateFeatureRequestByID)
		protected.DELETE("/deleteFeatureRequest/:id", controller.DeleteFeatureRequestByID)
		protected.GET("/getAllFeatureRequests", controller.GetAllFeatureRequests)

		//Document
		protected.POST("/createDocument", controller.CreateDocument)
		protected.GET("/getDocument/:id", controller.GetDocumentByID)
		protected.PUT("/updateDocument/:id", controller.UpdateDocumentById)
		protected.DELETE("/deleteDocument/:id", controller.DeleteDocumentById)
		protected.GET("/getAllDocuments", controller.GetAllDocuments)

		//router.POST("/test", controller.TestProductBoardAPI)
		protected.POST("/productBoard", controller.CreateProductBoardFeature)
		protected.POST("/jira", controller.CreateJiraIssue)
		protected.PUT("/jira/:id", controller.UpdateJiraIssue)
		protected.GET("/jira/:id", controller.GetJiraIssueByID)

		// Release Notes
		protected.POST("/createReleaseNote", controller.CreateReleaseNote)
		protected.GET("/getReleaseNote/:id", controller.GetReleaseNoteByID)
		protected.PUT("/updateReleaseNote/:id", controller.UpdateReleaseNoteByID)
		protected.DELETE("/deleteReleaseNote/:id", controller.DeleteReleaseNoteByID)
		protected.GET("/getAllReleaseNotes", controller.GetAllReleaseNotes)
		protected.POST("/uploadReleaseNoteImage", controller.UploadReleaseNoteImage)

	}
}
