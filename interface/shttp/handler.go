package shttp

import (
	"context"
	"errors"
	"example.com/imageProc/domain"
	domainsvc "example.com/imageProc/domain/service"
	"github.com/labstack/echo/v4"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type HttpServiceInterface interface {
	GetImage(c echo.Context) error
	UploadImage(c echo.Context) error
}

type httpService struct {
	imageSvc domainsvc.ImageServiceInterface
}

func (h httpService) GetImage(c echo.Context) error {
	queryPrms := c.QueryParams()

	imgName := c.Param("imgName")

	tenantCode := queryPrms.Get("tenant-code")
	orgCode := queryPrms.Get("org-code")

	if tenantCode == "" || orgCode == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "tenant-code and org-code are required")
	}
	tenantOpts := domain.TenantOpts{
		TenantCode: queryPrms.Get("tenant-code"),
		OrgCode:    queryPrms.Get("org-code"),
	}

	// get image aspect ratio and width from query params
	ar, err := domain.ParseAspectRatio(queryPrms.Get("ar"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid aspect ratio")
	}
	width, err := strconv.Atoi(queryPrms.Get("width"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid width")
	}
	// get image type from accepts header
	acceptHeader := c.Request().Header.Get("accept")
	imgType := distinguishImageType(strings.Split(acceptHeader, ","))

	image, err := h.imageSvc.GetImage(context.Background(), domainsvc.GetImageOpts{
		TenantOpts: tenantOpts,
		Name:       imgName,
		Width:      width,
		Ar:         ar,
		Type:       imgType,
	})
	if err != nil {
		if errors.Is(err, domainsvc.ErrNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching image").SetInternal(err)
	}
	return c.Blob(http.StatusOK, contentTypeString(imgType), image)
}

func (h httpService) UploadImage(c echo.Context) error {
	queryPrms := c.QueryParams()

	tenantCode := queryPrms.Get("tenant-code")
	orgCode := queryPrms.Get("org-code")

	if tenantCode == "" || orgCode == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "tenant-code and org-code are required")
	}
	tenantOpts := domain.TenantOpts{
		TenantCode: queryPrms.Get("tenant-code"),
		OrgCode:    queryPrms.Get("org-code"),
	}

	file, err := c.FormFile("img")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to get the file").SetInternal(err)
	}
	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open the file").SetInternal(err)
	}
	defer src.Close()
	// TODO: use io.Copy() instead to avoid loading entire file to memory
	img := make([]byte, file.Size)

	_, err = src.Read(img)
	if err != nil {
		return err
	}

	imgName, err := h.imageSvc.Upload(context.Background(), img, tenantOpts)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to upload the file")
	}
	err = c.JSON(http.StatusOK, map[string]string{"imgName": imgName})
	if err != nil {
		return err
	}
	return nil
}

func NewHttpService(imgSvc domainsvc.ImageServiceInterface) HttpServiceInterface {
	return httpService{
		imgSvc,
	}
}

func distinguishImageType(acceptHeader []string) domain.ImageType {
	if slices.Contains(acceptHeader, "image/avif") {
		return domain.ImageType_AVIF
	} else if slices.Contains(acceptHeader, "image/webp") {
		return domain.ImageType_WEBP
	} else {
		return domain.ImageType_JPEG
	}
}

func contentTypeString(imgType domain.ImageType) string {
	switch imgType {
	case domain.ImageType_AVIF:
		return "image/avif"
	case domain.ImageType_WEBP:
		return "image/webp"
	case domain.ImageType_JPEG:
		return "image/jpeg"
	case domain.ImageType_PNG:
		return "image/png"
	default:
		return ""
	}
}
