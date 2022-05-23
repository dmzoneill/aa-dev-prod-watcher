package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

type Repos struct {
	Repos []Repo `yaml:"repos"`
}

type User struct {
	Username      string   `yaml:"user"`
	LastSHA1      string   `yaml:"lastSHA1"`
	ReviewCommits []string `yaml:"reviewcommits"`
}

type Repo struct {
	Provider      string   `yaml:"provider"`
	Url           string   `yaml:"url"`
	Branch        string   `yaml:"branch"`
	LastSHA1      string   `yaml:"lastSHA1"`
	Users         []User   `yaml:"users"`
	ReviewCommits []string `yaml:"reviewcommits"`
}

var repos Repos
var streamBuffer []string

func getConfig(c echo.Context) error {
	data, _ := yaml.Marshal(repos)
	return c.String(http.StatusOK, string(data))
}

func getSSHKeys(c echo.Context) error {
	dirname, _ := os.UserHomeDir()
	path := dirname + "/.ssh"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
		if err == nil {
			return nil
		}
	}

	key_path := path + "/id_keys"
	if _, err := os.Stat(key_path); os.IsNotExist(err) {
		f, err := os.OpenFile(key_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		f.Close()
	}

	data, err := os.Open(key_path)
	if err != nil {
		stream_print(err.Error())
	}
	defer data.Close()
	byteValue, _ := ioutil.ReadAll(data)

	os.Chmod(key_path, 0600)

	return c.String(http.StatusOK, string(byteValue))
}

func setSSHKeys(c echo.Context) error {
	dirname, _ := os.UserHomeDir()
	path := dirname + "/.ssh"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
		if err == nil {
			return nil
		}
	}

	key_path := path + "/id_keys"
	f, err := os.OpenFile(key_path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	new_config := []byte(c.FormValue("ssh_keys"))
	f.Write(new_config)
	f.Close()
	os.Chmod(key_path, 0600)

	return c.String(http.StatusOK, string(new_config))
}

func getGitConfig(c echo.Context) error {
	dirname, _ := os.UserHomeDir()
	path := dirname + "/.gitconfig"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		f.Close()
	}

	data, err := os.Open(path)
	if err != nil {
		stream_print(err.Error())
	}
	defer data.Close()
	byteValue, _ := ioutil.ReadAll(data)
	return c.String(http.StatusOK, string(byteValue))
}

func setGitConfig(c echo.Context) error {
	dirname, _ := os.UserHomeDir()
	path := dirname + "/.gitconfig"

	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	new_config := []byte(c.FormValue("git_config"))
	f.Write(new_config)
	f.Close()

	return c.String(http.StatusOK, string(new_config))
}

func getConfigPretty(c echo.Context) error {
	var temp_config Repos

	temp_copy_string, _ := yaml.Marshal(repos)
	yaml.Unmarshal(temp_copy_string, &temp_config)

	for i := 0; i < len(temp_config.Repos); i++ {
		temp_config.Repos[i].ReviewCommits = nil
		for t := 0; t < len(temp_config.Repos[i].Users); t++ {
			temp_config.Repos[i].Users[t].ReviewCommits = nil
		}
	}
	temp_copy_string, _ = yaml.Marshal(temp_config)
	return c.String(http.StatusOK, string(temp_copy_string))
}

func getServerLog(c echo.Context) error {
	if len(streamBuffer) == 0 {
		return c.JSON(http.StatusOK, streamBuffer)
	}

	if len(streamBuffer) > 1000 {
		streamBuffer = streamBuffer[len(streamBuffer)-1000:]
	}

	return c.JSON(http.StatusOK, streamBuffer)
}

func getStatePretty(c echo.Context) error {
	data, _ := yaml.Marshal(repos)
	return c.String(http.StatusOK, string(data))
}

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func updateConfig(c echo.Context) error {
	var temp_config Repos
	if err := yaml.Unmarshal([]byte(c.FormValue("yaml_config")), &temp_config); err != nil {
		panic(err)
	}
	yaml.Unmarshal([]byte(c.FormValue("yaml_config")), &repos)
	data, _ := yaml.Marshal(repos)
	return c.String(http.StatusOK, string(data))
}

func addRepoConfig(c echo.Context) error {
	var new_repo Repo
	new_repo.Branch = c.FormValue("branch")
	new_repo.LastSHA1 = c.FormValue("sha1")
	new_repo.Provider = c.FormValue("provider")
	new_repo.Url = c.FormValue("url")
	repos.Repos = append(repos.Repos, new_repo)
	update_repos()
	data, _ := yaml.Marshal(repos)
	return c.String(http.StatusOK, string(data))
}

