package gounix

import (
	"os"
	"os/exec"
	"strings"
)

type nginxReverseProxy struct {
	name     string
	port     string
	domains  []string
	template TemplateEngine
}

func (driver *nginxReverseProxy) path() string {
	return "/etc/nginx/sites-available/" + driver.name
}

func (driver *nginxReverseProxy) link() string {
	return "/etc/nginx/sites-enabled/" + driver.name
}

func (driver *nginxReverseProxy) Name(name string) ServerBlock {
	driver.name = name
	return driver
}

func (driver *nginxReverseProxy) Port(port string) ServerBlock {
	driver.port = port
	return driver
}

func (driver *nginxReverseProxy) Domains(domains ...string) ServerBlock {
	driver.domains = append(driver.domains, domains...)
	return driver
}

func (driver *nginxReverseProxy) Template(engine TemplateEngine) ServerBlock {
	driver.template = engine
	return driver
}

func (driver *nginxReverseProxy) Disable() error {
	// Delete link
	err := os.Remove(driver.link())
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	// Restart nginx to apply the changes
	return cmdError(exec.Command("sudo", "systemctl", "restart", "nginx").Run())
}

func (driver *nginxReverseProxy) Enable() error {
	// Skip if exists
	exists, err := fileExists(driver.link())
	if err != nil {
		return err
	} else if exists {
		return nil
	}

	// Create link
	err = os.Symlink(driver.path(), driver.link())
	if err != nil {
		return err
	}

	// Restart nginx to apply the changes
	return cmdError(exec.Command("sudo", "systemctl", "restart", "nginx").Run())
}

func (driver *nginxReverseProxy) Exists() (bool, error) {
	return fileExists(driver.path())
}

func (driver *nginxReverseProxy) Enabled() (bool, error) {
	available, err := fileExists(driver.path())
	if err != nil {
		return false, err
	}

	enabled, err := fileExists(driver.link())
	if err != nil {
		return false, err
	}

	return available && enabled, nil
}

func (driver *nginxReverseProxy) Install(override bool) (bool, error) {
	// Check exists and override
	exists, err := fileExists(driver.path())
	if err != nil {
		return false, err
	} else if exists && !override {
		return false, nil
	}

	// Compile template
	content := []byte(
		driver.template.
			AddParameter("port", driver.port).
			AddParameter("domains", strings.Join(driver.domains, " ")).
			Compile(),
	)

	// Create server file
	err = os.WriteFile(driver.path(), content, 0644)
	if err != nil {
		return false, err
	}

	// Create link and Skip if link exists
	err = driver.Enable()
	if err != nil {
		return false, err
	}

	// Restart nginx to apply the changes
	err = cmdError(exec.Command("sudo", "systemctl", "restart", "nginx").Run())
	if err != nil {
		return false, err
	}

	return true, nil
}

func (driver *nginxReverseProxy) Uninstall() error {
	// Remove the enabled site link
	err := os.Remove(driver.link())
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Remove the available site file
	err = os.Remove(driver.path())
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Restart nginx to apply the changes
	return cmdError(exec.Command("sudo", "systemctl", "restart", "nginx").Run())
}
