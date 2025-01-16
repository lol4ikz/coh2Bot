package db

import "fmt"

type Match struct {
	ID     int
	MAP    string
	TEAM1  string
	TEAM2  string
	SCORE1 int
	SCORE2 int
	DATE   string
}

type Matches []Match

// type Match implements an interface Stringer
func (m *Match) String() string {
	return fmt.Sprintf("%-30s | %-8d | %-8d | %-10s\n", m.MAP, m.SCORE1, m.SCORE2, m.DATE)
}

func (ms Matches) String() string {
	res := "*Game Scores*\n| Map                     | Team2  | Team1 | Date       |\n|--------------------------------|--------|------|------------|\n"
	for _, m := range ms {
		res += m.String()
	}
	return res
}
