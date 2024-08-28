package main

import (
	domainsvc "example.com/imageProc/domain/service"
	"example.com/imageProc/infra/persist"
	"example.com/imageProc/interface/shttp"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	baseDir := "/home/errfunm/Projects/Go/imageProc/static-files"
	repo := persist.NewVipsImageRepo(baseDir)
	imgSvc := domainsvc.NewImageService(repo)
	httpSvc := shttp.NewHttpService(imgSvc)

	vips.Startup(nil)
	defer vips.Shutdown()

	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/:imgName", func(c echo.Context) error {
		return httpSvc.GetImage(c)
	})
	e.POST("/upload", func(c echo.Context) error {
		return httpSvc.UploadImage(c)
	})

	e.Logger.Fatal(e.Start(":2380"))
}
