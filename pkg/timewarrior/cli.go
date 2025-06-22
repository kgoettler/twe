package timewarrior

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type CLIError struct {
	Command string
	Stdout  string
	Stderr  string
	error   error
}

// Error implements the error interface.
func (e *CLIError) Error() string {
	return e.Stderr
}

// Unwrap allows errors.Is and errors.As to work with wrapped errors.
func (e *CLIError) Unwrap() error {
	return e.error
}

type CLI struct {
	baseCmd  string
	baseArgs []string
}

func NewCLI() CLI {
	return CLI{
		baseCmd: "timew",
	}
}

func (cli *CLI) GetIntervalByID(id int) (Interval, error) {
	cmd := cli.buildCommand("export", fmt.Sprintf("@%d", id))
	output, err := cmd.Output()
	if err != nil {
		return Interval{}, err
	}
	var intervals []Interval
	err = json.Unmarshal(output, &intervals)
	if err != nil {
		return Interval{}, fmt.Errorf("parsing command output: %w", err)
	}
	return intervals[0], nil
}

func (cli *CLI) Delete(interval Interval) error {
	_, err := cli.runCommand("delete", fmt.Sprintf("@%d", interval.ID))
	return err
}

func (cli *CLI) Modify(id int, field string, value string) error {
	_, err := cli.runCommand("modify", field, fmt.Sprintf("@%d", id), value, ":adjust")
	return err
}

func (cli *CLI) Undo() error {
	_, err := cli.runCommand("undo")
	return err
}

func (cli *CLI) Export(args ...string) ([]Interval, error) {
	// Run
	output, err := cli.runCommand(append([]string{"export"}, args...)...)
	if err != nil {
		return nil, err
	}

	// Parse output
	var intervals []Interval
	err = json.Unmarshal(output, &intervals)
	if err != nil {
		return nil, fmt.Errorf("parsing command output: %w", err)
	}
	return intervals, nil
}

func (cli *CLI) Report(args ...string) (io.Reader, error) {
	output, err := cli.runCommand(args...)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(output), nil
}

func (cli *CLI) Retag(id int, tags []string) error {
	// #nosec G204
	args := []string{
		"retag",
		fmt.Sprintf("@%d", id),
	}
	args = append(args, tags...)
	_, err := cli.runCommand(args...)
	return err
}

func (cli *CLI) Track(interval Interval) error {
	args := []string{
		"track",
		interval.Start.LocalString(),
		"-",
		interval.End.LocalString(),
	}
	args = append(args, interval.GetTags()...)
	args = append(args, ":debug", ":adjust")

	// #nosec G204
	cmd := exec.Command(
		cli.baseCmd,
		args...,
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = bufio.NewWriter(&stdout)
	cmd.Stderr = bufio.NewWriter(&stderr)
	err := cmd.Run()
	if err != nil {
		//nolint: lll // for debugging only right now
		return fmt.Errorf("running `%s`: (%w)\nSTDOUT: %s\nSTDERR: %s", strings.Join(cmd.Args, " "), err, stdout.String(), stderr.String())
	}
	return nil
}

func (cli *CLI) buildCommand(args ...string) *exec.Cmd {
	// #nosec G204
	cmd := exec.Command(
		cli.baseCmd,
		append(cli.baseArgs, args...)...,
	)
	return cmd
}

func (cli *CLI) runCommand(args ...string) ([]byte, error) {
	cmd := cli.buildCommand(args...)
	output, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			newErr := &CLIError{
				Command: strings.Join(cmd.Args, " "),
				Stdout:  string(output),
				Stderr:  string(ee.Stderr),
				error:   err,
			}
			return nil, newErr
		}
	}
	return output, nil
}
