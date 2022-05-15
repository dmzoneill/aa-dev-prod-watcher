package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func execute(app string, dir string, args []string) string {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	cmd := exec.Command(app, args...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Dir = path + "/repos/" + dir
	err1 := cmd.Run()

	if err != nil {
		fmt.Println(fmt.Sprint(err1) + ": " + stderr.String())
		os.Exit(1)
	}

	return out.String()
}

func clone(name string, url string) {
	execute("/usr/bin/git", name, []string{"clone", url})
}

func pull(name string) {
	execute("/usr/bin/git", name, []string{"pull"})
}

func getLastSHA1(name string) string {
	return strings.TrimSuffix(execute("/usr/bin/git", name, []string{"rev-parse", "--verify", "HEAD"}), "\n")
}

func getLastSHA1User(name string, user string) string {
	return strings.TrimSuffix(execute("/usr/bin/git", name, []string{"log", "-i", "--author" + user, "-n", "1", "--pretty=format:\"%H\""}), "\n")[1:41]
}

func print_commits(name string, url string, sha1 string, provider string) {
	commits := strings.Split(execute("/usr/bin/git", name, []string{"rev-list", sha1 + "..HEAD"}), "\n")
	if len(commits) > 0 {
		fmt.Printf("\n")
	}
	for i := 0; i < len(commits)-1; i++ {
		if provider == "github" {
			fmt.Printf("     >> %s/commit/%s\n", url, commits[i])
		} else {
			fmt.Printf("     >> %s/-/commit/%s\n", url, commits[i])
		}
	}
	if len(commits) > 0 {
		fmt.Printf("\n")
	}
}

func update_repos() {

	for i := 0; i < len(repos.Repos); i++ {
		parts := strings.Split(repos.Repos[i].Url, "/")
		name := parts[len(parts)-1]

		if _, err := os.Stat("repos/" + name); os.IsExist(err) {
			pull(name)
		}

		if len(repos.Repos[i].Users) == 0 {
			last_commit := getLastSHA1(name)
			if repos.Repos[i].LastSHA1 == "" {
				repos.Repos[i].LastSHA1 = last_commit
			}
			fmt.Printf(" %2s) %-7s %-7s %-9s %-41s %s \n", strconv.Itoa(i), repos.Repos[i].Provider, repos.Repos[i].Branch, "", repos.Repos[i].LastSHA1, repos.Repos[i].Url)
			if last_commit != repos.Repos[i].LastSHA1 {
				print_commits(name, repos.Repos[i].Url, repos.Repos[i].LastSHA1, repos.Repos[i].Provider)
			}
		}

		for t := 0; t < len(repos.Repos[i].Users); t++ {
			if repos.Repos[i].Users[t].LastSHA1 == "" {
				repos.Repos[i].Users[t].LastSHA1 = getLastSHA1User(name, repos.Repos[i].Users[t].Username)
			}
			fmt.Printf(" %2s) %-7s %-7s %-9s %-41s %s \n", strconv.Itoa(i), repos.Repos[i].Provider, repos.Repos[i].Branch, repos.Repos[i].Users[t].Username, repos.Repos[i].Users[t].LastSHA1, repos.Repos[i].Url)
		}

		if _, err := os.Stat("repos"); os.IsNotExist(err) {
			if err := os.Mkdir("repos", os.ModePerm); err != nil {
				log.Fatal(err)
			}
		}

		if _, err := os.Stat("repos/" + name); os.IsNotExist(err) {
			clone(name, repos.Repos[i].Url)
			if _, err := os.Stat("repos/" + name); os.IsNotExist(err) {
				os.Exit(3)
			}
		}

	}

	save_json()
}
