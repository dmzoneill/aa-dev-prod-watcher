package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
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

func clone(name string, url string) {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	app := "/usr/bin/git"
	arg0 := "clone"
	arg1 := url
	arg2 := path + "/repos/" + name

	// fmt.Println(app + " " + arg0 + " " + arg1 + " " + arg2)
	cmd := exec.Command(app, arg0, arg1, arg2)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err1 := cmd.Run()
	if err1 != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		os.Exit(47)
	}
}

func pull(name string) {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	app := "/usr/bin/git"
	arg0 := "pull"

	// fmt.Println(app + " " + arg0)
	cmd := exec.Command(app, arg0)
	cmd.Dir = path + "/repos/" + name

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err1 := cmd.Run()

	if err1 != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		os.Exit(37)
	}
}

func update_repos() {
	jsonFile, err := os.Open("watch.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var repos Repos
	json.Unmarshal(byteValue, &repos)

	for i := 0; i < len(repos.Repos); i++ {
		parts := strings.Split(repos.Repos[i].Url, "/")

		if len(repos.Repos[i].Users) == 0 {
			fmt.Printf("Checking: %-7s %-8s %-8s %-45s %s \n", repos.Repos[i].Provider, repos.Repos[i].Branch, "", repos.Repos[i].LastSHA1, repos.Repos[i].Url)
		}

		for t := 0; t < len(repos.Repos[i].Users); t++ {
			fmt.Printf("Checking: %-7s %-8s %-8s %-45s %s \n", repos.Repos[i].Provider, repos.Repos[i].Branch, repos.Repos[i].Users[t].Username, repos.Repos[i].Users[t].LastSHA1, repos.Repos[i].Url)
		}

		if _, err := os.Stat("repos"); os.IsNotExist(err) {
			if err := os.Mkdir("repos", os.ModePerm); err != nil {
				log.Fatal(err)
			}
		}

		if _, err := os.Stat("repos/" + parts[len(parts)-1]); os.IsNotExist(err) {
			clone(parts[len(parts)-1], repos.Repos[i].Url)
			if _, err := os.Stat("repos/" + parts[len(parts)-1]); os.IsNotExist(err) {
				os.Exit(27)
			}
		}

		pull(parts[len(parts)-1])
	}

	// Convert structs to JSON.
	data, err := json.Marshal(repos)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", data)
}
