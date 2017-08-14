package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func sendResp(w http.ResponseWriter, err error) {
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch err.(type) {
	case ValidationError:
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (t *Tournament) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("requested path: ", r.URL.Path)
	switch r.URL.Path {
	case "/take":
		err := t.handleTake(r)
		sendResp(w, err)
	case "/fund":
		err := t.handleFund(r)
		sendResp(w, err)
	case "/announceTournament":
		err := t.handleAnnounceTournament(r)
		sendResp(w, err)
	case "/joinTournament":
		err := t.handleJoinTournament(r)
		sendResp(w, err)
	case "/resultTournament":
		err := t.handlerResultTournament(r)
		sendResp(w, err)
	case "/balance":
		pId, balance, err := t.handleBalance(r)
		if err != nil {
			sendResp(w, err)
		} else {
			resp := BalanceResponse{pId, balance}
			json.NewEncoder(w).Encode(resp)
		}
	case "/reset":
		err := t.handleReset(r)
		sendResp(w, err)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (t *Tournament) handleTake(r *http.Request) error {
	if r.Method != http.MethodGet {
		return Invalid("invalid HTTP method")
	}
	q := r.URL.Query()
	pId := q.Get("playerId")
	if pId == "" {
		return Invalid("no playerId specified")
	}
	points := q.Get("points")
	if points == "" {
		return Invalid("no points specified")
	}

	iPoints, err := strconv.Atoi(points)
	if err != nil {
		return err
	}

	tr, err := t.db.Begin()
	if err != nil {
		return err
	}

	err = t.TakePoints(tr, pId, iPoints)
	if err != nil {
		tr.Rollback()
	} else {
		err = tr.Commit()
	}
	return err
}

func (t *Tournament) handleFund(r *http.Request) error {
	if r.Method != http.MethodGet {
		return Invalid("invalid HTTP method")
	}
	q := r.URL.Query()
	pId := q.Get("playerId")
	if pId == "" {
		return Invalid("no playerId specified")
	}
	points := q.Get("points")
	if points == "" {
		return Invalid("no points specified")
	}

	iPoints, err := strconv.Atoi(points)
	if err != nil {
		return err
	}

	tr, err := t.db.Begin()
	if err != nil {
		return err
	}

	err = t.FundPoints(tr, pId, iPoints)
	if err != nil {
		tr.Rollback()
	} else {
		err = tr.Commit()
	}

	return err

}

func (t *Tournament) handleAnnounceTournament(r *http.Request) error {
	if r.Method != http.MethodGet {
		return Invalid("invalid HTTP method")
	}
	q := r.URL.Query()
	tId := q.Get("tournamentId")
	if tId == "" {
		return Invalid("no tournamentId specified")
	}
	deposit := q.Get("deposit")
	if deposit == "" {
		return Invalid("no deposit specified")
	}
	iDeposit, err := strconv.Atoi(deposit)
	if err != nil {
		return err
	}

	return t.AnnounceTournament(t.db, tId, iDeposit)
}

func (t *Tournament) handleJoinTournament(r *http.Request) error {
	if r.Method != http.MethodGet {
		return Invalid("invalid HTTP method")
	}
	q := r.URL.Query()
	tId := q.Get("tournamentId")
	if tId == "" {
		return Invalid("no tournamentId specified")
	}
	pId := q.Get("playerId")
	if pId == "" {
		return Invalid("no playerId specified")
	}

	backers := q["backerId"]

	tr, err := t.db.Begin()
	if err != nil {
		return err
	}
	err = t.JoinTournament(tr, tId, pId, backers)
	if err != nil {
		tr.Rollback()
	} else {
		err = tr.Commit()
	}
	return err
}

func (t *Tournament) handlerResultTournament(r *http.Request) error {
	if r.Method != http.MethodPost {
		return Invalid("invalid HTTP method")
	}

	req := TournamentResult{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}
	r.Body.Close()

	tr, err := t.db.Begin()
	if err != nil {
		return err
	}
	err = t.ResultTournament(tr, req.TournamentId, req.Winners)
	if err != nil {
		tr.Rollback()
	} else {
		err = tr.Commit()
	}
	return err

}

func (t *Tournament) handleBalance(r *http.Request) (string, int, error) {
	if r.Method != http.MethodGet {
		return "", 0, Invalid("invalid HTTP method")
	}

	pId := r.URL.Query().Get("playerId")
	if pId == "" {
		return "", 0, Invalid("no playerId specified")
	}

	balance, err := t.Balance(t.db, pId)
	return pId, balance, err
}

func (t *Tournament) handleReset(r *http.Request) error {
	tr, err := t.db.Begin()
	if err != nil {
		return err
	}
	if _, err = tr.Exec("TRUNCATE TABLE tournaments CASCADE"); err != nil {
		tr.Rollback()
		return err
	}
	if _, err = tr.Exec("TRUNCATE TABLE players CASCADE"); err != nil {
		tr.Rollback()
		return err
	}

	return tr.Commit()
}
