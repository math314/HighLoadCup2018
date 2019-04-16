package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
)

type X struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func indexHandler(c echo.Context) error {
	x := X{"1", "user name"}
	return c.JSON(http.StatusOK, x)
}

func main() {
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "request:\"${method} ${uri}\" status:${status} latency:${latency} (${latency_human}) bytes:${bytes_out}\n",
	}))

	e.GET("/", indexHandler)
	e.Start(":8080")
}
