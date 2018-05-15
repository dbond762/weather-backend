# weather-backend
## Описание
Бекенд для сервиса прогноза погоды

Для прогноза погоды используется [OpenWeatherMap API][1]. Для уменьшения нагрузки на API прогнозы кешируются в redis.

## Как запустить
```bash
git clone https://github.com/dbond762/weather-backend.git
cd weather-backend
go get -u github.com/golang/dep/cmd/dep
dep ensure
go build src/main/main.go
./main
```

[1]: https://openweathermap.org/api
