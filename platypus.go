package main

import (
	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
	"github.com/urfave/cli"
	"github.com/vincer/platypus/lib"
	"net"
	"os"
	"strings"
)

var Log = lib.Log

func validate(c *cli.Context) *cli.ExitError {
	if strings.TrimSpace(lib.Config.Ip) == "" {
		return cli.NewExitError("IP address is required. `platypus -h` for usage info.", 1)
	}
	if net.ParseIP(lib.Config.Ip) == nil {
		return cli.NewExitError("Invalid IP address", 1)
	}
	if lib.Config.Port < 0 || lib.Config.Port > 65535 {
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
			Name:  "hdp-ip",
			Usage: "ip address of the Hunter Douglas Platinum Gateway. Required.",
			EnvVar: "PLATYPUS_HDP_IP",
		},
		cli.IntFlag{
			Name:  "hdp-port",
			Value: 522,
			Usage: "port of the Hunter Douglas Platinum Gateway.",
		},
		cli.IntFlag{
			Name:  "ttl",
			Value: 10,
			Usage: "How long, in seconds, to keep shade data cached in memory.",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "Output debug logs",
		},
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "Show usage help.",
		},
	}

	app.Flags = flags

	app.Action = func(c *cli.Context) error {
		lib.Config = lib.ConfigType{
			CacheTimeoutSeconds: c.Int("ttl"),
			Ip:                  c.String("hdp-ip"),
			Port:                c.Int("hdp-port"),
			Verbose:             c.Bool("d"),
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

		lib.RefreshShadeCache()

		// start workers
		lib.StartDispatcher(1)

		r.Run() // listen and server on 0.0.0.0:8080
		return nil
	}

	app.Run(os.Args)
}

func GetShades(c *gin.Context) {
	Log.Debug("Getting all shade information")
	c.JSON(200, lib.GetShadeViews())
}

func GetShade(c *gin.Context) {
	shade, err := lib.FindShade(c.Param("id"))
	if err != nil {
		c.JSON(404, lib.Response{Code: 400, Message: "Not found"})
	} else {
		c.JSON(200, lib.ShadeViewFromShade(shade))
	}
}

func UpdateShade(c *gin.Context) {
	id := c.Param("id")
	var newShade lib.ShadeView
	bindErr := c.BindJSON(&newShade)
	if bindErr == nil {
		newHeight := int(float64(newShade.Height)/100*lib.MaxShadeHeight + 0.5)
		Log.Info("Queueing update request")
		work := lib.UpdateRequest{Id: id, Height: newHeight}
		lib.UpdateQueue <- work

		//shade, _ := findShade(id)
		c.JSON(200, newShade)
	} else {
		c.JSON(400, lib.Response{Code: 400, Message: "Bad request"})
	}

	//return
	//log.Info("Updating shade", id)
	//shade, err := findShade(id)
	//if (err != nil) {
	//	c.JSON(404, lib.Response{Code: 404, Message: "Not found"})
	//} else {
	//	shade.SetHeight(newHeight)
	//}
}

func initLogging() {
	var formatter = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfile} ▶ %{level} %{id:03x}%{color:reset} %{message}`)
	logging.SetFormatter(formatter)
	level := logging.INFO
	if lib.Config.Verbose {
		level = logging.DEBUG
	}
	logging.SetLevel(level, "platypus")
}
