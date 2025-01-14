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

func (m *Match) String() string {
	return fmt.Sprintf("%-3d | %-25s | %-10s | %-10s | %-8d | %-8d | %-10s\n", m.ID, m.MAP, m.TEAM1, m.TEAM2, m.SCORE1, m.SCORE2, m.DATE)
}

func (ms Matches) String() string {
	res := ""
	for _, m := range ms {
		res += m.String()
	}
	return res
}
