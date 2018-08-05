package controller

import (
	"fmt"
	"net/http"
	"time"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/rutesun/reservation/log"
	"github.com/rutesun/reservation/reservation"
)

const dateFormat = "2006-01-02"

func RoomsController(s *reservation.Service) func(context *gin.Context) {
	return func(c *gin.Context) {
		if res, err := s.RoomList(); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"result": res,
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	}
}

func ListController(s *reservation.Service) func(context *gin.Context) {
	return func(c *gin.Context) {
		startStr := c.Query("startDate")
		endStr := c.Query("endDate")

		loc, _ := time.LoadLocation("Local")
		startDate, err := time.ParseInLocation(dateFormat, startStr, loc)
		endDate, err := time.ParseInLocation(dateFormat, endStr, loc)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 날짜 형식입니다 (ex: yyyy-MM-dd)"})
			return
		}

		if res, err := s.List(startDate, endDate); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"result": res,
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	}
}

type reservationRequest struct {
	RoomID    string    `form:"room_id" binding:"required"`
	UserName  string    `form:"user_name" binding:"required"`
	Repeat    string    `form:"repeat"`
	StartTime time.Time `form:"start_time" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime   time.Time `form:"end_time" binding:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	//StartTime time.Time `form:"start_time" binding:"required" time_format:"2006-01-02T15:04"`
	//EndTime   time.Time `form:"end_time" binding:"required" time_format:"2006-01-02T15:04"`
}

func MakeController(s *reservation.Service) func(context *gin.Context) {
	return func(c *gin.Context) {
		req := reservationRequest{}
		err := c.ShouldBindWith(&req, binding.Form)
		if err != nil {
			log.Error(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Println(req)
		roomId, err := strconv.Atoi(req.RoomID)
		if err != nil {
			log.Error(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		repeat, err := strconv.Atoi(req.Repeat)
		if err != nil {
			log.Error(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.Make(int64(roomId), req.UserName, req.StartTime, req.EndTime, reservation.ExtraInfo{Repeat: repeat}); err == nil {
			c.JSON(http.StatusOK, gin.H{"result": "OK"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	}
}

func CancelController(s *reservation.Service) func(context *gin.Context) {
	return func(c *gin.Context) {
		idStr := c.Param("id")

		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 id 형식입니다."})
			return
		}

		if res, err := s.Cancel(int64(id)); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"result": res,
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	}
}
