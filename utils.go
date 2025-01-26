package gounix

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// cmdError handle execute error with exit code.
func cmdError(err error) error {
	if err == nil {
		return nil
	} else if exitErr, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("Exit %d, %s", exitErr.ExitCode(), string(exitErr.Stderr))
	} else {
		return err
	}
}

// fileExists check if file exists.
func fileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

// parseCommand extracts the command from a cron expression.
func parseCommand(cronExpr string) (bool, string) {
	// Handle predefined constants
	aliases := []string{
		"@reboot ", "@yearly ", "@annually ",
		"@monthly ", "@weekly ", "@daily ",
		"@midnight ", "@hourly ",
	}
	for _, alias := range aliases {
		if strings.HasPrefix(cronExpr, alias) {
			return true, strings.TrimPrefix(cronExpr, alias)
		}
	}

	// Handle custom cron expressions
	parts := strings.Fields(cronExpr)
	if len(parts) < 6 {
		return false, ""
	}
	return true, strings.Join(parts[5:], " ")
}

// allCrons get all system cron jobs.
func allCrons() ([]string, error) {
	out, err := exec.Command("sudo", "crontab", "-l").Output()
	err = cmdError(err)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(out), "\n"), nil
}
