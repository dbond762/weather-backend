package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
)

var (
	incorrectParams = echo.NewHTTPError(http.StatusNotFound, "Please specify the correct parameters")
)

type (
	Forecast struct {
		Cod  string `json:"cod"`
		List []struct {
			Dt   int `json:"dt"`
			Main struct {
				Temp     float64 `json:"temp"`
				Pressure float64 `json:"pressure"`
				Humidity int     `json:"humidity"`
			} `json:"main"`
			Weather []struct {
				Description string `json:"description"`
				Icon        string `json:"icon"`
			} `json:"weather"`
			Clouds struct {
				All int `json:"all"`
			} `json:"clouds"`
			Wind struct {
				Speed float64 `json:"speed"`
				Deg   float64 `json:"deg"`
			} `json:"wind"`
		} `json:"list"`
		City struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Coord struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"coord"`
			Country string `json:"country"`
		} `json:"city"`
	}

	DefaultParams struct {
		Units string `query:"units" validate:"omitempty,oneof=metric imperial"`
		Lang  string `query:"lang" validate:"omitempty,oneof=ar bg ca cz de el en fa fi fr gl
															 hr hu it ja kr la lt mk nl pl pt
															 ro ru se sk sl es tr ua vi zh_cn
															 zh_tw"`
	}

	ById struct {
		Id int `query:"id" validate:"required"`
		DefaultParams
	}
)

func forecast(c echo.Context) error {
	q := new(ById)
	if err := c.Bind(q); err != nil {
		return incorrectParams
	}
	if err := c.Validate(q); err != nil {
		return incorrectParams
	}

	params := fmt.Sprintf("id=%d&units=%s&lang=%s", q.Id, q.Units, q.Lang)

	conn := pool.Get()
	defer conn.Close()

	reply, err := redis.Int(conn.Do("EXISTS", params+"_lock"))
	if reply == 1 {
		time.Sleep(sleepTime)
	}

	res, err := redis.Bytes(conn.Do("GET", params))
	if err == redis.ErrNil {
		conn.Do("SET", params+"_lock", 1)
		defer conn.Do("DEL", params+"_lock")

		res, err = query(params)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound)
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

		if data.Cod == "200" {
			ttl := 3 * 60 * 60
			_, err = conn.Do("SET", params, res, "EX", ttl)
			if err != nil {
				return err
			}
		}

		c.Logger().Print("From API")
	} else if err != nil {
		return err
	} else {
		c.Logger().Print("From cache")
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().Write(res)
	return c.NoContent(http.StatusOK)
}

func query(params string) (res []byte, err error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/forecast?%s&APPID=%s", endpoint, params, apiKey)
	resp, err := client.Get(url)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound)
	}
	defer resp.Body.Close()

	res, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound)
	}

	return res, nil
}
