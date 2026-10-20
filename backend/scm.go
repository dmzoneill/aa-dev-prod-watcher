package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const gitCommandFailed = "nope"

var updateRunning bool

func execute(app string, dir string, args []string) string {
	path, pathErr := os.Getwd()
	if pathErr != nil {
		log.Println(pathErr)
		os.Exit(2)
	}

	for i := 0; i < 2; i++ {
		cmd := exec.Command(app, args...)

		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		if args[0] != "clone" {
			cmd.Dir = path + "/repos/" + dir
		} else {
			cmd.Dir = path
		}
		err := cmd.Run()

		if err != nil {
			streamPrint(app + " " + strings.Join(args, " "))
			streamPrint(fmt.Sprint(err) + ": " + stderr.String())
			time.Sleep(time.Second)
		} else {
			return out.String()
		}
	}

	return gitCommandFailed
}

func clone(name string, url string) {
	if _, err := os.Stat("repos/" + name); os.IsNotExist(err) {
		res := execute("/usr/bin/git", name, []string{"clone", url, "repos/" + name})
		if res == gitCommandFailed {
			streamPrint("Problem cloning " + url + ", skipping ...")
			return
		}
		streamPrint(res)
	}
}

func pull(name string) {
	res := strings.Trim(execute("/usr/bin/git", name, []string{"pull", "--rebase"}), "\n")
	if res == gitCommandFailed {
		streamPrint("Problem pulling, skipping ...")
		return
	}
	if res != "Already up to date." {
		streamPrint(res)
	}
}

func getLastSHA1(name string) string {
	res := execute("/usr/bin/git", name, []string{"rev-parse", "--verify", "HEAD"})
	if res == gitCommandFailed {
		streamPrint("Problem with rev-parse on " + name + ", skipping ...")
		return res
	}
	return strings.TrimSuffix(res, "\n")
}

func getLastSHA1User(name string, user string) string {
	res := execute("/usr/bin/git", name, []string{"log", "-i", "--author", user, "-n", "1", "--pretty=format:\"%H\""})
	if res == gitCommandFailed {
		streamPrint("Problem with log on " + name + ", skipping ...")
		return res
	}
	return strings.TrimSuffix(res, "\n")[1:41]
}

func getCommitTitle(name string, sha1Commit string) string {
	res := execute("/usr/bin/git", name, []string{"show", "--name-only", "--pretty=format:\"%ad || %s %d || %an\"", "--date=short", sha1Commit})
	if res == gitCommandFailed {
		streamPrint("Problem show on " + name + ", skipping ...")
		return res
	}
	reviewTitle := strings.Split(res, "\n")[0]
	reviewTitle = reviewTitle[1 : len(reviewTitle)-1]
	return reviewTitle
}

