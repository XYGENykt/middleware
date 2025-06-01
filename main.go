package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

// WeatherResponse - структура для парсинга JSON-ответа
type WeatherResponse struct {
	Address           string  `json:"address"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	Timezone          string  `json:"timezone"`
	Date              string  `json:"datetime"`
	CurrentConditions struct {
		TempF      float64 `json:"temp"`
		TempC      float64 `json:"tempC"`
		Date       string  `json:"datetime"`
		Humidity   float64 `json:"humidity"`
		Conditions string  `json:"conditions"`
	} `json:"currentConditions"`
}

func main() {
	// Загружаем .env-файл
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	fmt.Println("Server running")

	s := echo.New()

	s.GET("/status", Handler)
	s.GET("/weather", getWeather)   //последовательные запросы
	s.GET("/weatherg", getWeatherG) //вариант с горутиной

	err = s.Start(":8080")

	if err != nil {
		log.Fatal(err)
	}

}

func getWeather(ctx echo.Context) error {
	// Формируем URL запроса
	apiKey := os.Getenv("WEATHER_API_KEY")
	baseURL := "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/Sankt-Peterburg/"

	// вытаскиваю все четверги
	thursdays := getThursdays(2023, 2025)

	// Проверка
	// thursdays := []string{
	// 	"2022-01-01",
	// 	"2022-01-02",
	// 	"2022-01-03",
	// }

	// Создаем массив для хранения прогнозов
	var forecasts []WeatherResponse

	// Цикл с прогнозом для каждой даты

	for _, date := range thursdays {
		url := baseURL + date + "T15:00:00?key=" + apiKey + "&include=current"

		resp, err := http.Get(url)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to fetch weather data for date %s", date),
			})
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to read response for date %s", date),
			})
		}

		var weatherData WeatherResponse
		weatherData.Date = date

		if err := json.Unmarshal(body, &weatherData); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to parse weather data for date %s", date),
			})
		}

		// Формула переноса сельция в фаренгейт (1 °C × 9/5) + 32 = 33,8 и наоборот (32 °F − 32) × 5/9 = 0 °C
		weatherData.CurrentConditions.TempC = (weatherData.CurrentConditions.TempF - 32.0) * 5 / 9

		forecasts = append(forecasts, weatherData)
	}

	return ctx.JSON(http.StatusOK, forecasts)
}

func getWeatherG(ctx echo.Context) error {
	apiKey := os.Getenv("WEATHER_API_KEY")
	baseURL := "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/Sankt-Peterburg/"
	thursdays := getThursdays(2023, 2025)

	var (
		wg        sync.WaitGroup
		forecasts []WeatherResponse
		mu        sync.Mutex
		errors    []string
	)

	maxConcurrent := make(chan struct{}, 10)

	for _, date := range thursdays {
		wg.Add(1)
		maxConcurrent <- struct{}{}

		go func(d string) {
			defer wg.Done()
			defer func() { <-maxConcurrent }()

			url := baseURL + d + "T15:00:00?key=" + apiKey + "&include=current"

			resp, err := http.Get(url)

			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Failed to fetch weather data for date %s: %v", d, err))
				mu.Unlock()
				return
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Failed to read response for date %s: %v", d, err))
				mu.Unlock()
				return
			}

			var weatherData WeatherResponse
			if err := json.Unmarshal(body, &weatherData); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Sprintf("Failed to parse weather data for date %s: %v", d, err))
				mu.Unlock()
				return
			}

			weatherData.Date = d
			weatherData.CurrentConditions.TempC = (weatherData.CurrentConditions.TempF - 32.0) * 5 / 9

			mu.Lock()
			forecasts = append(forecasts, weatherData)
			mu.Unlock()
		}(date)
	}

	wg.Wait()

	if len(errors) > 0 {
		return ctx.JSON(http.StatusPartialContent, map[string]interface{}{
			"forecasts": forecasts,
			"errors":    errors,
			"message":   fmt.Sprintf("Success: %d, Failed: %d", len(forecasts), len(errors)),
		})
	}

	return ctx.JSON(http.StatusOK, forecasts)
}

func getThursdays(startYear, endYear int) []string {

	var thursdays []string

	// Устанавливаем начальную и конечную даты
	start := time.Date(startYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(endYear, time.Now().Month(), 0, 23, 59, 59, 0, time.UTC)

	// Перебираем дни и добавляем четверги в массив
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Thursday {
			thursdays = append(thursdays, d.Format("2006-01-02"))
		}
	}

	return thursdays
}

func Handler(ctx echo.Context) error {

	thursdays := getThursdays(2023, 2025)

	// Выводим результат
	fmt.Println("Все четверги с 2023 по 2025 год:")
	for _, thursday := range thursdays {
		fmt.Printf("%s\n", thursday)
	}

	// Можно использовать массив `thursdays` дальше
	s := fmt.Sprintf("\nВсего четвергов: %d\n", len(thursdays))

	err := ctx.String(http.StatusOK, s)
	if err != nil {
		return (err)
	}
	return nil
}
