package main

import (
	"errors"
	"net/http"
	"strings"

	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Uuid() string {
	uuidWithHyphen := uuid.New()
	return strings.Replace(uuidWithHyphen.String(), "-", "", -1)
}

func GetExtension(filename string) string {
	return filepath.Ext(filename)
}

func EnsureLoggedIn(c *gin.Context) (string, error) {
	session := sessions.Default(c)

	username := session.Get("authed_user")
	if username == nil {
		c.Redirect(http.StatusFound, "/login")
		return "", errors.New("Unauthenticated")
	}

	return username.(string), nil
}
