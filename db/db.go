package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	Conn *sql.DB
}

// initializing new db instance
func NewDatabase(name string) (*Database, error) {
	db, err := sql.Open("sqlite3", name)
	if err != nil {
		return nil, err
	}
	return &Database{Conn: db}, nil
}

func (db *Database) CreateNewTables() error {
	sqlStmt := []string{
		`CREATE TABLE IF NOT EXISTS teams(team_id INTEGER PRIMARY KEY AUTOINCREMENT, team_name TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS matches(
			match_id INTEGER PRIMARY KEY AUTOINCREMENT, 
			map TEXT NOT NULL, 
			team1_id INTEGER DEFAULT 1, 
			team2_id INTEGER DEFAULT 2, 
			date TEXT DEFAULT (DATE('now')),
			FOREIGN KEY (team1_id) REFERENCES teams (team_id), 
			FOREIGN KEY (team2_id) REFERENCES teams (team_id)
		);`,
		`CREATE TABLE IF NOT EXISTS scores(
			score_id INT, 
			team1_score INTEGER DEFAULT 0, 
			team2_score INTEGER DEFAULT 0, 
			FOREIGN KEY (score_id) REFERENCES matches (match_id)
		);`,
	}
	for _, stmt := range sqlStmt {
		_, err := db.Conn.Exec(stmt)
		if err != nil {
			//return err
			return fmt.Errorf("failed to execute statement %q: %w", stmt, err)
		}
	}
	return nil
}

func (db *Database) NewTeams(teamsStr string) error {
	var teamCounts int
	err := db.Conn.QueryRow("SELECT COUNT (*) FROM teams").Scan(&teamCounts)
	if teamCounts > 0 {
		return fmt.Errorf("Cant add teams")
	}
	teams := strings.SplitN(teamsStr, " ", 2)
	if len(teams) != 2 {
		return fmt.Errorf("Teams should be a 2 values")
	}
	// Prepare the sql stmt
	stmt, err := db.Conn.Prepare("INSERT INTO teams (team_name) VALUES (?);")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, team := range teams {
		// Execute the stmt
		_, err := stmt.Exec(team)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *Database) NewMatch(newMap string, score string) error {
	scores := strings.SplitN(score, "-", 2)
	if len(scores) != 2 {
		return fmt.Errorf("Scores should be 2 values")
	}
	// inserting a new match
	res, err := db.Conn.Exec("INSERT INTO matches (map) VALUES (?)", newMap)
	if err != nil {
		return err
	}
	matchId, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// splitting and converting a scores to int
	intScores := make([]int, len(scores))
	for i, s := range scores {
		scoreInt, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		intScores[i] = scoreInt
	}
	// inserting a scores
	res, err = db.Conn.Exec("INSERT INTO scores (score_id, team1_score, team2_score) VALUES (?,?,?)", matchId, intScores[0], intScores[1])
	if err != nil {
		return err
	}
	return nil
}

// getting a 10 last matches stats
func (db *Database) GetLastMatches() (Matches, error) {
	rows, err := db.Conn.Query(`SELECT m.match_id, m.map, t1.team_name AS TEAM1, t2.team_name AS TEAM2, s.team1_score, s.team2_score, m.date 
								FROM matches m 
								JOIN teams t1 ON m.team1_id = t1.team_id
								JOIN teams t2 ON m.team2_id = t2.team_id
								JOIN scores s ON m.match_id = s.score_id 
								ORDER BY m.date DESC
								LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var matches []Match
	for rows.Next() {
		var match Match
		err = rows.Scan(
			&match.ID,
			&match.MAP,
			&match.TEAM1,
			&match.TEAM2,
			&match.SCORE1,
			&match.SCORE2,
			&match.DATE,
		)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return matches, nil
}

func (db *Database) GetTotalScores() (int, int, error) {
	var total1, total2 int
	err := db.Conn.QueryRow(`SELECT SUM(team1_score), SUM(team2_score) FROM scores;`).Scan(&total1, &total2)
	if err != nil {
		return 0, 0, err
	}

	return total1, total2, nil
}
