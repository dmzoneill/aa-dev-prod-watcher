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

func execute(app string, dir string, args []string) string {
	path, path_err := os.Getwd()
	if path_err != nil {
		log.Println(path_err)
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
			stream_print(app + " " + strings.Join(args, " "))
			stream_print(fmt.Sprint(err) + ": " + stderr.String())
			time.Sleep(1)
		} else {
			return out.String()
		}
	}

	return "nope"
}

func clone(name string, url string) {
	if _, err := os.Stat("repos/" + name); os.IsNotExist(err) {
		res := execute("/usr/bin/git", name, []string{"clone", url, "repos/" + name})
		if res == "nope" {
			stream_print("Problem cloning " + url + ", skipping ...")
			return
		}
		stream_print(res)
	}
}

func pull(name string) {
	res := strings.Trim(execute("/usr/bin/git", name, []string{"pull", "--rebase"}), "\n")
	if res == "nope" {
		stream_print("Problem pulling, skipping ...")
		return
	}
	if res != "Already up to date." {
		stream_print(res)
	}
}

func getLastSHA1(name string) string {
	res := execute("/usr/bin/git", name, []string{"rev-parse", "--verify", "HEAD"})
	if res == "nope" {
		stream_print("Problem with rev-parse on " + name + ", skipping ...")
		return res
	}
	return strings.TrimSuffix(res, "\n")
}

func getLastSHA1User(name string, user string) string {
	res := execute("/usr/bin/git", name, []string{"log", "-i", "--author", user, "-n", "1", "--pretty=format:\"%H\""})
	if res == "nope" {
		stream_print("Problem with log on " + name + ", skipping ...")
		return res
	}
	return strings.TrimSuffix(res, "\n")[1:41]
}

