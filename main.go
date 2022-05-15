package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Repos struct {
	Repos []Repo `json:"repos"`
}

type User struct {
	Username string `json:"user"`
	LastSHA1 string `json:"lastSHA1"`
}

type Repo struct {
	Provider string `json:"provider"`
	Url      string `json:"url"`
	Branch   string `json:"branch"`
	LastSHA1 string `json:"lastSHA1"`
	Users    []User `json:"users"`
}

var repos Repos

func getConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, repos)
}

func getConfigPretty(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, repos, "  ")
}

func updateConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, "{}")
}

func wget() {

	resp, err := http.Get("http://gobyexample.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)

	scanner := bufio.NewScanner(resp.Body)
	for i := 0; scanner.Scan() && i < 5; i++ {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func save_json() {
	f, err := os.Create("last.json")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	// Convert structs to JSON.
	data, err := json.Marshal(repos)
	if err != nil {
		log.Fatal(err)
	}

	_, err2 := f.Write(data)

	if err2 != nil {
		log.Fatal(err2)
	}
}

func main() {

	jsonFile, err := os.Open("watch.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &repos)

	fmt.Printf("\n")
	fmt.Printf("	██████  ▄████▄   ███▄ ▄███▓    █     █░ ▄▄▄     ▄▄▄█████▓ ▄████▄   ██░ ██ ▓█████  ██▀███  \n")
	fmt.Printf("  ▒██    ▒ ▒██▀ ▀█  ▓██▒▀█▀ ██▒   ▓█░ █ ░█░▒████▄   ▓  ██▒ ▓▒▒██▀ ▀█  ▓██░ ██▒▓█   ▀ ▓██ ▒ ██▒\n")
	fmt.Printf("  ░ ▓██▄   ▒▓█    ▄ ▓██    ▓██░   ▒█░ █ ░█ ▒██  ▀█▄ ▒ ▓██░ ▒░▒▓█    ▄ ▒██▀▀██░▒███   ▓██ ░▄█ ▒\n")
	fmt.Printf("	▒   ██▒▒▓▓▄ ▄██▒▒██    ▒██    ░█░ █ ░█ ░██▄▄▄▄██░ ▓██▓ ░ ▒▓▓▄ ▄██▒░▓█ ░██ ▒▓█  ▄ ▒██▀▀█▄  \n")
	fmt.Printf("  ▒██████▒▒▒ ▓███▀ ░▒██▒   ░██▒   ░░██▒██▓  ▓█   ▓██▒ ▒██▒ ░ ▒ ▓███▀ ░░▓█▒░██▓░▒████▒░██▓ ▒██▒\n")
	fmt.Printf("  ▒ ▒▓▒ ▒ ░░ ░▒ ▒  ░░ ▒░   ░  ░   ░ ▓░▒ ▒   ▒▒   ▓▒█░ ▒ ░░   ░ ░▒ ▒  ░ ▒ ░░▒░▒░░ ▒░ ░░ ▒▓ ░▒▓░\n")
	fmt.Printf("  ░ ░▒  ░ ░  ░  ▒   ░  ░      ░     ▒ ░ ░    ▒   ▒▒ ░   ░      ░  ▒    ▒ ░▒░ ░ ░ ░  ░  ░▒ ░ ▒░\n")
	fmt.Printf("  ░  ░  ░  ░        ░      ░        ░   ░    ░   ▒    ░      ░         ░  ░░ ░   ░     ░░   ░ \n")
	fmt.Printf("		░  ░ ░             ░          ░          ░  ░        ░ ░       ░  ░  ░   ░  ░   ░     \n")
	fmt.Printf("		   ░                                                 ░                                \n")
	fmt.Printf("\n")

	update_repos()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/config/", getConfig)
	e.GET("/config/pretty", getConfigPretty)
	e.PUT("/config/", updateConfig)
	e.Logger.Fatal(e.Start(":1323"))
}
