package main

import (
	"fmt"

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

	// Search for the player to add.
	add, err := lg.SearchPlayers("Malik Beasley")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Search for the player to drop.
	drop, err := lg.SearchPlayers("Cedi Osman")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get your team.
	tm, err := lg.UserTeam()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Perform the add/drop transaction.
	if err := tm.AddDrop(add[0].PlayerKey, drop[0].PlayerKey); err != nil {
		fmt.Println(err)
		return
	}
}
