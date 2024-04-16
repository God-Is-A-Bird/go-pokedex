package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/God-Is-A-Bird/pokedexcli/internal/pokecache"
	"github.com/God-Is-A-Bird/pokedexcli/internal/pokedex"
)

type cliCommand struct {
	name        string
	description string
	callback    func(args ...string) error
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Print help information.",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the program",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Print the next 20 locations",
			callback:    commandMapf,
		},
		"mapb": {
			name:        "mapb",
			description: "Print the previous 20 locations",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Usage: `explore [area]` Print the pokemon found in this area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Usage: `catch [pokemon]` Attempt to catch pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Usage: `inspect [pokemon]` Print information about a pokemon you've caught",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Print all the pokemon found in your pokedex",
			callback:    commandPokedex,
		},
	}
}

func commandHelp(arg ...string) error {
	fmt.Print("\nWelcome to Pokedex!\nUsage:\n\n")
	for _, command := range getCommands() {
		fmt.Print(command.name, ": ", command.description, "\n")
	}
	fmt.Println()
	return nil
}

func commandExit(arg ...string) error {
	os.Exit(0)
	return nil
}

type config struct {
	previous *string
	next     *string
}

var site string = "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"
var conf config = config{next: &site, previous: nil}

func processMap(url string) error {

	type Response struct {
		Count    int     `json:"count"`
		Next     *string `json:"next"`
		Previous *string `json:"previous"`
		Results  []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"results"`
	}

	body, inCache := pokecache.Cache.Get(url)

	if !inCache {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", url, nil)
		res, _ := client.Do(req)

		body, _ = io.ReadAll(res.Body)
		pokecache.Cache.Add(url, body)

	}

	r := Response{}
	err := json.Unmarshal(body, &r)
	if err != nil {
		fmt.Print(err)
	}

	for _, location := range r.Results {
		fmt.Println(location.Name)
	}

	conf.previous = r.Previous
	conf.next = r.Next

	return nil
}

func commandMapf(arg ...string) error {

	if conf.next == nil {
		fmt.Println("At the end!")
		return nil
	}

	processMap(*conf.next)

	return nil
}

func commandMapb(arg ...string) error {

	if conf.previous == nil {
		fmt.Println("At the start!")
		return nil
	}

	processMap(*conf.previous)

	return nil
}

func processExplore(url string) error {

	type Response struct {
		EncounterMethodRates []struct {
			EncounterMethod struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"encounter_method"`
			VersionDetails []struct {
				Rate    int `json:"rate"`
				Version struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"version"`
			} `json:"version_details"`
		} `json:"encounter_method_rates"`
		GameIndex int `json:"game_index"`
		ID        int `json:"id"`
		Location  struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"location"`
		Name  string `json:"name"`
		Names []struct {
			Language struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"language"`
			Name string `json:"name"`
		} `json:"names"`
		PokemonEncounters []struct {
			Pokemon struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"pokemon"`
			VersionDetails []struct {
				EncounterDetails []struct {
					Chance          int   `json:"chance"`
					ConditionValues []any `json:"condition_values"`
					MaxLevel        int   `json:"max_level"`
					Method          struct {
						Name string `json:"name"`
						URL  string `json:"url"`
					} `json:"method"`
					MinLevel int `json:"min_level"`
				} `json:"encounter_details"`
				MaxChance int `json:"max_chance"`
				Version   struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"version"`
			} `json:"version_details"`
		} `json:"pokemon_encounters"`
	}

	body, inCache := pokecache.Cache.Get(url)

	if !inCache {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", url, nil)
		res, _ := client.Do(req)

		body, _ = io.ReadAll(res.Body)
		pokecache.Cache.Add(url, body)

	}

	r := Response{}
	err := json.Unmarshal(body, &r)
	if err != nil {
		fmt.Print(err)
	}

	fmt.Println("Found Pokemon:")
	for _, pokemon := range r.PokemonEncounters {
		fmt.Println(" - ", pokemon.Pokemon.Name)
	}

	return nil
}

func commandExplore(arg ...string) error {
	baseURL := "https://pokeapi.co/api/v2/location-area/"
	fmt.Println("Exploring", arg[1])
	processExplore(baseURL + arg[1])
	return nil
}

func processCatch(url string) error {

	body, inCache := pokecache.Cache.Get(url)
	if !inCache {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", url, nil)
		res, _ := client.Do(req)

		body, _ = io.ReadAll(res.Body)
		pokecache.Cache.Add(url, body)

	}

	r := pokedex.Pokemon{}
	err := json.Unmarshal(body, &r)
	if err != nil {
		fmt.Println("No such pokemon")
		return nil
	}

	fmt.Printf("Throwing a Pokeball at %s:\n", r.Name)

	if r.BaseExperience > rand.Intn(2*r.BaseExperience) {
		fmt.Println(r.Name, "escaped!")
	} else {
		fmt.Println(r.Name, "was caught!")
		pokedex.Pokedex.Add(r.Name, r)

	}

	return nil
}

func commandCatch(arg ...string) error {
	baseURL := "https://pokeapi.co/api/v2/pokemon/"
	processCatch(baseURL + arg[1])
	return nil
}

func commandInspect(arg ...string) error {
	pokemon, ok := pokedex.Pokedex.Get(arg[1])
	if !ok {
		fmt.Println("You have not caught this pokemon")
	} else {
		fmt.Println("Name:", pokemon.Name)
		fmt.Println("Height:", pokemon.Height)
		fmt.Println("Weight:", pokemon.Weight)
		fmt.Println("Stats:")
		for _, stat := range pokemon.Stats {
			fmt.Printf(" -%s: %v\n", stat.Stat.Name, stat.BaseStat)
		}
		fmt.Println("Types:")
		for _, typ := range pokemon.Types {
			fmt.Printf(" - %s\n", typ.Type.Name)
		}
	}
	return nil
}

func commandPokedex(arg ...string) error {
	fmt.Println("Your Pokedex:")
	pokedex.Pokedex.Print()
	return nil
}

func main() {
	reader := bufio.NewScanner(os.Stdin)
	pokecache.NewCache(time.Second * 5)

	for {
		fmt.Print("Pokedex > ")
		reader.Scan()
		words := strings.Fields(strings.ToLower(reader.Text()))
		command, exists := getCommands()[strings.ToLower(words[0])]

		if exists {
			err := command.callback(words...)
			if err != nil {
				fmt.Print(err)
			}
		} else {
			fmt.Print("Invalid Command\n\n")
		}
	}
}
