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
	"github.com/op/go-logging"
	"time"
)

const MaxShadeHeight = 255

var log = logging.MustGetLogger("platypus")

type ShadeView struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	RoomId string `json:"roomId"`
	Height int `json:"height"`
}

func ShadeViewFromShade(s libhdplatinum.Shade) ShadeView {
	return ShadeView{Id: s.Id(), Name: s.Name(), RoomId: s.RoomId(), Height: s.Height() * 100 / MaxShadeHeight}
}

type Response struct {
	Code    int `json:"code"`
	Message string `json:"message"`
}

type ShadeDataCache struct {
	ShadeData ([]libhdplatinum.Shade)
	CacheTime time.Time
}

type Config struct {
	CacheTimeoutSeconds int
	Ip                  string
	Port                int
	Verbose             bool
}

// CACHES
var shadeDataCache ShadeDataCache

// MAIN CONFIG
var config Config

func validate(c *cli.Context) *cli.ExitError {
	if strings.TrimSpace(config.Ip) == "" {
		return cli.NewExitError("IP address is required. `platypus -h` for usage info.", 1)
	}
	if net.ParseIP(config.Ip) == nil {
		return cli.NewExitError("Invalid IP address", 1)
	}
	if config.Port < 0 || config.Port > 65535 {
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
		cli.IntFlag{
			Name: "ttl",
			Value: 10,
			Usage: "How long, in seconds, to keep shade data cached in memory.",
		},
		cli.BoolFlag{
			Name: "d",
			Usage: "Output debug logs",
		},
		cli.BoolFlag{
			Name: "help, h",
			Usage: "Show usage help.",
		},
	}

	app.Flags = flags

	app.Action = func(c *cli.Context) error {
		config = Config{
			CacheTimeoutSeconds: c.Int("ttl"),
			Ip: c.String("hdp-ip"),
			Port: c.Int("hdp-port"),
			Verbose: c.Bool("d"),
		}
		initLogging()
		exitError := validate(c)
		if exitError != nil {
			return exitError
		}
		r := gin.Default()
		r.GET("/shades", GetShades)
		r.GET("/shades/:id", GetShade)
		r.PUT("/shades/:id", UpdateShade)
		r.PUT("/shades/:id/height", UpdateShade)

		// super not restful, but makes scripting so easy...
		r.GET("/shades/:id/height", UpdateShade)

		refreshShadeCache()
		r.Run() // listen and server on 0.0.0.0:8080
		return nil
	}

	app.Run(os.Args)
}

func getShadeViews() []ShadeView {
	output := []ShadeView{}
	for _, s := range getShadeData() {
		output = append(output, ShadeViewFromShade(s))
	}

	return output
}

func findShade(id string) (libhdplatinum.Shade, error) {
	shades := getShadeData()
	for _, s := range shades {
		if s.Id() == id {
			return s, nil
		}
	}
	return libhdplatinum.Shade{}, errors.New("Not found")
}

func getShadeData() ([]libhdplatinum.Shade) {
	if (time.Since(shadeDataCache.CacheTime).Seconds() > 10) {
		log.Info("Shade data cache is too old. Refreshing.")
		refreshShadeCache()
	}
	return shadeDataCache.ShadeData
}

func refreshShadeCache() {
	log.Debug("Refreshed shade data")
	shadeDataCache = ShadeDataCache{ShadeData: libhdplatinum.GetShades(config.Ip, config.Port), CacheTime: time.Now()}
}

func GetShades(c *gin.Context) {
	log.Debug("Getting all shade information")
	c.JSON(http.StatusOK, getShadeViews())
}

func GetShade(c *gin.Context) {
	shade, err := findShade(c.Param("id"))
	if (err != nil) {
		c.JSON(http.StatusNotFound, Response{Code: http.StatusNotFound, Message: "Not found"})
	} else {
		c.JSON(http.StatusOK, ShadeViewFromShade(shade))
	}
}

func UpdateShade(c *gin.Context) {
	id := c.Param("id")
	log.Info("Updating shade", id)
	shade, err := findShade(id)
	if (err != nil) {
		c.JSON(http.StatusNotFound, Response{Code: http.StatusNotFound, Message: "Not found"})
	} else {
		var newShade ShadeView
		bindErr := c.BindJSON(&newShade)
		if bindErr == nil {
			newHeight := int(float64(newShade.Height) / 100 * MaxShadeHeight + 0.5)
			shade.SetHeight(newHeight)
			shade, _ := findShade(id)
			c.JSON(http.StatusOK, ShadeViewFromShade(shade))
		} else {
			c.JSON(http.StatusBadRequest, Response{Code: http.StatusBadRequest, Message: "Bad request"})
		}
	}
}

func initLogging() {
	var formatter = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfile} ▶ %{level} %{id:03x}%{color:reset} %{message}`)
	logging.SetFormatter(formatter)
	level := logging.INFO
	if config.Verbose {
		level = logging.DEBUG
	}
	logging.SetLevel(level, "platypus")
}
