package handlers

import (
	"net/http"

	. "github.com/eburlingame/fstop/resources"

	"github.com/gin-gonic/gin"
)

func AdminGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {

		c.HTML(http.StatusOK, "admin.html", gin.H{
			"title": "Dashboard",
		})
	}
}
