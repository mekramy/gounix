package gounix

// ServerBlock nginx site block manager.
type ServerBlock interface {
	// Name sets the name of the site.
	Name(name string) ServerBlock
	// Port sets the port for the site.
	Port(port string) ServerBlock
	// Domains sets the domains for the site.
	Domains(domains ...string) ServerBlock
	// Template sets the template for the site.
	// template string can contain {domains} and {port} placeholders.
	Template(engine TemplateEngine) ServerBlock
	// Disable disables the site manually.
	Disable() error
	// Enable enables the site manually.
	Enable() error
	// Exists checks if the site exists.
	Exists() (bool, error)
	// Enabled checks if the site exists and enabled.
	Enabled() (bool, error)
	// Install installs the site.
	// override parameter indicating whether to override existing configurations.
	// returns false if site exists and not override.
	Install(override bool) (bool, error)
	// Uninstall uninstalls the site.
	Uninstall() error
}

// NewNginxReverseProxy create new nginx reverse proxy block.
func NewNginxReverseProxy(name, port string) ServerBlock {
	server := new(nginxReverseProxy)
	server.name = name
	server.port = port
	server.template = NewTemplate()
	server.template.SetTemplate(`
server {
        listen 80;
        listen [::]:80;
        server_name {domains};

        location / {
            client_max_body_size 1M;
            proxy_pass http://localhost:{port};
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_set_header Referer $http_referer;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header X-Forwarded-For $remote_addr;
            proxy_set_header X-Forwarded-Referer $http_referer;
            proxy_cache_bypass $http_upgrade;
        }
}
	`)
	return server
}
