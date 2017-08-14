
CREATE TABLE players (
        id VARCHAR(64) PRIMARY KEY,
        balance INTEGER DEFAULT 0
);

CREATE TABLE tournaments (
	id VARCHAR(64) PRIMARY KEY,
	deposit INTEGER,
	status INTEGER
);

CREATE TABLE tournament_players (
	tournament_id VARCHAR(64) REFERENCES tournaments,
	player_id VARCHAR(64) REFERENCES players,
	backers VARCHAR(64)[]
)

