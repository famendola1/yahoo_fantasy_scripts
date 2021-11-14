package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/famendola1/yauth"
	"github.com/famendola1/yfantasy"
)

func main() {
	// Initialize Yahoo OAuth.
	auth, err := yauth.CreateYAuthFromJSON("<path to yauth JSON file>")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Initiialize Yahoo Fantasy API Client
	yf := yfantasy.New(auth.Client())

	// Create the game and get the league object.
	gm := yfantasy.NewGame("nba", yf)
	lg, err := gm.GetLeagueByName("<name of league>")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Fetch all add transactions"
	transactions, err := lg.Transactions([]string{"add"})
	if err != nil {
		fmt.Println(err)
		return
	}

	// Filter transactions for transactions on Nov 12, 2021.
	day := "2021-11-12"
	date, _ := time.Parse("2006-01-02", day)
	var filteredTransactions []*yfantasy.Transaction
	for _, t := range transactions {
		timestamp := toInt(t.Timestamp)
		transactionTime := time.Unix(int64(timestamp), 0)
		// Transactions are sorted by time, so we can exit early once we pass the
		// desired day.
		if transactionTime.Before(date) {
			break
		}
		tYear, tMonth, tDay := transactionTime.Date()
		dYear, dMonth, dDay := date.Date()
		if tDay == dDay && tMonth == dMonth && tYear == dYear {
			filteredTransactions = append(filteredTransactions, t)
		}
	}

	// Extract players from the transactions that were added.
	var addedPlayers []*yfantasy.Player
	for _, t := range filteredTransactions {
		for _, p := range t.Players.Player {
			if p.TransactionData.Type == "add" {
				addedPlayers = append(addedPlayers, p)
			}
		}
	}

	// Gather all the player keys for the added players.
	keys := make([]string, len(addedPlayers))
	for i, p := range addedPlayers {
		keys[i] = p.PlayerKey
	}

	// Get the stats from 2021-11-12 for each of the added players.
	stats, err := lg.GetPlayersStats(keys, yfantasy.StatDuration{DurationType: "date", Date: day})
	if err != nil {
		fmt.Println(err)
		return
	}

	// Calculate impact of each player by converting the 9CAT stats to points.
	// FGMi, FTMi, 3PM, PTS, REB, AST, STL, BLK, TOV
	points := []float64{-.5, -.5, 1.1, 1, 1.2, 1.2, 2, 2, -1}
	type player struct {
		Name   string
		Points float64
	}
	var players []player
	for _, s := range stats {
		statVals := extractStats(s.PlayerStats.Stats.Stat)
		total := dot(statVals, points)
		if total != 0 {
			players = append(players, player{s.Name.Full, total})
		}
	}

	// Sort by calculated points descending
	sort.Slice(players, func(i, j int) bool {
		return players[i].Points > players[j].Points
	})

	// Print the ranked players
	for i, p := range players {
		fmt.Printf("#%v %v: %v\n", i+1, p.Name, p.Points)
	}
}

// extractStats parses the stats from Yahoo and returns the parsed stats as a
// list of floats.
func extractStats(stats []yfantasy.Stat) []float64 {
	// Stats should be ordered as:
	// FGM/FGA
	// FG %
	// FTM/FGA
	// FT %
	// 3PM
	// PTS
	// REB
	// AST
	// STL
	// BLK
	// TOV
	parsedStats := make([]float64, 9)

	// Calculate FG missed.
	fg := strings.Split(stats[0].Value, "/")
	parsedStats[0] = toFloat(fg[1]) - toFloat(fg[0])

	// Calculate FT missed.
	ft := strings.Split(stats[2].Value, "/")
	parsedStats[1] = toFloat(ft[1]) - toFloat(ft[0])

	for i, stat := range stats[4:] {
		parsedStats[i+2] = toFloat(stat.Value)
	}

	return parsedStats
}

// toInt converts string to int.
func toInt(str string) int {
	asInt, _ := strconv.Atoi(str)
	return asInt
}

// toFloat converts string to float.
func toFloat(str string) float64 {
	asFloat, _ := strconv.ParseFloat(str, 64)
	return asFloat
}

// dot takes the dot product of two lists.
func dot(a []float64, b []float64) float64 {
	var sum float64
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}
