package jenkins

import (
	"fmt"
	"git/inspursoft/board/src/common/utils"
	"net/http"
	"time"

	"github.com/astaxie/beego/logs"
)

var jenkinsBaseURL = utils.GetConfig("JENKINS_BASE_URL")
var maxRetryCount = 120

type jenkinsHandler struct{}

func NewJenkinsHandler() *jenkinsHandler {
	for i := 0; i < maxRetryCount; i++ {
		logs.Debug("Ping Jenkins server %d time(s)...", i+1)
		resp, err := utils.RequestHandle(http.MethodGet, fmt.Sprintf("%s/job/base", jenkinsBaseURL()), nil, nil)
		if err != nil {
			logs.Error("Failed to request Jenkins server: %+v", err)
		}
		if resp != nil {
			if resp.StatusCode <= 400 {
				break
			}
		} else if i == maxRetryCount-1 {
			logs.Warn("Failed to ping Gogits due to exceed max retry count.")
		}
		time.Sleep(time.Second)
	}
	return &jenkinsHandler{}
}

func (j *jenkinsHandler) CreateJob(projectName string) error {
	resp, err := utils.RequestHandle(http.MethodPost, fmt.Sprintf("%s/createItem?name=%s&mode=copy&from=base", jenkinsBaseURL(), projectName), func(req *http.Request) error {
		req.Header = http.Header{
			"Authorization": []string{"token " + utils.BasicAuthEncode("admin", "admin")},
		}
		return nil
	}, nil)
	if err != nil {
		return err
	}
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode >= http.StatusInternalServerError {
			return fmt.Errorf("Internal error: %+v", err)
		}
		logs.Info("Requested Jenkins clone job with response status code: %d", resp.StatusCode)
	}
	return nil
}

func (j *jenkinsHandler) ToggleJob(projectName, action string) error {
	resp, err := utils.RequestHandle(http.MethodPost, fmt.Sprintf("%s/job/%s/%s", jenkinsBaseURL(), projectName, action), func(req *http.Request) error {
		req.Header = http.Header{
			"Authorization": []string{"token " + utils.BasicAuthEncode("admin", "admin")},
		}
		return nil
	}, nil)
	if err != nil {
		return err
	}
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode >= http.StatusInternalServerError {
			return fmt.Errorf("Internal error: %+v", err)
		}
		logs.Info("Requested Jenkins toggle job with action %s and response status code: %d", action, resp.StatusCode)
	}
	return nil
}