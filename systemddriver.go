package gounix

import (
	"os"
	"os/exec"
	"strings"
)

type systemdDriver struct {
	name     string
	root     string
	command  string
	template TemplateEngine
}

func (driver systemdDriver) path() string {
	return "/etc/systemd/system/" + driver.name + ".service"
}

func (driver *systemdDriver) Name(name string) SystemdService {
	driver.name = name
	return driver
}

func (driver *systemdDriver) Root(dir string) SystemdService {
	driver.root = dir
	return driver
}

func (driver *systemdDriver) Command(command string) SystemdService {
	driver.command = command
	return driver
}

func (driver *systemdDriver) Template(engine TemplateEngine) SystemdService {
	driver.template = engine
	return driver
}

func (driver *systemdDriver) Exists() bool {
	_, err := exec.Command("sudo", "systemctl", "status", driver.name).Output()
	return err == nil
}

func (driver *systemdDriver) Enabled() bool {
	output, _ := exec.Command("sudo", "systemctl", "is-enabled", driver.name).Output()
	return strings.HasPrefix(string(output), "enabled")
}

func (driver *systemdDriver) Install(override bool) (bool, error) {
	// Check exists and override
	exists := driver.Exists()
	if exists && !override {
		return false, nil
	}

	// Compile template
	content := []byte(
		driver.template.
			AddParameter("name", driver.name).
			AddParameter("root", driver.root).
			AddParameter("command", driver.command).
			Compile(),
	)

	// Create service file
	err := os.WriteFile(driver.path(), []byte(content), 0644)
	if err != nil {
		return false, err
	}

	// Reload services
	err = cmdError(exec.Command("sudo", "systemctl", "daemon-reload").Run())
	if err != nil {
		return false, err
	}

	// Enable service on startup
	err = cmdError(exec.Command("sudo", "systemctl", "enable", driver.name).Run())
	if err != nil {
		return false, err
	}

	// Start service
	err = cmdError(exec.Command("sudo", "systemctl", "start", driver.name).Run())
	if err != nil {
		return false, err
	}

	return true, nil
}

func (driver *systemdDriver) Uninstall() error {
	if driver.Exists() {
		// Stop service
		err := cmdError(exec.Command("sudo", "systemctl", "stop", driver.name).Run())
		if err != nil {
			return err
		}

		// Disable service
		err = cmdError(exec.Command("sudo", "systemctl", "disable", driver.name).Run())
		if err != nil {
			return err
		}
	}

	err := os.Remove(driver.path())
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