func editRepoConfig(c echo.Context) error {
	for index, v := range repos.Repos {
		if v.Url == c.FormValue("url") {
			repos.Repos[index].Branch = c.FormValue("branch")
			repos.Repos[index].LastSHA1 = c.FormValue("sha1")
			repos.Repos[index].Provider = c.FormValue("provider")
			break
		}
	}
	data, _ := yaml.Marshal(repos)
	return c.String(http.StatusOK, string(data))
}

func deleteRepoConfig(c echo.Context) error {
	for index, v := range repos.Repos {
		if v.Url == c.FormValue("url") {
			repos.Repos[index] = repos.Repos[len(repos.Repos)-1]
			repos.Repos = repos.Repos[:len(repos.Repos)-1]
			break
		}
	}
	parts := strings.Split(c.FormValue("url"), "/")
	name := parts[len(parts)-1]
	os.RemoveAll("repos/" + name)
	data, _ := yaml.Marshal(repos)
	return c.String(http.StatusOK, string(data))
}

func reviewedCommit(c echo.Context) error {
	id := c.Param("id")

	if !strings.Contains(id, "_") {
		temp_copy_string, _ := yaml.Marshal(repos)
		return c.String(http.StatusOK, string(temp_copy_string))
	}

	split := strings.Split(id, "_")
	id_repo := 0
	id_user := 0
	id_commit := 0

	if len(split) > 2 {
		id_repo, _ = strconv.Atoi(split[0])
		id_user, _ = strconv.Atoi(split[1])
		id_commit, _ = strconv.Atoi(split[2])
		parts := strings.Split(repos.Repos[id_repo].Users[id_user].ReviewCommits[id_commit], "/")
		repos.Repos[id_repo].Users[id_user].LastSHA1 = parts[len(parts)-1]
		slice := removeIndex(repos.Repos[id_repo].Users[id_user].ReviewCommits, id_commit)
		repos.Repos[id_repo].Users[id_user].ReviewCommits = slice
	} else {
		id_repo, _ = strconv.Atoi(split[0])
		id_commit, _ = strconv.Atoi(split[1])
		parts := strings.Split(repos.Repos[id_repo].ReviewCommits[id_commit], "/")
		repos.Repos[id_repo].LastSHA1 = parts[len(parts)-1]
		slice := removeIndex(repos.Repos[id_repo].ReviewCommits, id_commit)
		repos.Repos[id_repo].ReviewCommits = slice
	}

	save_yaml()
	temp_copy_string, _ := yaml.Marshal(repos)
	return c.String(http.StatusOK, string(temp_copy_string))
}

func save_yaml() {
	f, err := os.Create("watch.yaml")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	// Convert structs to yaml.
	data, err := yaml.Marshal(repos)
	if err != nil {
		log.Fatal(err)
	}

	_, err2 := f.Write(data)

	if err2 != nil {
		log.Fatal(err2)
	}
}

func stream_print(line string) {
	dt := time.Now()
	streamBuffer = append(streamBuffer, dt.Format("01-02-2006 15:04:05")+": "+line)
	if len(streamBuffer) > 1000 {
		streamBuffer = streamBuffer[len(streamBuffer)-1000:]
	}
	fmt.Println(line)
}

func main() {

	yamlFile, err := os.Open("watch.yaml")
	if err != nil {
		stream_print(err.Error())
	}
	defer yamlFile.Close()
	byteValue, _ := ioutil.ReadAll(yamlFile)
	yaml.Unmarshal(byteValue, &repos)

	update_repos()

	c := cron.New(cron.WithSeconds())
	c.AddFunc("0 */5 * * * *", func() { update_repos() })
	c.Start()

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/review/:id", reviewedCommit)
	e.GET("/config", getConfig)
	e.GET("/gitconfig", getGitConfig)
	e.POST("/gitconfig", setGitConfig)
	e.GET("/sshkeys", getSSHKeys)
	e.POST("/sshkeys", setSSHKeys)
	e.GET("/pretty", getConfigPretty)
	e.GET("/state", getStatePretty)
	e.GET("/serverlog", getServerLog)
	e.POST("/update", updateConfig)
	e.POST("/add", addRepoConfig)
	e.POST("/edit", editRepoConfig)
	e.DELETE("/delete", deleteRepoConfig)
	e.Logger.Fatal(e.Start(":1323"))
}
