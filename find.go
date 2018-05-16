package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
)

type (
	Find struct {
		Cod   string `json:"cod"`
		Count int    `json:"count"`
		List  []struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Coord struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"coord"`
			Main struct {
				Temp     float64 `json:"temp"`
				Pressure float64 `json:"pressure"`
				Humidity int     `json:"humidity"`
			} `json:"main"`
			Dt   int `json:"dt"`
			Wind struct {
				Speed float64 `json:"speed"`
				Deg   float64 `json:"deg"`
			} `json:"wind"`
			Clouds struct {
				All int `json:"all"`
			} `json:"clouds"`
			Weather []struct {
				Description string `json:"description"`
				Icon        string `json:"icon"`
			} `json:"weather"`
		} `json:"list"`
	}

	FindQuery struct {
		Q string `query:"q" validate:"required"`
		DefaultParams
	}
)

func find(c echo.Context) error {
	q := new(FindQuery)
	if err := c.Bind(q); err != nil {
		return incorrectParams
	}
	if err := c.Validate(q); err != nil {
		return incorrectParams
	}

	params := fmt.Sprintf("q=%s&units=%s&lang=%s", q.Q, q.Units, q.Lang)

	client := http.Client{}
	url := fmt.Sprintf("%s/find?%s&APPID=%s", endpoint, params, apiKey)
	resp, err := client.Get(url)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	defer resp.Body.Close()

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	data := new(Find)
	err = json.Unmarshal(res, data)
	if err != nil {
		return err
	}
	res, err = json.Marshal(data)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().Write(res)
	return c.NoContent(http.StatusOK)
}
