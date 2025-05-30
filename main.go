package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
	fmt.Println("Server running")

	s := echo.New()

	s.GET("/status", Handler)

	err := s.Start(":8080")

	if err != nil {
		log.Fatal(err)
	}

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
