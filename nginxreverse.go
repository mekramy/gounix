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

func (n *nginxReverseProxy) path() string {
	return "/etc/nginx/sites-available/" + n.name
}

func (n *nginxReverseProxy) link() string {
	return "/etc/nginx/sites-enabled/" + n.name
}

func (n *nginxReverseProxy) Name(name string) ServerBlock {
	n.name = name
	return n
}

func (n *nginxReverseProxy) Port(port string) ServerBlock {
	n.port = port
	return n
}

func (n *nginxReverseProxy) Domains(domains ...string) ServerBlock {
	n.domains = append(n.domains, domains...)
	return n
}

func (n *nginxReverseProxy) Template(engine TemplateEngine) ServerBlock {
	n.template = engine
	return n
}

func (n *nginxReverseProxy) Disable() error {
	// Delete link
	err := os.Remove(n.link())
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	// Restart nginx to apply the changes
	return cmdError(exec.Command("sudo", "systemctl", "restart", "nginx").Run())
}

func (n *nginxReverseProxy) Enable() error {
	// Skip if exists
	exists, err := fileExists(n.link())
	if err != nil {
		return err
	} else if exists {
		return nil
	}

	// Create link
	err = os.Symlink(n.path(), n.link())
	if err != nil {
		return err
	}

	// Restart nginx to apply the changes
	return cmdError(exec.Command("sudo", "systemctl", "restart", "nginx").Run())
}

func (n *nginxReverseProxy) Exists() (bool, error) {
	return fileExists(n.path())
}

func (n *nginxReverseProxy) Enabled() (bool, error) {
	available, err := fileExists(n.path())
	if err != nil {
		return false, err
	}

	enabled, err := fileExists(n.link())
	if err != nil {
		return false, err
	}

	return available && enabled, nil
}

func (n *nginxReverseProxy) Install(override bool) (bool, error) {
	// Check exists and override
	exists, err := fileExists(n.path())
	if err != nil {
		return false, err
	} else if exists && !override {
		return false, nil
	}

	// Compile template
	content := []byte(
		n.template.
			AddParameter("port", n.port).
			AddParameter("domains", strings.Join(n.domains, " ")).
			Compile(),
	)

	// Create server file
	err = os.WriteFile(n.path(), content, 0644)
	if err != nil {
		return false, err
	}

	// Create link and Skip if link exists
	err = n.Enable()
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

func (n *nginxReverseProxy) Uninstall() error {
	// Remove the enabled site link
	err := os.Remove(n.link())
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Remove the available site file
	err = os.Remove(n.path())
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Restart nginx to apply the changes
	return cmdError(exec.Command("sudo", "systemctl", "restart", "nginx").Run())
}
