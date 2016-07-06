package main

import (
	"github.com/gin-gonic/gin"
	"github.com/vincer/libhdplatinum"
	"github.com/Azure/azure-sdk-for-go/core/http"
	"github.com/urfave/cli"
	"errors"
	"os"
	"strings"
	"net"
)

const MaxShadeHeight = 255

type ShadeView struct {
	Id string `json:"id"`
	Name string `json:"name"`
	RoomId string `json:"roomId"`
	Height int `json:"height"`
}

func ShadeViewFromShade(s libhdplatinum.Shade) ShadeView {
	return ShadeView{Id: s.Id(), Name: s.Name(), RoomId: s.RoomId(), Height: s.Height() * 100 / MaxShadeHeight}
}

type Response struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

func validate(ip string, port int, c *cli.Context) *cli.ExitError {
	if strings.TrimSpace(ip) == "" {
		return cli.NewExitError("IP address is required. `platypus -h` for usage info.", 1)
	}
	if net.ParseIP(ip) == nil {
		return cli.NewExitError("Invalid IP address", 1)
	}
	if port < 0 || port > 65535 {
		return cli.NewExitError("Invalid port", 2)
	}
	return nil
}
func main() {
	app := cli.NewApp()
	app.Name = "hdp"
	app.Usage = "Hunter Douglas Platinum REST API"
	app.Version = "0.0.1"
	app.HideHelp = true
	app.EnableBashCompletion = true

	flags := []cli.Flag{
		cli.StringFlag{
			Name: "hdp-ip",
			Usage: "ip address of the Hunter Douglas Platinum Gateway. Required.",
		},
		cli.IntFlag{
			Name: "hdp-port",
			Value: 522,
			Usage: "port of the Hunter Douglas Platinum Gateway.",
		},
		cli.BoolFlag{
			Name: "help, h",
			Usage: "Show usage help.",
		},
	}

	app.Flags = flags


	app.Action = func(c *cli.Context) error {
		ip := c.String("hdp-ip")
		port := c.Int("hdp-port")
		exitError := validate(ip, port, c)
		if exitError != nil {
			return exitError
		}
		r := gin.Default()
		r.GET("/shades", GetShades(ip, port))
		r.GET("/shades/:id", GetShade(ip, port))
		r.PUT("/shades/:id", UpdateShade(ip, port))
		r.PUT("/shades/:id/height", UpdateShade(ip, port))

		r.Run() // listen and server on 0.0.0.0:8080
		return nil
	}

	app.Run(os.Args)


}

func getShadeViews(ip string, port int) []ShadeView {
	output := []ShadeView{}
	for _, s := range libhdplatinum.GetShades(ip, port) {
		output = append(output, ShadeViewFromShade(s))
	}

	return output
}

func findShade(id string, ip string, port int) (libhdplatinum.Shade, error) {
	shades := libhdplatinum.GetShades(ip, port)
	for _, s := range shades {
		if s.Id() == id {
			return s, nil
		}
	}
	return libhdplatinum.Shade{}, errors.New("Not found")
}

func GetShades(ip string, port int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, getShadeViews(ip, port))
	}
}

func GetShade(ip string, port int) gin.HandlerFunc {
	return func(c *gin.Context) {
		shade, err := findShade(c.Param("id"), ip,port)
		if (err != nil) {
			c.JSON(http.StatusNotFound, Response{Code: http.StatusNotFound, Message: "Not found"})
		} else {
			c.JSON(http.StatusOK, ShadeViewFromShade(shade))
		}
	}
}

func UpdateShade(ip string, port int) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		shade, err := findShade(id, ip, port)
		if (err != nil) {
			c.JSON(http.StatusNotFound, Response{Code: http.StatusNotFound, Message: "Not found"})
		} else {
			var newShade ShadeView
			bindErr := c.BindJSON(&newShade)
			if bindErr == nil {
				newHeight := int(float64(newShade.Height) / 100 * MaxShadeHeight + 0.5)
				shade.SetHeight(newHeight)
				shade, _ := findShade(id, ip, port)
				c.JSON(http.StatusOK, ShadeViewFromShade(shade))
			} else {
				c.JSON(http.StatusBadRequest, Response{Code: http.StatusBadRequest, Message: "Bad request"})
			}
		}
	}
}
