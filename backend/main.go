package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Repos struct {
	Repos []Repo `json:"repos"`
}

type User struct {
	Username      string   `json:"user"`
	LastSHA1      string   `json:"lastSHA1"`
	ReviewCommits []string `json:"reviewcommits"`
}

type Repo struct {
	Provider      string   `json:"provider"`
	Url           string   `json:"url"`
	Branch        string   `json:"branch"`
	LastSHA1      string   `json:"lastSHA1"`
	Users         []User   `json:"users"`
	ReviewCommits []string `json:"reviewcommits"`
}

var repos Repos

func getConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, repos)
}

func getConfigPretty(c echo.Context) error {
	var temp_config Repos

	temp_copy_string, _ := json.Marshal(repos)
	json.Unmarshal(temp_copy_string, &temp_config)

	for i := 0; i < len(temp_config.Repos); i++ {
		temp_config.Repos[i].ReviewCommits = nil
		for t := 0; t < len(temp_config.Repos[i].Users); t++ {
			temp_config.Repos[i].Users[t].ReviewCommits = nil
		}
	}

	return c.JSONPretty(http.StatusOK, temp_config, "  ")
}

func getStatePretty(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, repos, "  ")
}

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func updateConfig(c echo.Context) error {
	var temp_config Repos
	re := regexp.MustCompile(`\r?\n`)
	new_config := []byte(re.ReplaceAllString(c.FormValue("json_config"), ""))
	if err := json.Unmarshal(new_config, &temp_config); err != nil {
		panic(err)
	}
	json.Unmarshal(new_config, &repos)
	update_repos()
	return c.JSON(http.StatusOK, "{'result': 'true'}")
}

func reviewedCommit(c echo.Context) error {
	id := c.Param("id")
	split := strings.Split(id, "_")
	id_repo, _ := strconv.Atoi(split[0])
	id_commit, _ := strconv.Atoi(split[1])

	parts := strings.Split(repos.Repos[id_repo].ReviewCommits[id_commit], "/")
	repos.Repos[id_repo].LastSHA1 = parts[len(parts)-1]
	slice := removeIndex(repos.Repos[id_repo].ReviewCommits, id_commit)
	repos.Repos[id_repo].ReviewCommits = slice
	save_json()
	return c.JSON(http.StatusOK, repos)
}

func save_json() {
	f, err := os.Create("watch.json")

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
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/review/:id", reviewedCommit)
	e.GET("/config", getConfig)
	e.GET("/pretty", getConfigPretty)
	e.GET("/state", getStatePretty)
	e.POST("/update", updateConfig)
	e.Logger.Fatal(e.Start(":1323"))
}
