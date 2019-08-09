package server

import (
	"fmt"
	"github.com/kohkimakimoto/cofu/ext/structs"
	"github.com/labstack/echo"
	"net/http"
)

func ErrorHandler(err error, c echo.Context) {
	e := c.Echo()

	e.Logger.Error(fmt.Sprintf("%+v", err))

	var statusCode int
	var message string

	if httperr, ok := err.(*echo.HTTPError); ok {
		statusCode = httperr.Code
		if msg, ok := httperr.Message.(string); ok {
			message = msg
		} else {
			message = http.StatusText(statusCode)
		}
	} else {
		statusCode = http.StatusInternalServerError

		message = err.Error()
		if message == "" {
			message = http.StatusText(statusCode)
		}
	}

	if err2 := c.JSON(statusCode, &structs.ErrorResponse{
		Status: statusCode,
		Error:  message,
	}); err2 != nil {
		e.Logger.Error(fmt.Sprintf("%+v", err2))
	}
}

func errorHandler(srv *Server) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		e := c.Echo()
		if c.Response().Committed {
			goto ERROR
		}

		ErrorHandler(err, c)
	ERROR:
		e.Logger.Error(err)
	}
}