func getCommitTitle(name string, sha1Commit string) string {
	res := execute("/usr/bin/git", name, []string{"show", "--name-only", "--pretty=format:\"%ad || %s %d || %an\"", "--date=short", sha1Commit})
	if res == "nope" {
		stream_print("Problem show on " + name + ", skipping ...")
		return res
	}
	review_title := strings.Split(res, "\n")[0]
	review_title = review_title[1 : len(review_title)-1]
	return review_title
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

func print_commits(index int, name string, url string, sha1 string, provider string) {
	res := execute("/usr/bin/git", name, []string{"rev-list", sha1 + "..HEAD"})
	if res == "nope" {
		stream_print("Problem with rev-list on " + url + ", skipping ...")
		return
	}

	commits := strings.Split(res, "\n")
	if len(commits) > 0 {
		stream_print("")
	}
	for i := 0; i < len(commits)-1; i++ {
		if isCommitInCommits(string(commits[i]), repos.Repos[index].ReviewCommits) {
			stream_print("     .. dupe")
			continue
		}
		if provider == "github" {
			review := getCommitTitle(name, string(commits[i])) + ",#," + fmt.Sprintf("%s/commit/%s", url, string(commits[i]))
			repos.Repos[index].ReviewCommits = append(repos.Repos[index].ReviewCommits, review)
			stream_print(fmt.Sprintf("     >> %s/commit/%s", url, commits[i]))
		} else {
			review := getCommitTitle(name, string(commits[i])) + ",#," + fmt.Sprintf("%s/-/commit/%s", url, string(commits[i]))
			repos.Repos[index].ReviewCommits = append(repos.Repos[index].ReviewCommits, review)
			stream_print(fmt.Sprintf("     >> %s/-/commit/%s", url, commits[i]))
		}
	}
	if len(commits) > 0 {
		stream_print("")
	}
	repos.Repos[index].ReviewCommits = removeDuplicateStr(repos.Repos[index].ReviewCommits)
}

func print_commits_user(index int, user_index int, name string, url string, user string, sha1 string, provider string) {
	res := execute("/usr/bin/git", name, []string{"log", "-i", "--author", user, "-n", "100", "--pretty=format:\"%H\"", sha1 + "..HEAD"})
	if res == "nope" {
		stream_print("Problem with log on " + url + ", skipping ...")
		return
	}

	commits := strings.Split(res, "\n")
	if len(commits) > 0 {
		stream_print("")
	}
	for i := 0; i < len(commits)-1; i++ {
		if isCommitInCommits(string(commits[i])[1:41], repos.Repos[index].ReviewCommits) {
			stream_print("     .. dupe")
			continue
		}
		if provider == "github" {
			review := getCommitTitle(name, string(commits[i])[1:41]) + ",#," + fmt.Sprintf("%s/commit/%s", url, string(commits[i])[1:41])
			repos.Repos[index].Users[user_index].ReviewCommits = append(repos.Repos[index].Users[user_index].ReviewCommits, review)
			stream_print(fmt.Sprintf("     >> %s/commit/%s", url, string(commits[i])[1:41]))
		} else {
			review := getCommitTitle(name, string(commits[i])[1:41]) + ",#," + fmt.Sprintf("%s/-/commit/%s", url, string(commits[i])[1:41])
			repos.Repos[index].Users[user_index].ReviewCommits = append(repos.Repos[index].Users[user_index].ReviewCommits, review)
			stream_print(fmt.Sprintf("     >> %s/-/commit/%s", url, string(commits[i])[1:41]))
		}
	}
	if len(commits) > 0 {
		stream_print("")
	}
	repos.Repos[index].Users[user_index].ReviewCommits = removeDuplicateStr(repos.Repos[index].Users[user_index].ReviewCommits)
}

func update_repos() {

	stream_print("")
	stream_print(" " + strings.Repeat("=", 80))
	stream_print(" Updating.... ")
	stream_print(" " + strings.Repeat("=", 80))
	stream_print("")

	for i := 0; i < len(repos.Repos); i++ {
		parts := strings.Split(repos.Repos[i].Url, "/")
		name := parts[len(parts)-1]

		clone(name, repos.Repos[i].Url)
		pull(name)

		if len(repos.Repos[i].Users) == 0 {
			last_commit := getLastSHA1(name)
			if last_commit == "nope" {
				continue
			}
			if repos.Repos[i].LastSHA1 == "" {
				repos.Repos[i].LastSHA1 = last_commit
			}
			stream_print(fmt.Sprintf(" %2s) %-7s %-7s %-9s %-41s %s", strconv.Itoa(i), repos.Repos[i].Provider, repos.Repos[i].Branch, "", repos.Repos[i].LastSHA1, repos.Repos[i].Url))
			if last_commit != repos.Repos[i].LastSHA1 {
				print_commits(i, name, repos.Repos[i].Url, repos.Repos[i].LastSHA1, repos.Repos[i].Provider)
			}
		}

		for t := 0; t < len(repos.Repos[i].Users); t++ {
			last_commit := getLastSHA1User(name, repos.Repos[i].Users[t].Username)
			if last_commit == "nope" {
				continue
			}
			if repos.Repos[i].Users[t].LastSHA1 == "" {
				repos.Repos[i].Users[t].LastSHA1 = last_commit
			}
			stream_print(fmt.Sprintf(" %2s) %-7s %-7s %-9s %-41s %s", strconv.Itoa(i), repos.Repos[i].Provider, repos.Repos[i].Branch, repos.Repos[i].Users[t].Username, repos.Repos[i].Users[t].LastSHA1, repos.Repos[i].Url))
			if last_commit != repos.Repos[i].Users[t].LastSHA1 {
				print_commits_user(i, t, name, repos.Repos[i].Url, repos.Repos[i].Users[t].Username, repos.Repos[i].Users[t].LastSHA1, repos.Repos[i].Provider)
			}
		}

		if _, err := os.Stat("repos"); os.IsNotExist(err) {
			if err := os.Mkdir("repos", os.ModePerm); err != nil {
				log.Fatal(err)
			}
		}
	}
	save_yaml()
}
