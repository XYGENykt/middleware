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

	d := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	dur := time.Until(d)

	s := fmt.Sprintf("Количество дней:%d", int64(dur.Hours())/24)

	err := ctx.String(http.StatusOK, s)
	if err != nil {
		return (err)
	}
	return nil
}
