package shttp

import (
	"context"
	"errors"
	"example.com/imageProc/internal/domain"
	"example.com/imageProc/internal/domain/service"
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

var (
	ErrInvalidAspectRatio = errors.New("invalid aspect ratio")
	ErrInvalidWidth       = errors.New("invalid width")
	ErrInvalidHeight      = errors.New("invalid height")
)

func (h httpService) GetImage(c echo.Context) error {
	imgName := c.Param("imgName")

	queryPrms := c.QueryParams()
	ar := queryPrms.Get("ar")
	width := queryPrms.Get("width")
	height := queryPrms.Get("height")
	tenantCode := queryPrms.Get("tenant-code")
	orgCode := queryPrms.Get("org-code")

	getImgOpts, err := prepareGetImageOpts(width, height, ar)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if tenantCode == "" || orgCode == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "tenant-code and org-code are required")
	}
	tenantOpts := domain.TenantOpts{
		TenantCode: queryPrms.Get("tenant-code"),
		OrgCode:    queryPrms.Get("org-code"),
	}

	// get image type from accepts header
	acceptHeader := c.Request().Header.Get("accept")
	_imgType := distinguishImageType(strings.Split(acceptHeader, ","))

	opts := domainsvc.NewServiceGetImageOpts()
	opts = getImgOpts.SetFormat(_imgType).SetTenantOpts(tenantOpts).SetName(imgName)

	image, err := h.imageSvc.GetImage(context.Background(), opts)
	if err != nil {
		if errors.Is(err, domainsvc.ErrNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "image not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching image").SetInternal(err)
	}
	return c.Blob(http.StatusOK, contentTypeString(_imgType), image)
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
		return domain.ImageType_AUTO
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

func prepareGetImageOpts(width, height, ar string) (domainsvc.GetImageOpts, error) {
	svcGetImgOpts := domainsvc.NewServiceGetImageOpts()
	switch {
	case width != "" && height != "" && ar != "":
		validAr, err := domain.ParseAspectRatio(ar)
		if err != nil {
			return svcGetImgOpts, ErrInvalidAspectRatio
		}
		validWidth, err := strconv.Atoi(width)
		if err != nil {
			return svcGetImgOpts, ErrInvalidWidth
		}
		validHeight, err := strconv.Atoi(height)
		if err != nil {
			return svcGetImgOpts, ErrInvalidHeight
		}
		svcGetImgOpts = svcGetImgOpts.SetWidth(validWidth).SetHeight(validHeight).SetAr(validAr)
	case width != "" && height == "" && ar == "":
		validWidth, err := strconv.Atoi(width)
		if err != nil {
			return svcGetImgOpts, ErrInvalidWidth
		}
		svcGetImgOpts = svcGetImgOpts.SetWidth(validWidth)
	case width == "" && height != "" && ar == "":
		validHeight, err := strconv.Atoi(height)
		if err != nil {
			return svcGetImgOpts, ErrInvalidHeight
		}
		svcGetImgOpts = svcGetImgOpts.SetHeight(validHeight)
	case width == "" && height == "" && ar != "":
		validAr, err := domain.ParseAspectRatio(ar)
		if err != nil {
			return svcGetImgOpts, ErrInvalidAspectRatio
		}
		svcGetImgOpts = svcGetImgOpts.SetAr(validAr)
	case width != "" && height != "" && ar == "":
		validWidth, err := strconv.Atoi(width)
		if err != nil {
			return svcGetImgOpts, ErrInvalidWidth
		}
		validHeight, err := strconv.Atoi(height)
		if err != nil {
			return svcGetImgOpts, ErrInvalidHeight
		}
		svcGetImgOpts = svcGetImgOpts.SetWidth(validWidth).SetHeight(validHeight)
	case width != "" && height == "" && ar != "":
		validWidth, err := strconv.Atoi(width)
		if err != nil {
			return svcGetImgOpts, ErrInvalidWidth
		}
		validAr, err := domain.ParseAspectRatio(ar)
		if err != nil {
			return svcGetImgOpts, ErrInvalidAspectRatio
		}
		svcGetImgOpts = svcGetImgOpts.SetWidth(validWidth).SetAr(validAr)
	case width == "" && height != "" && ar != "":
		validHeight, err := strconv.Atoi(height)
		if err != nil {
			return svcGetImgOpts, ErrInvalidHeight
		}
		validAr, err := domain.ParseAspectRatio(ar)
		if err != nil {
			return svcGetImgOpts, ErrInvalidAspectRatio
		}
		svcGetImgOpts = svcGetImgOpts.SetHeight(validHeight).SetAr(validAr)
	}
	return svcGetImgOpts, nil
}
