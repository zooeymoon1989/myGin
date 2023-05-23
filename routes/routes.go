package routes

import (
	"github.com/gin-gonic/gin"
	"myGin/handlers"
)

func Routes() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", handlers.Pong)
	return r
}
