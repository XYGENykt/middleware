package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	CurrentConditions struct {
		Temp       float64 `json:"temp"`
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
	s.GET("/weather", Weather)

	err = s.Start(":8080")

	if err != nil {
		log.Fatal(err)
	}

}

func Weather(ctx echo.Context) error {
	// Формируем URL запроса
	apiKey := os.Getenv("WEATHER_API_KEY")
	url := "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/Sankt-Peterburg/2022-01-01T15:00:00?key=" + apiKey + "&include=current"
	fmt.Print(url)
	// Выполняем GET-запрос
	resp, err := http.Get(url)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch weather data"})
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read response"})
	}

	// Возвращаем сырой JSON (можно распарсить в структуру WeatherResponse)
	return ctx.JSONBlob(http.StatusOK, body)
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

func getThursdays(startYear, endYear int) []string {

	var thursdays []string

	// Устанавливаем начальную и конечную даты
	start := time.Date(startYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(endYear, time.December, 31, 23, 59, 59, 0, time.UTC)

	// Перебираем дни и добавляем четверги в массив
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Thursday {
			thursdays = append(thursdays, d.Format("2006-01-02"))
		}
	}

	return thursdays
}
