package main

import (
	"fmt"
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
	URL           string   `yaml:"url"`
	Branch        string   `yaml:"branch"`
	LastSHA1      string   `yaml:"lastSHA1"`
	Users         []User   `yaml:"users"`
	ReviewCommits []string `yaml:"reviewcommits"`
}

var repos Repos
var streamBuffer []string
var updateQueued bool

func getConfig(c echo.Context) error {
	return writeReposResponse(c)
}

func sshKeyPath() (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := dirname + "/.ssh"
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		if mkErr := os.MkdirAll(path, os.ModePerm); mkErr != nil {
			return "", mkErr
		}
	} else if statErr != nil {
		return "", statErr
	}

	return path + "/id_keys", nil
}

func gitConfigPath() (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return dirname + "/.gitconfig", nil
}

func ensureFile(path string) error {
	if _, statErr := os.Stat(path); statErr == nil {
		return nil
	} else if !os.IsNotExist(statErr) {
		return statErr
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}

func getSSHKeys(c echo.Context) error {
	keyPath, err := sshKeyPath()
	if err != nil {
		return err
	}
	if err = ensureFile(keyPath); err != nil {
		return err
	}

	byteValue, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	if err := os.Chmod(keyPath, 0o600); err != nil {
		return err
	}

	return c.String(http.StatusOK, string(byteValue))
}

func setSSHKeys(c echo.Context) error {
	keyPath, err := sshKeyPath()
	if err != nil {
		return err
	}

	newConfig := []byte(c.FormValue("ssh_keys"))
	f, err := os.OpenFile(keyPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(newConfig)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Chmod(keyPath, 0o600); err != nil {
		return err
	}
	updateQueued = true

	return c.String(http.StatusOK, string(newConfig))
}

func getGitConfig(c echo.Context) error {
	path, err := gitConfigPath()
	if err != nil {
		return err
	}
	if err = ensureFile(path); err != nil {
		return err
	}

	byteValue, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, string(byteValue))
}

func setGitConfig(c echo.Context) error {
	path, err := gitConfigPath()
	if err != nil {
		return err
	}

	newConfig := []byte(c.FormValue("git_config"))
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(newConfig)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	updateQueued = true
	saveYaml()
	return c.String(http.StatusOK, string(newConfig))
}

func getConfigPretty(c echo.Context) error {
	tempCopy, err := yaml.Marshal(repos)
	if err != nil {
		return err
	}

	var tempConfig Repos
	if err = yaml.Unmarshal(tempCopy, &tempConfig); err != nil {
		return err
	}

	for i := 0; i < len(tempConfig.Repos); i++ {
		tempConfig.Repos[i].ReviewCommits = nil
		for t := 0; t < len(tempConfig.Repos[i].Users); t++ {
			tempConfig.Repos[i].Users[t].ReviewCommits = nil
		}
	}
	pretty, err := yaml.Marshal(tempConfig)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, string(pretty))
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
	return writeReposResponse(c)
}

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func updateConfig(c echo.Context) error {
	if err := yaml.Unmarshal([]byte(c.FormValue("yaml_config")), &repos); err != nil {
		return err
	}
	updateQueued = true
	saveYaml()
	return writeReposResponse(c)
}

func addRepoConfig(c echo.Context) error {
	newRepo := Repo{
		Branch:   c.FormValue("branch"),
		LastSHA1: c.FormValue("sha1"),
		Provider: c.FormValue("provider"),
		URL:      c.FormValue("url"),
	}
	repos.Repos = append(repos.Repos, newRepo)
	updateQueued = true
	saveYaml()
	return writeReposResponse(c)
}

func editRepoConfig(c echo.Context) error {
	for index, v := range repos.Repos {
		if v.URL == c.FormValue("url") {
			repos.Repos[index].Branch = c.FormValue("branch")
			repos.Repos[index].LastSHA1 = c.FormValue("sha1")
			repos.Repos[index].Provider = c.FormValue("provider")
			break
		}
	}
	updateQueued = true
	saveYaml()
	return writeReposResponse(c)
}

