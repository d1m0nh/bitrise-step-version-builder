package main

import (
	"fmt"
	"github.com/bitrise-io/go-steputils/stepconf"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type config struct {
	VersionName *string `env:"version_name"`
	Bump    	*string `env:"bump"`
}

func main() {
	var cfg config
	if err := stepconf.Parse(&cfg); err != nil {
		fmt.Printf("Issue with input: %s", err)
	}
	stepconf.Print(cfg)
	fmt.Println()

	if cfg.VersionName == nil {
		fmt.Println("VersionName not provided, however this is required.")
	}

	bump := "patch"
	if cfg.Bump != nil {
		bump = *cfg.Bump
	}

	if bump != "patch" && bump != "minor" {
		fmt.Println("Bump should be patch or minor")
	}

	version, err := incrementVersion(*cfg.VersionName, bump)
	fmt.Println(fmt.Sprintf("New version %s", version))

	cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", "VERSION_NAME", "--value", version).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err, cmdLog)
		os.Exit(1)
	}

	os.Exit(0)
}

func incrementVersion(version string, bump string) (ver string, err error) {
	if bump == "minor" {
		ver, err = incrementMinor(version)
	} else {
		ver, err = incrementPatch(version)
	}

	return ver, nil
}

func incrementMinor(version string) (string, error) {
	splitted := strings.Split(version, ".")
	major, err := strconv.Atoi(splitted[0])
	if err != nil {
		return version, err
	}

	minor, err := strconv.Atoi(splitted[1])
	if err != nil {
		return version, err
	}

	minor++
	return fmt.Sprintf("%v.%v.%v", major, minor, 0), nil
}

func incrementPatch(version string) (string, error) {
	splitted := strings.Split(version, ".")
	major, err := strconv.Atoi(splitted[0])
	if err != nil {
		return version, err
	}
	minor, err := strconv.Atoi(splitted[1])
	if err != nil {
		return version, err
	}

	patch,err := strconv.Atoi(splitted[2])
	if err != nil {
		return version, err
	}

	patch++
	return fmt.Sprintf("%v.%v.%v", major, minor, patch), nil
}


