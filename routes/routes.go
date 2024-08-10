package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/prajjwal-w/golang-choicetech/controller"
)

// routes
func Routes(r *gin.Engine) {
	r.POST("/upload", controller.UploadFile())
	r.GET("/viewdata/*email", controller.ViewData())
	r.PUT("/update", controller.UpdateData())

}
