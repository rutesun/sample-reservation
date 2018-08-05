package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rutesun/reservation/config"
	"github.com/rutesun/reservation/controller"
	"github.com/rutesun/reservation/mariadb"
	"github.com/rutesun/reservation/reservation"
)

func main() {
	r := gin.Default()
	r.Static("public", "public")

	r.LoadHTMLGlob("public/*.html")

	var setting *config.Setting
	if conf, err := config.Parse(); err != nil {
		panic(err)
	} else {
		if setting, err = config.Make(conf); err != nil {
			panic(err)
		}
	}

	db := mariadb.New(setting.DB)
	reservationService := reservation.New(db)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/rooms", controller.RoomsController(reservationService))
	r.GET("/reservations", controller.ListController(reservationService))
	r.POST("/reservation", controller.MakeController(reservationService))
	r.DELETE("/reservation/:id", controller.CancelController(reservationService))
	r.Run() // listen and serve on 0.0.0.0:8080

}
