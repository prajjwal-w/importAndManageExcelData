package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/prajjwal-w/golang-choicetech/controller"
	"github.com/prajjwal-w/golang-choicetech/middleware"
)

// routes
func Routes(r *gin.Engine) {
	r.GET("/viewdata/*email", controller.ViewData())

	r.POST("/upload", middleware.Auth(),controller.UploadFile())
	r.PUT("/update", middleware.Auth(), controller.UpdateData())

}
