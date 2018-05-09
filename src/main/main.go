package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	endpoint = "http://api.openweathermap.org/data/2.5"
	apiKey   = "42cf266142d52481c3e95edb22cad945"

	sleepTime = 50 * time.Millisecond
)

var (
	pool *redis.Pool
)

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(addr)
		},
	}
}

func forecast(w http.ResponseWriter, r *http.Request) {
	const (
		idKey   = "id"
		unitKey = "units"
		langKey = "lang"

		defaultUnit = "metric"
	)

	id := r.URL.Query().Get(idKey)
	if len(id) == 0 {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}

	unit := r.URL.Query().Get(unitKey)
	if len(unit) == 0 {
		unit = defaultUnit
	}

	lang := r.URL.Query().Get(langKey)

	conn := pool.Get()
	defer conn.Close()

	reply, err := redis.Int(conn.Do("EXISTS", id+"_lock"))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	if reply == 1 {
		time.Sleep(sleepTime)
	}

	res, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		conn.Do("SET", id+"_lock", 1)

		client := http.Client{}
		url := fmt.Sprintf("%s/forecast?id=%s&units=%s&lang=%s&APPID=%s", endpoint, id, unit, lang, apiKey)
		resp, err := client.Get(url)
		if err != nil {
			writeErr(w, http.StatusNotFound, err)
			return
		}
		defer resp.Body.Close()

		res, err = ioutil.ReadAll(resp.Body)

		ttl := 3 * 60 * 60
		_, err = conn.Do("SET", id, res, "EX", ttl)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}

		_, err = conn.Do("DEL", id+"_lock")
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}

		log.Print("From API")
	} else if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	} else {
		log.Print("From cache")
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func writeErr(w http.ResponseWriter, code int, err error) {
	log.Print(err)
	http.Error(w, http.StatusText(code), code)
}

func main() {
	pool = newPool("redis://weather:@localhost:6379/0")
	defer pool.Close()

	http.HandleFunc("/forecast", forecast)

	port := "8080"
	log.Printf("Service run on localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
