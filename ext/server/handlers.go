package server

import (
	"github.com/labstack/echo"
	"net/http"
)

func IndexHandler(c echo.Context) error {
	return c.String(http.StatusOK, "hello")
}
