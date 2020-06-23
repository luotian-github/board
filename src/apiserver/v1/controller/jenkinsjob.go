package controller

import (
	"fmt"
	c "git/inspursoft/board/src/apiserver/controllers/commons"
	"git/inspursoft/board/src/apiserver/service"
	"git/inspursoft/board/src/common/utils"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
)

const jenkinsBuildConsoleTemplateURL = "%s/job/{{.JobName}}/{{.BuildSerialID}}/consoleText"
const jenkinsStopBuildTemplateURL = "%s/job/{{.JobName}}/{{.BuildSerialID}}/stop"
const maxRetryCount = 300
const buildNumberCacheExpireSecond = time.Duration(maxRetryCount * time.Second)
const toggleBuildingCacheExpireSecond = time.Duration(maxRetryCount * time.Second)

type jobConsole struct {
	JobName       string `json:"job_name"`
	BuildSerialID string `json:"build_serial_id"`
}

type JenkinsJobCallbackController struct {
	c.BaseController
}

func (j *JenkinsJobCallbackController) Prepare() {
	j.EnableXSRF = false
}

func (j *JenkinsJobCallbackController) BuildNumberCallback() {
	userID := j.Ctx.Input.Param(":userID")
	buildNumber, _ := strconv.Atoi(j.Ctx.Input.Param(":buildNumber"))
	logs.Info("Get build number from Jenkins job callback: %d", buildNumber)
	c.MemoryCache.Put(userID+"_buildNumber", buildNumber, buildNumberCacheExpireSecond)
}

func (j JenkinsJobCallbackController) CustomPushEventPayload() {
	nodeSelection := utils.GetConfig("NODE_SELECTION")
	currentNode := nodeSelection()
	if currentNode == "" {
		currentNode = "slave1"
		logs.Warning("No Jenkins node selection found, will use default as: %s", currentNode)
	}
	data, err := ioutil.ReadAll(j.Ctx.Request.Body)
	if err != nil {
		j.InternalError(err)
	}
	logs.Debug("%s", string(data))
	service.CurrentDevOps().CustomHookPushPayload(data, currentNode)
}

type JenkinsJobController struct {
	c.BaseController
}

func (j *JenkinsJobController) getBuildNumber() (int, error) {
	if buildNumber, ok := c.MemoryCache.Get(strconv.Itoa(int(j.CurrentUser.ID)) + "_buildNumber").(int); ok {
		logs.Info("Get build number with key %s from cache is %d", strconv.Itoa(int(j.CurrentUser.ID))+"_lastBuildNumber", buildNumber)
		return buildNumber, nil
	}
	return 0, fmt.Errorf("cannot get build number from cache currently")
}

func (j *JenkinsJobController) clearBuildNumber() {
	key := strconv.Itoa(int(j.CurrentUser.ID)) + "_buildNumber"
	if c.MemoryCache.IsExist(key) {
		c.MemoryCache.Delete(key)
		logs.Info("Build number stored with key %s has been deleted from cache.", key)
	}
}

func (j *JenkinsJobController) toggleBuild(status bool) {
	logs.Info("Set building signal as %+v currently.", status)
	c.MemoryCache.Put(strconv.Itoa(int(j.CurrentUser.ID))+"_buildSignal", status, toggleBuildingCacheExpireSecond)
}

func (j *JenkinsJobController) getBuildSignal() bool {
	key := strconv.Itoa(int(j.CurrentUser.ID)) + "_buildSignal"
	if buildingSignal, ok := c.MemoryCache.Get(key).(bool); ok {
		return buildingSignal
	}
	return false
}

