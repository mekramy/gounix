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

func (s systemdDriver) path() string {
	return "/etc/systemd/system/" + s.name + ".service"
}

func (s *systemdDriver) Name(name string) SystemdService {
	s.name = name
	return s
}

func (s *systemdDriver) Root(dir string) SystemdService {
	s.root = dir
	return s
}

func (s *systemdDriver) Command(command string) SystemdService {
	s.command = command
	return s
}

func (s *systemdDriver) Template(engine TemplateEngine) SystemdService {
	s.template = engine
	return s
}

func (s *systemdDriver) Exists() bool {
	_, err := exec.Command("sudo", "systemctl", "status", s.name).Output()
	return err == nil
}

func (s *systemdDriver) Enabled() bool {
	output, _ := exec.Command("sudo", "systemctl", "is-enabled", s.name).Output()
	return strings.HasPrefix(string(output), "enabled")
}

func (s *systemdDriver) Install(override bool) (bool, error) {
	// Check exists and override
	exists := s.Exists()
	if exists && !override {
		return false, nil
	}

	// Compile template
	content := []byte(
		s.template.
			AddParameter("name", s.name).
			AddParameter("root", s.root).
			AddParameter("command", s.command).
			Compile(),
	)

	// Create service file
	err := os.WriteFile(s.path(), []byte(content), 0644)
	if err != nil {
		return false, err
	}

	// Reload services
	err = cmdError(exec.Command("sudo", "systemctl", "daemon-reload").Run())
	if err != nil {
		return false, err
	}

	// Enable service on startup
	err = cmdError(exec.Command("sudo", "systemctl", "enable", s.name).Run())
	if err != nil {
		return false, err
	}

	// Start service
	err = cmdError(exec.Command("sudo", "systemctl", "start", s.name).Run())
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *systemdDriver) Uninstall() error {
	if s.Exists() {
		// Stop service
		err := cmdError(exec.Command("sudo", "systemctl", "stop", s.name).Run())
		if err != nil {
			return err
		}

		// Disable service
		err = cmdError(exec.Command("sudo", "systemctl", "disable", s.name).Run())
		if err != nil {
			return err
		}
	}

	err := os.Remove(s.path())
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
