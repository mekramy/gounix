# gounix

`gounix` is a Go library for managing cron jobs, Nginx server blocks, and systemd services. This library provides a simple and consistent API for creating, installing, and managing these services.

## Installation

To install `gounix`, use `go get`:

```sh
go get github.com/mekramy/gounix
```

## Usage

### Cron Jobs

The `CronJob` interface provides methods for scheduling and managing cron jobs. You can set time zone (e.g. +3:30 for Asia/Tehran) to run cron based on your timezone.

**CAUTION**: `AtReboot`, `Yearly` ,`Monthly`, `Weekly` and `Daily` method should called before other method otherwise it's override previous settings.

- `AtReboot() CronJob`
- `Yearly() CronJob`
- `Monthly() CronJob`
- `Weekly(wd Weekday) CronJob`
- `Daily() CronJob`
- `EveryXHours(hours int) CronJob`
- `EveryXMinutes(minutes int) CronJob`
- `SetMinute(minute int) CronJob`
- `SetHour(hour int) CronJob`
- `SetDayOfMonth(day int) CronJob`
- `SetMonth(month int) CronJob`
- `SetDayOfWeek(day Weekday) CronJob`
- `Command(command string) CronJob`
- `Compile() string`
- `Exists() (bool, error)`
- `Install() (bool, error)`
- `Uninstall() error`

```go
package main

import (
    "fmt"
    "github.com/mekramy/gounix"
)

func main() {
    tz := gounix.NewTZ().Hour(-2).Minute(30).Weekend(gounix.Friday)
    cronJob := gounix.NewCronJob("echo 'Hello, World!'", tz).
        Daily().
        SetHour(2).
        SetMinute(30).
        Command("echo 'Hello, World!'")

    if installed, err := cronJob.Install(); err != nil {
        fmt.Println("Error installing cron job:", err)
    } else if installed {
        fmt.Println("Cron job installed successfully")
    } else {
        fmt.Println("Cron job already exists")
    }
}
```

### Nginx Server Blocks

The `ServerBlock` interface provides methods for managing Nginx server blocks.

- `Name(name string) ServerBlock`
- `Port(port string) ServerBlock`
- `Domains(domains ...string) ServerBlock`
- `Template(engine TemplateEngine) ServerBlock`
- `Disable() error`
- `Enable() error`
- `Exists() (bool, error)`
- `Enabled() (bool, error)`
- `Install(override bool) (bool, error)`
- `Uninstall() error`

```go
package main

import (
    "fmt"
    "github.com/mekramy/gounix"
)

func main() {
    serverBlock := gounix.NewNginxReverseProxy("example", "8080").
        Domains("example.com", "www.example.com")

    if installed, err := serverBlock.Install(true); err != nil {
        fmt.Println("Error installing server block:", err)
    } else if installed {
        fmt.Println("Server block installed successfully")
    } else {
        fmt.Println("Server block already exists")
    }
}
```

### Systemd Services

The `SystemdService` interface provides methods for managing systemd services.

- `Name(name string) SystemdService`
- `Root(dir string) SystemdService`
- `Command(command string) SystemdService`
- `Template(engine TemplateEngine) SystemdService`
- `Exists() bool`
- `Enabled() bool`
- `Install(override bool) (bool, error)`
- `Uninstall() error`

```go
package main

import (
    "fmt"
    "github.com/mekramy/gounix"
)

func main() {
    service := gounix.NewSystemdService("example-service", "/path/to/service", "service-command")

    if installed, err := service.Install(true); err != nil {
        fmt.Println("Error installing systemd service:", err)
    } else if installed {
        fmt.Println("Systemd service installed successfully")
    } else {
        fmt.Println("Systemd service already exists")
    }
}
```

### Template Engine

The `TemplateEngine` interface provides methods for managing `{bracket wrapped}` templates.

- `SetTemplate(template string) TemplateEngine`
- `AddParameter(name, value string) TemplateEngine`
- `Compile() string`

```go
package main

import (
    "fmt"
    "github.com/mekramy/gounix"
)

func main() {
    engine := gounix.NewTemplate().
        SetTemplate("Hello, {name}!").
        AddParameter("name", "World")

    fmt.Println(engine.compile()) // Output: Hello, World!
}
```

### Utility Functions

#### `IsSudo`

Checks if the program is running with sudo privileges.

```go
package main

import (
    "fmt"
    "github.com/mekramy/gounix"
)

func main() {
    if gounix.IsSudo() {
        fmt.Println("Running with sudo privileges")
    } else {
        fmt.Println("Not running with sudo privileges")
    }
}
```
