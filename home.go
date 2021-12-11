package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HomeGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "Main website",
		})
	}
}
