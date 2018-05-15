package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

type (
	ByPlace struct {
		Lat float32 `query:"lat" validate:"required"`
		Lon float32 `query:"lon" validate:"required"`
		DefaultParams
	}
)

func forPlace(c echo.Context) error {
	q := new(ByPlace)
	if err := c.Bind(q); err != nil {
		return incorrectParams
	}
	if err := c.Validate(q); err != nil {
		return incorrectParams
	}

	params := fmt.Sprintf("lat=%f&lon=%f&units=%s&lang=%s", q.Lat, q.Lon, q.Units, q.Lang)

	conn := pool.Get()
	defer conn.Close()

	res, err := query(params)
	if err != nil {
		return err
	}

	data := new(Forecast)
	err = json.Unmarshal(res, data)
	if err != nil {
		return err
	}
	res, err = json.Marshal(data)
	if err != nil {
		return err
	}

	params = fmt.Sprintf("id=%d&units=%s&lang=%s", data.City.ID, q.Units, q.Lang)

	if data.Cod == "200" {
		ttl := 3 * 60 * 60
		_, err = conn.Do("SET", params, res, "EX", ttl)
		if err != nil {
			return err
		}
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().Write(res)
	return c.NoContent(http.StatusOK)
}
