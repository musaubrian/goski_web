package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/musaubrian/goski_web/views"
)

const port = "9999"

type AsciiResponse struct {
	Ascii string `json:"ascii"`
	Error string `json:"error"`
}

func main() {
	app := echo.New()
	app.Static("/static", "static")
	app.GET("/", func(c echo.Context) error {
		return render(c, views.Index())
	})
	app.POST("/to-ascii", func(c echo.Context) error {
		response := AsciiResponse{}
		image, err := c.FormFile("image")
		if err != nil {
			slog.Error(fmt.Sprintf("error getting file: %s", err.Error()))
			response.Error = "Failed to process image"
			return c.JSON(http.StatusBadRequest, response)
		}
		viewportWidth, err := strconv.Atoi(c.FormValue("viewportWidth"))
		if err != nil {
			slog.Error(err.Error())
			response.Error = "Something went wrong"
			return c.JSON(http.StatusBadRequest, response)
		}

		viewportHeight, err := strconv.Atoi(c.FormValue("viewportHeight"))
		if err != nil {
			slog.Error(err.Error())
			response.Error = "Something went wrong"
			return c.JSON(http.StatusBadRequest, response)
		}

		ascii, err := convertToASCII(image, viewportWidth, viewportHeight)
		if err != nil {
			response.Error = err.Error()
			return c.JSON(http.StatusInternalServerError, response)
		}

		response.Ascii = ascii

		return c.JSON(http.StatusOK, response)
	})
	app.Logger.Fatal(app.Start(fmt.Sprintf(":%s", port)))
}

func render(c echo.Context, component templ.Component) error {
	return component.Render(c.Request().Context(), c.Response())
}

func convertToASCII(src *multipart.FileHeader, vw, vh int) (string, error) {
	asciiChars := " .,:;i1tfLCG08@"
	font := strings.Split(asciiChars, "")

	slog.Info(src.Filename)

	f, err := src.Open()
	if err != nil {
		return "", fmt.Errorf("Failed to process image")
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", fmt.Errorf("Failed to decode image")
	}

	_, _ = vw, vh
	// nw, nh := autoScale(img.Bounds(), vw, vh)
	// img = resize.Resize(uint(nw), uint(nh), img, resize.MitchellNetravali)
	result := grayScaledAscii(img, font)

	return result, nil
}

func grayScaledAscii(img image.Image, font []string) string {
	var asciiImgChars strings.Builder

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			char := int(c.Y) * (len(font) - 1) / 255
			asciiImgChars.WriteString(font[char])

		}
		asciiImgChars.WriteString("\n")
	}
	return asciiImgChars.String()
}

func autoScale(size image.Rectangle, vw, vh int) (int, int) {
	targetHeight := int(float64(vh) * 0.8)

	ratio := float64(targetHeight) / float64(size.Dy())
	newWidth := int(float64(size.Dx()) * ratio * 1.6)
	if newWidth > vw {
		ratio = float64(vw) / float64(size.Dx())
		targetHeight = int(float64(size.Dy()) * ratio)
		newWidth = vw
	}

	return newWidth, targetHeight
}
