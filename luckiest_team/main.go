package main

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/famendola1/yauth"
	"github.com/famendola1/yfantasy"
)

func main() {
	// Initialize Yahoo OAuth.
	auth, err := yauth.CreateYAuthFromJSON("yauth.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Initiialize Yahoo Fantasy API Client
	yf := yfantasy.New(auth.Client())

	// Create the game and get the league object.
	gm := yfantasy.NewGame("nba", yf)
	lg, err := gm.GetLeagueByName("NBA Fantasy 2K22")
	if err != nil {
		fmt.Println(err)
		return
	}

	margins := map[int]float64{
		5:  .002, // FG%
		8:  .002, // FT%
		10: 2,    // 3PM
		12: 2,    // PTS
		15: 2,    // REB
		16: 2,    // AST
		17: 2,    // STL
		18: 2,    // BLK
		19: 2,    // TOV
	}

	results1 := []*result{}
	results2 := []*result{}
	for i := 1; i < 21; i++ {
		matchups, err := lg.GetScoreboard(i)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, matchup := range matchups.Matchup {
			results1 = append(results1, populateResult(&matchup))
			results2 = append(results2, adjustMatchupResult(&matchup, margins))
		}
	}

	standings := calcWinPct(calculateRecords(results2))
	sort.Sort(standings)
	printStandings(standings)
}

type result struct {
	teamA    string
	teamB    string
	teamAwon int
	teamBwon int
}

type record struct {
	wins   int
	losses int
	ties   int
}

func (r *record) lost() {
	r.losses++
}

func (r *record) won() {
	r.wins++
}

func (r *record) tied() {
	r.ties++
}

type pair struct {
	name  string
	value float64
}

type pairlist []pair

func (p pairlist) Len() int           { return len(p) }
func (p pairlist) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p pairlist) Less(i, j int) bool { return p[i].value > p[j].value }

func calculateRecords(results []*result) map[string]*record {
	records := make(map[string]*record)
	for _, res := range results {
		if _, found := records[res.teamA]; !found {
			records[res.teamA] = &record{}
		}
		if _, found := records[res.teamB]; !found {
			records[res.teamB] = &record{}
		}
		if res.teamAwon == res.teamBwon {
			records[res.teamA].tied()
			records[res.teamB].tied()
			continue
		}

		if res.teamAwon > res.teamBwon {
			records[res.teamA].won()
			records[res.teamB].lost()
			continue
		}

		if res.teamBwon > res.teamAwon {
			records[res.teamB].won()
			records[res.teamA].lost()
			continue
		}
	}

	return records
}

func populateResult(matchup *yfantasy.Matchup) *result {
	return &result{
		teamA:    matchup.Teams.Team[0].Name,
		teamB:    matchup.Teams.Team[1].Name,
		teamAwon: matchup.Teams.Team[0].TeamPoints.Total,
		teamBwon: matchup.Teams.Team[1].TeamPoints.Total,
	}
}

func adjustMatchupResult(matchup *yfantasy.Matchup, margins map[int]float64) *result {
	var teamAPoints int
	var teamBPoints int

	for i := range matchup.Teams.Team[0].TeamStats.Stats.Stat {
		statID := matchup.Teams.Team[0].TeamStats.Stats.Stat[i].StatID
		if _, found := margins[statID]; !found {
			continue
		}

		teamAValue, _ := strconv.ParseFloat(matchup.Teams.Team[0].TeamStats.Stats.Stat[i].Value, 3)
		teamBValue, _ := strconv.ParseFloat(matchup.Teams.Team[1].TeamStats.Stats.Stat[i].Value, 3)

		if statID == 19 {
			teamAValue *= -1
			teamBValue *= -1
		}

		if teamAValue == teamBValue {
			continue
		}

		if teamAValue > teamBValue && teamAValue-teamBValue < margins[statID] {
			teamBPoints++
			continue
		}

		if teamBValue > teamAValue && teamBValue-teamAValue < margins[statID] {
			teamAPoints++
			continue
		}

		if teamAValue > teamBValue {
			teamAPoints++
			continue
		}

		if teamBValue > teamAValue {
			teamBPoints++
			continue
		}
	}

	return &result{
		teamA:    matchup.Teams.Team[0].Name,
		teamB:    matchup.Teams.Team[1].Name,
		teamAwon: teamAPoints,
		teamBwon: teamBPoints,
	}
}

func calcWinPct(records map[string]*record) pairlist {
	pairs := []pair{}
	for k, v := range records {
		winPct := (float64(v.wins) + (float64(v.ties) / 2)) / float64(v.losses+v.wins+v.ties)
		pairs = append(pairs, pair{name: k, value: winPct})
	}
	return pairs
}

func printStandings(standings pairlist) {
	for i, pair := range standings {
		fmt.Printf("%d. %s: %.3f\n", i+1, pair.name, pair.value)
	}
}

func printResults(res1, res2 *result) {
	fmt.Printf("%s vs %s: %d/%d/%d --> %d/%d/%d\n", res1.teamA, res1.teamB, res1.teamAwon, res1.teamBwon, 9-res1.teamAwon-res1.teamBwon, res2.teamAwon, res2.teamBwon, 9-res2.teamAwon-res2.teamBwon)
}