func isCommitInCommits(sha1 string, commits []string) bool {
	for i := 0; i < len(commits)-1; i++ {
		if strings.Contains(commits[i], sha1) {
			return true
		}
	}
	return false
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func commitLink(provider string, url string, sha string) string {
	if provider == "github" {
		return fmt.Sprintf("%s/commit/%s", url, sha)
	}
	return fmt.Sprintf("%s/-/commit/%s", url, sha)
}

func printCommits(index int, name string, url string, sha1 string, provider string) {
	res := execute("/usr/bin/git", name, []string{"rev-list", sha1 + "..HEAD"})
	if res == gitCommandFailed {
		streamPrint("Problem with rev-list on " + url + ", skipping ...")
		return
	}

	commits := strings.Split(res, "\n")
	if len(commits) > 0 {
		streamPrint("")
	}
	for i := 0; i < len(commits)-1; i++ {
		commit := commits[i]
		if isCommitInCommits(commit, repos.Repos[index].ReviewCommits) {
			continue
		}
		link := commitLink(provider, url, commit)
		review := getCommitTitle(name, commit) + ",#," + link
		repos.Repos[index].ReviewCommits = append(repos.Repos[index].ReviewCommits, review)
		streamPrint("     >> " + link)
	}
	if len(commits) > 0 {
		streamPrint("")
	}
	repos.Repos[index].ReviewCommits = removeDuplicateStr(repos.Repos[index].ReviewCommits)
}

func printCommitsUser(index int, userIndex int, name string, url string, user string, sha1 string, provider string) {
	res := execute("/usr/bin/git", name, []string{"log", "-i", "--author", user, "-n", "100", "--pretty=format:\"%H\"", sha1 + "..HEAD"})
	if res == gitCommandFailed {
		streamPrint("Problem with log on " + url + ", skipping ...")
		return
	}

	commits := strings.Split(res, "\n")
	if len(commits) > 0 {
		streamPrint("")
	}
	for i := 0; i < len(commits)-1; i++ {
		sha := commits[i][1:41]
		if isCommitInCommits(sha, repos.Repos[index].Users[userIndex].ReviewCommits) {
			continue
		}
		link := commitLink(provider, url, sha)
		review := getCommitTitle(name, sha) + ",#," + link
		repos.Repos[index].Users[userIndex].ReviewCommits = append(repos.Repos[index].Users[userIndex].ReviewCommits, review)
		streamPrint("     >> " + link)
	}
	if len(commits) > 0 {
		streamPrint("")
	}
	repos.Repos[index].Users[userIndex].ReviewCommits = removeDuplicateStr(repos.Repos[index].Users[userIndex].ReviewCommits)
}

func updateRepos() {
	if updateRunning {
		return
	}

	updateRunning = true

	streamPrint("")
	streamPrint(" " + strings.Repeat("=", 80))
	streamPrint(" Updating.... ")
	streamPrint(" " + strings.Repeat("=", 80))
	streamPrint("")

	for i := 0; i < len(repos.Repos); i++ {
		parts := strings.Split(repos.Repos[i].URL, "/")
		name := parts[len(parts)-1]

		clone(name, repos.Repos[i].URL)
		pull(name)

		if len(repos.Repos[i].Users) == 0 {
			lastCommit := getLastSHA1(name)
			if lastCommit == gitCommandFailed {
				continue
			}
			if repos.Repos[i].LastSHA1 == "" {
				repos.Repos[i].LastSHA1 = lastCommit
			}
			streamPrint(fmt.Sprintf(" %2s) %-7s %-7s %-9s %-41s %s", strconv.Itoa(i), repos.Repos[i].Provider, repos.Repos[i].Branch, "", repos.Repos[i].LastSHA1, repos.Repos[i].URL))
			if lastCommit != repos.Repos[i].LastSHA1 {
				printCommits(i, name, repos.Repos[i].URL, repos.Repos[i].LastSHA1, repos.Repos[i].Provider)
			}
		}

		for t := 0; t < len(repos.Repos[i].Users); t++ {
			lastCommit := getLastSHA1User(name, repos.Repos[i].Users[t].Username)
			if lastCommit == gitCommandFailed {
				continue
			}
			if repos.Repos[i].Users[t].LastSHA1 == "" {
				repos.Repos[i].Users[t].LastSHA1 = lastCommit
			}
			streamPrint(fmt.Sprintf(" %2s) %-7s %-7s %-9s %-41s %s", strconv.Itoa(i), repos.Repos[i].Provider, repos.Repos[i].Branch, repos.Repos[i].Users[t].Username, repos.Repos[i].Users[t].LastSHA1, repos.Repos[i].URL))
			if lastCommit != repos.Repos[i].Users[t].LastSHA1 {
				printCommitsUser(i, t, name, repos.Repos[i].URL, repos.Repos[i].Users[t].Username, repos.Repos[i].Users[t].LastSHA1, repos.Repos[i].Provider)
			}
		}

		if _, err := os.Stat("repos"); os.IsNotExist(err) {
			if err := os.Mkdir("repos", os.ModePerm); err != nil {
				log.Fatal(err)
			}
		}
	}
	saveYaml()
	updateRunning = false
}
