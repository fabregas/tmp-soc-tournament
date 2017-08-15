package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
)

var dbConnStr string

func init() {
	flag.StringVar(
		&dbConnStr,
		"db-conn",
		"postgres://postgres@pg-host/test_tournament?sslmode=disable",
		"Connection string to database",
	)
}

type ValidationError struct {
	msg string
}

func (e ValidationError) Error() string {
	return e.msg
}

func Invalid(msg string) ValidationError {
	return ValidationError{msg}
}

type BalanceResponse struct {
	PlayerId string `json:"playerId"`
	Balance  int    `json:"balance"`
}

type Winner struct {
	PlayerId string `json:"playerId"`
	Prize    int    `json:"prize"`
}

type TournamentResult struct {
	TournamentId string   `json:"tournamentId"`
	Winners      []Winner `json:"winners"`
}

func main() {
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)
	flag.Parse()

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	tournamentService := &Tournament{db}
	http.Handle("/", tournamentService)
	fmt.Println("Server is initialized and hosted at :8080")

	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	<-stopSignal
	db.Close()
}
