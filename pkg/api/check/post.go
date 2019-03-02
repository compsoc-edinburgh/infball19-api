package check

import (
	"net/http"
	"strconv"
	"time"

	"github.com/compsoc-edinburgh/infball19-api/pkg/api/base"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

func (i *Impl) Post(c *gin.Context) {

	startTime, _ := time.Parse(time.RFC3339, "2019-04-06T16:30:00+00:00")
	currentTime, _ := time.Now()

	if startTime.Unix() > currentTime.Unix() {
		base.BadRequest(c, "Guests can't be checked in until 16:30 on the 6th of April!")
		return
	}

	var result struct {
		AuthToken string
	}

	if err := c.BindJSON(&result); err != nil {
		base.BadRequest(c, err.Error())
		return
	}

	client := redis.NewClient(&redis.Options{
		Addr:     i.Config.Redis.Address,
		Password: i.Config.Redis.Password,
		DB:       i.Config.Redis.DB,
	})

	checkedIn := redis.NewClient(&redis.Options{
		Addr:     i.Config.Redis.Address,
		Password: i.Config.Redis.Password,
		DB:       i.Config.Redis.DB + 1,
	})

	val, err := client.Get(result.AuthToken).Result()
	if err == redis.Nil {
		timestamp, err := checkedIn.Get(result.AuthToken).Result()
		if err == redis.Nil {
			base.BadRequest(c, "Invalid auth token")
			return
		} else {
			base.BadRequest(c, "This ticket was already checked in at "+timestamp)
			return
		}
	}

	hour, min, sec := currentTime.Clock()
	checkedIn.Set(result.AuthToken, strconv.FormatInt(hour, 10)+":"+strconv.FormatInt(min, 10)+":"+strconv.FormatInt(sec, 10), 0)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   true,
	})

}
