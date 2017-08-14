package main

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
)

const (
	// tournament statuses
	TS_OPEN   = 0
	TS_CLOSED = 1
)

type Tournament struct {
	db *sql.DB
}

func RoundUp(value float64) int {
	return int(value + 0.99)
}

// TakePoints takes points from player's account
func (t *Tournament) TakePoints(tr *sql.Tx, playerId string, points int) error {
	if points <= 0 {
		return Invalid("Invalid points")
	}

	var balance int
	err := tr.QueryRow(
		"SELECT balance FROM players WHERE id=$1 FOR UPDATE", playerId,
	).Scan(&balance)
	switch {
	case err == sql.ErrNoRows:
		return Invalid("Unknown player")
	case err != nil:
		return err
	}

	newBalance := balance - points
	if newBalance < 0 {
		return Invalid("No enough points")
	}
	_, err = tr.Exec("UPDATE players SET balance=$1 WHERE id=$2", newBalance, playerId)
	if err != nil {
		return err
	}
	return nil
}

// FundPoints adds points to player's balance
// If player is not exists, it will be created
func (t *Tournament) FundPoints(tr *sql.Tx, playerId string, points int) error {
	if points <= 0 {
		return Invalid("Invalid points")
	}

	var balance int
	err := tr.QueryRow(
		"SELECT balance FROM players WHERE id=$1 FOR UPDATE", playerId,
	).Scan(&balance)

	if err == sql.ErrNoRows {
		// No user found, create it!
		_, err = tr.Exec("INSERT INTO players (id, balance) VALUES ($1, $2)", playerId, points)
		return err
	}
	if err != nil {
		return err
	}

	newBalance := balance + points
	_, err = tr.Exec("UPDATE players SET balance=$1 WHERE id=$2", newBalance, playerId)
	if err != nil {
		return err
	}
	return nil
}

// AnnounceTournament creates tournament with some deposit value
func (t *Tournament) AnnounceTournament(db *sql.DB, id string, deposit int) error {
	if len(id) == 0 {
		return Invalid("Invalid tournament id")
	}
	if deposit <= 0 {
		return Invalid("Invalid deposit value")
	}

	_, err := db.Exec(
		"INSERT INTO tournaments (id, deposit, status) VALUES ($1, $2, $3)",
		id, deposit, TS_OPEN,
	)
	if err != nil {
		return err
	}

	return nil
}

// JoinTournament joins player into tournament and is he backed by a set of backers
// Backing is not mandatory and a player can be play on his own money
func (t *Tournament) JoinTournament(tr *sql.Tx, id string, playerId string, backers []string) error {
	var deposit int
	err := tr.QueryRow(
		"SELECT deposit FROM tournaments WHERE id=$1 AND status=$2 FOR UPDATE",
		id, TS_OPEN,
	).Scan(&deposit)
	switch {
	case err == sql.ErrNoRows:
		return Invalid("Unknown or closed tournament")
	case err != nil:
		return err
	}

	perPlayerPoints := RoundUp(float64(deposit) / float64(len(backers)+1))
	// try to take points from players balances
	for _, pId := range append(backers, playerId) {
		err = t.TakePoints(tr, pId, perPlayerPoints)
		if err != nil {
			return err
		}
	}
	_, err = tr.Exec(
		"INSERT INTO tournament_players (tournament_id, player_id, backers) VALUES ($1, $2, $3)",
		id, playerId, pq.Array(backers),
	)

	return err
}

// ResultTournament closes tournament and update balances for winners
func (t *Tournament) ResultTournament(tr *sql.Tx, id string, winners []Winner) error {
	var deposit int
	err := tr.QueryRow(
		"SELECT deposit FROM tournaments WHERE id=$1 AND status=$2 FOR UPDATE",
		id, TS_OPEN,
	).Scan(&deposit)
	switch {
	case err == sql.ErrNoRows:
		return Invalid("Unknown or closed tournament")
	case err != nil:
		return err
	}

	for _, winner := range winners {
		var backers pq.StringArray
		err = tr.QueryRow(
			"SELECT backers FROM tournament_players WHERE tournament_id=$1 AND player_id=$2",
			id, winner.PlayerId,
		).Scan(&backers)
		switch {
		case err == sql.ErrNoRows:
			return Invalid(fmt.Sprintf("Unknown player %s for this tournament", winner.PlayerId))
		case err != nil:
			return err
		}

		perPlayerPrize := int(winner.Prize / (len(backers) + 1))

		for _, pId := range append(backers, winner.PlayerId) {
			err = t.FundPoints(tr, pId, perPlayerPrize)
			if err != nil {
				return err
			}
		}
	}

	_, err = tr.Exec("UPDATE tournaments SET status=$1 WHERE id=$2", TS_CLOSED, id)
	return err
}

func (t *Tournament) Balance(db *sql.DB, playerId string) (int, error) {
	var balance int
	err := db.QueryRow("SELECT balance FROM players WHERE id=$1", playerId).Scan(&balance)
	switch {
	case err == sql.ErrNoRows:
		return 0, Invalid("Unknown player")
	case err != nil:
		return 0, err
	}
	return balance, nil
}