func (j *JenkinsJobController) Console() {
	j.clearBuildNumber()
	j.toggleBuild(true)
	getBuildNumberRetryCount := 0
	var buildNumber int
	var err error
	for true {
		buildNumber, err = j.getBuildNumber()
		if j.getBuildSignal() == false || getBuildNumberRetryCount >= maxRetryCount {
			logs.Debug("User canceled current process or exceeded max retry count, will exit.")
			return
		} else if err != nil {
			logs.Debug("Error occurred: %+v", err)
		} else {
			break
		}
		getBuildNumberRetryCount++
		time.Sleep(time.Second * 2)
	}

	jobName := j.GetString("job_name")

	if jobName == "" {
		j.CustomAbortAudit(http.StatusBadRequest, "No job name found.")
		return
	}

	repoName, err := service.ResolveRepoName(jobName, j.CurrentUser.Username)
	if err != nil {
		j.InternalError(err)
		return
	}

	query := jobConsole{JobName: repoName}
	query.BuildSerialID = strconv.Itoa(buildNumber)
	buildConsoleURL, err := utils.GenerateURL(fmt.Sprintf(jenkinsBuildConsoleTemplateURL, c.JenkinsBaseURL()), query)
	if err != nil {
		j.InternalError(err)
		return
	}
	logs.Debug("Requested Jenkins build console URL: %s", buildConsoleURL)
	ws, err := websocket.Upgrade(j.Ctx.ResponseWriter, j.Ctx.Request, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		j.CustomAbortAudit(http.StatusBadRequest, "Not a websocket handshake.")
		return
	} else if err != nil {
		j.CustomAbortAudit(http.StatusInternalServerError, "Cannot setup websocket connection.")
		return
	}
	defer ws.Close()

	req, err := http.NewRequest("GET", buildConsoleURL, nil)
	client := http.Client{}

	buffer := make(chan []byte, 1024)
	done := make(chan bool)

	retryCount := 0
	expiryTimer := time.NewTimer(time.Second * 900)
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for range ticker.C {
			resp, err := client.Do(req)
			if err != nil {
				j.InternalError(err)
				logs.Error("Failed to get console response: %+v", err)
				done <- true
				return
			}
			if resp.StatusCode == http.StatusNotFound {
				if retryCount >= maxRetryCount {
					done <- true
				} else {
					retryCount++
					if retryCount%50 == 0 {
						logs.Debug("Jenkins console is not ready at this moment, will retry for next %d request...", retryCount)
					}
					continue
				}
			}
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				j.InternalError(err)
				logs.Error("Failed to read data from response body: %+v", err)
				done <- true
				return
			}
			buffer <- data
			resp.Body.Close()
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "Finished:") {
					done <- true
				}
			}
		}
	}()

	for {
		select {
		case content := <-buffer:
			err = ws.WriteMessage(websocket.TextMessage, content)
			if err != nil {
				done <- true
			}
		case <-done:
			ticker.Stop()
			err = ws.Close()
			logs.Debug("WS is being closed.")
		case <-expiryTimer.C:
			ticker.Stop()
			err = ws.Close()
			logs.Debug("WS is being closed due to timeout.")
		}
		if err != nil {
			logs.Error("Failed to write message: %+v", err)
		}
	}
}

func (j *JenkinsJobController) Stop() {
	lastBuildNumber, err := j.getBuildNumber()
	if err != nil {
		logs.Error("Failed to get job number: %+v", err)
		j.toggleBuild(false)
		return
	}

	jobName := j.GetString("job_name")
	if jobName == "" {
		j.CustomAbortAudit(http.StatusBadRequest, "No job name found.")
		return
	}

	repoName, err := service.ResolveRepoName(jobName, j.CurrentUser.Username)
	if err != nil {
		j.InternalError(err)
		return
	}

	query := jobConsole{JobName: repoName}
	query.BuildSerialID = j.GetString("build_serial_id", strconv.Itoa(lastBuildNumber))
	stopBuildURL, err := utils.GenerateURL(fmt.Sprintf(jenkinsStopBuildTemplateURL, c.JenkinsBaseURL()), query)
	if err != nil {
		j.InternalError(err)
		return
	}
	logs.Debug("Requested stop Jenkins build URL: %s", stopBuildURL)
	resp, err := http.PostForm(stopBuildURL, nil)
	if err != nil {
		j.InternalError(err)
		return
	}
	err = service.ReleaseKVMRegistryByJobName(jobName)
	if err != nil {
		logs.Error("Failed to release KVM registry by job name: %s, error: %+v", jobName, err)
	}
	defer func() {
		resp.Body.Close()
		j.clearBuildNumber()
	}()
	logs.Debug("Response status of stopping Jenkins jobs: %d", resp.StatusCode)
	j.ServeStatus(resp.StatusCode, "")
}
