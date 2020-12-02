package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bitrise-io/go-steputils/stepconf"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
)

type config struct {
	AppName  *string `env:"app_name"`
	Platform *string `env:"platform"`
	Bump     *string `env:"bump"`
}

type application struct {
	Name     string
	Bundle   string
	Platform string
	Version  string
	Build    int32
}

func failf(format string, v ...interface{}) {
	fmt.Errorf(format, v...)
	os.Exit(1)
}

func main() {
	var cfg config
	if err := stepconf.Parse(&cfg); err != nil {
		fmt.Printf("Issue with input: %s", err)
	}
	stepconf.Print(cfg)
	fmt.Println()

	if cfg.AppName == nil {
		failf("AppName not provided, however this is required.")
	}

	if cfg.Platform == nil {
		failf("Platform not provided, however this is required.")
	}

	bump := "patch"
	if cfg.Bump != nil {
		bump = *cfg.Bump
	}

	if bump != "patch" && bump != "minor" {
		fmt.Println("Bump should be patch or minor")
	}

	app, err := incrementVersion(*cfg.AppName, *cfg.Platform, bump)
	fmt.Println(fmt.Sprintf("Version: %s Build: %d", app.Version, app.Build))

	cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", "VERSION_NAME", "--value", app.Version).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err, cmdLog)
		os.Exit(1)
	}

	cmdLog, err = exec.Command("bitrise", "envman", "add", "--key", "BITRISE_BUILD_NUMBER", "--value", fmt.Sprintf("%d", app.Build)).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err, cmdLog)
		os.Exit(1)
	}

	os.Exit(0)
}

func incrementVersion(name string, platform string, bump string) (app *application, err error) {
	type increment struct {
		Name     string `json:"name"`
		Platform string `json:"platform"`
		Bump     string `json:"bump"`
	}

	var inc = increment{
		Name:     name,
		Platform: platform,
		Bump:     bump,
	}

	incrementURL, err := url.Parse(fmt.Sprintf("%s/version/increment", os.Getenv("VERSION_BUILDER_API_URL")))
	if err != nil {
		failf("Could not build increment url")
	}

	body, err := json.Marshal(inc)
	if err != nil {
		failf("Could not create serialize increment body")
	}

	reqBody := ioutil.NopCloser(bytes.NewBuffer(body))
	basic := getBasicAuth()

	req := &http.Request{
		Method: http.MethodPut,
		URL:    incrementURL,
		Header: map[string][]string{
			"Content-Type":  {"application/json; charset=utf-8"},
			"Authorization": {basic},
		},
		Body: reqBody,
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		failf("Call /version/increment failed %s", err.Error())
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			failf("Could not close response body %s", err.Error())
		}
	}()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		failf("Could not read response body %s", err.Error())
	}

	app = &application{}
	err = json.Unmarshal(bodyBytes, app)
	if err != nil {
		failf("Could not parse response body %s", err.Error())
	}

	return app, nil
}

func getBasicAuth() string {
	data := fmt.Sprintf("%s:%s", os.Getenv("VERSION_BUILDER_API_USERNAME"), os.Getenv("VERSION_BUILDER_API_SECRET"))
	sEnc := base64.StdEncoding.EncodeToString([]byte(data))
	return fmt.Sprintf("Basic %s", sEnc)
}
