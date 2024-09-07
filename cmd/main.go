package main

import (
	"example.com/imageProc/interface/shttp"
	appsvc "example.com/imageProc/internal/app/service"
	"example.com/imageProc/internal/domain/service"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	baseDir := os.Getenv("StorageDir")
	localImageStorageSvc := appsvc.NewLocalImageStorageService(baseDir)
	vipsImageProcessorSvc := appsvc.NewVipsImageProcessorService()
	imgSvc := domainsvc.NewImageService(localImageStorageSvc, vipsImageProcessorSvc)

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
