package main

import (
	"time"

	"github.com/go-playground/validator"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	endpoint = "http://api.openweathermap.org/data/2.5"
	apiKey   = "42cf266142d52481c3e95edb22cad945"

	sleepTime = 250 * time.Millisecond
)

var (
	pool *redis.Pool
)

type (
	CustomValidator struct {
		validator *validator.Validate
	}
)

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func newCustomValidator() *CustomValidator {
	return &CustomValidator{validator.New()}
}

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(addr)
		},
	}
}

func main() {
	pool = newPool("redis://weather:@localhost:6379/0")
	defer pool.Close()

	e := echo.New()
	e.Validator = newCustomValidator()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/forecast", forecast)
	e.GET("/find", find)
	e.GET("/for-place", forPlace)

	e.Logger.Fatal(e.Start(":1323"))
}
