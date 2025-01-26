package gounix

// ServerBlock systemd service manager.
type SystemdService interface {
	// Name sets the name of the service.
	Name(name string) SystemdService
	// Root sets the root path of the service.
	Root(dir string) SystemdService
	// Command sets the command of the service.
	Command(command string) SystemdService
	// Template sets the template for the service.
	// template string can contain {name}, {root} and {command} placeholders.
	Template(engine TemplateEngine) SystemdService
	// Exists checks if the service exists.
	Exists() bool
	// Enabled checks if the service exists and enabled on startup.
	Enabled() bool
	// Install installs the service.
	// override parameter indicating whether to override existing configurations.
	// returns false if service exists and not override.
	Install(override bool) (bool, error)
	// Uninstall uninstalls the service.
	Uninstall() error
}

// NewSystemdService create new systemd service block.
func NewSystemdService(name, root, command string) SystemdService {
	service := new(systemdDriver)
	service.name = name
	service.root = root
	service.command = command
	service.template = NewTemplate()
	service.template.SetTemplate(`
[Unit]
Description={name}
ConditionPathExists={root}
After=network.target

[Service]
Type=simple
User=root
Group=root
LimitNOFILE=1024

Restart=on-failure
RestartSec=10

WorkingDirectory={root}
ExecStart=/usr/bin/sudo {root}/{command}

PermissionsStartOnly=true
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier={name}

[Install]
WantedBy=multi-user.target
	`)
	return service
}
