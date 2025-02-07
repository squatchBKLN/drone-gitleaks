// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level        string `envconfig:"PLUGIN_LOG_LEVEL"`
	Path         string `envconfig:"PLUGIN_PATH" default:"."`
	Conf         string `envconfig:"PLUGIN_CONFIG"`
	CardFilePath string `envconfig:"DRONE_CARD_PATH"`
}

type Card struct {
	NumIssues int     `json:"NumIssues"`
	Issues    []match `json:"Issues"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {
	// generate a temp file to store the report
	file, err := ioutil.TempFile("", "gitleaks")
	if err != nil {
		logrus.WithError(err).Warnln("Cannot generate temporary file")
		return err
	}

	gCmd := exec.Command("git", "config", "--global", "--add", "safe.directory", "/harness")
	gCmd.Stderr = os.Stderr
	gCmd.Stdout = os.Stdout

	err = gCmd.Run()

	//	cmd := exec.Command("gitleaks", "--path="+args.Path, "--commit="+args.Commit.Rev, "--report="+file.Name())
	cmd := exec.Command("gitleaks", "detect", "-f", "junit", "-r", "jUnitReport.xml", "-v")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if s := args.Conf; s != "" {
		cmd.Args = append(cmd.Args, "--config-path="+s)
	}

	// execute the command
	err = cmd.Run()

	// read the generated report and unmarshal
	dat := []match{}
	out, ferr := ioutil.ReadFile(file.Name())
	if ferr != nil {
		logrus.WithError(ferr).Warnln("Cannot read report")
	}
	if jsonerr := json.Unmarshal(out, &dat); jsonerr != nil {
		logrus.WithError(jsonerr).Warnln("Cannot unmarshal report, ignore if 'No leaks found'")
	}

	issues := []match{}
	// loop through and print each violation
	for _, match := range dat {
		match.Line = "*****"
		match.Offender = "*****"
		logrus.Errorf("%s violation in %q at line %d\n", match.Rule, match.File, match.Linenumber)
		issues = append(issues, match)
	}

	cardData := Card{
		NumIssues: len(issues),
		Issues:    issues,
	}

	if err := args.writeCard(cardData); err != nil {
		fmt.Printf("Could not create adaptive card. %s\n", err)
	}

	return err
}