func deleteRepoConfig(c echo.Context) error {
	for index, v := range repos.Repos {
		if v.URL == c.FormValue("url") {
			repos.Repos[index] = repos.Repos[len(repos.Repos)-1]
			repos.Repos = repos.Repos[:len(repos.Repos)-1]
			break
		}
	}
	parts := strings.Split(c.FormValue("url"), "/")
	name := parts[len(parts)-1]
	if err := os.RemoveAll("repos/" + name); err != nil {
		return err
	}
	updateQueued = true
	saveYaml()
	return writeReposResponse(c)
}

func reviewedCommit(c echo.Context) error {
	id := c.Param("id")

	if !strings.Contains(id, "_") {
		return writeReposResponse(c)
	}

	split := strings.Split(id, "_")
	if len(split) > 2 {
		idRepo, err := strconv.Atoi(split[0])
		if err != nil {
			return err
		}
		idUser, err := strconv.Atoi(split[1])
		if err != nil {
			return err
		}
		idCommit, err := strconv.Atoi(split[2])
		if err != nil {
			return err
		}
		parts := strings.Split(repos.Repos[idRepo].Users[idUser].ReviewCommits[idCommit], "/")
		repos.Repos[idRepo].Users[idUser].LastSHA1 = parts[len(parts)-1]
		slice := removeIndex(repos.Repos[idRepo].Users[idUser].ReviewCommits, idCommit)
		repos.Repos[idRepo].Users[idUser].ReviewCommits = slice
	} else {
		idRepo, err := strconv.Atoi(split[0])
		if err != nil {
			return err
		}
		idCommit, err := strconv.Atoi(split[1])
		if err != nil {
			return err
		}
		parts := strings.Split(repos.Repos[idRepo].ReviewCommits[idCommit], "/")
		repos.Repos[idRepo].LastSHA1 = parts[len(parts)-1]
		slice := removeIndex(repos.Repos[idRepo].ReviewCommits, idCommit)
		repos.Repos[idRepo].ReviewCommits = slice
	}

	updateQueued = true
	saveYaml()
	return writeReposResponse(c)
}

func writeReposResponse(c echo.Context) error {
	data, err := yaml.Marshal(repos)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, string(data))
}

func saveYaml() {
	data, err := yaml.Marshal(repos)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("watch.yaml")
	if err != nil {
		log.Fatal(err)
	}

	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		log.Fatal(writeErr)
	}
	if closeErr != nil {
		log.Fatal(closeErr)
	}
}

func streamPrint(line string) {
	dt := time.Now()
	streamBuffer = append(streamBuffer, dt.Format("01-02-2006 15:04:05")+": "+line)
	if len(streamBuffer) > 1000 {
		streamBuffer = streamBuffer[len(streamBuffer)-1000:]
	}
	fmt.Println(line)
}

func pendingUpdate() {
	if updateQueued {
		updateQueued = false
		updateRepos()
	}
}

func loadRepos() error {
	byteValue, err := os.ReadFile("watch.yaml")
	if err != nil {
		return err
	}
	return yaml.Unmarshal(byteValue, &repos)
}

func requestLogger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogLatency:       true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURI:           true,
		LogRequestID:     true,
		LogUserAgent:     true,
		LogStatus:        true,
		LogError:         true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				log.Printf("method=%s uri=%s status=%d latency=%s error=%v", v.Method, v.URI, v.Status, v.Latency, v.Error)
				return nil
			}
			log.Printf("method=%s uri=%s status=%d latency=%s", v.Method, v.URI, v.Status, v.Latency)
			return nil
		},
	})
}

func main() {
	if err := loadRepos(); err != nil {
		log.Fatal(err)
	}

	updateRepos()

	cronRunner := cron.New(cron.WithSeconds())
	if _, err := cronRunner.AddFunc("0 */5 * * * *", func() { updateRepos() }); err != nil {
		log.Fatal(err)
	}
	if _, err := cronRunner.AddFunc("*/1 * * * * *", func() { pendingUpdate() }); err != nil {
		log.Fatal(err)
	}
	cronRunner.Start()

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(requestLogger())
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
