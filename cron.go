package gounix

import (
	"os/exec"
	"strings"
)

// CronJob represents a cron job.
type CronJob interface {
	// AtReboot schedules the cron job to run at reboot.
	AtReboot() CronJob
	// Yearly schedules the cron job to run every year.
	Yearly() CronJob
	// Monthly schedules the cron job to run every month.
	Monthly() CronJob
	// Weekly schedules the cron job to run every week.
	Weekly(wd Weekday) CronJob
	// Daily schedules the cron job to run every day.
	Daily() CronJob
	// EveryXHours schedules the cron job to run every specified number of hours.
	EveryXHours(hours int) CronJob
	// EveryXMinutes schedules the cron job to run every specified number of minutes.
	EveryXMinutes(minutes int) CronJob
	// SetMinute sets the minute of the cron job.
	SetMinute(minute int) CronJob
	// SetHour sets the hour of the cron job.
	SetHour(hour int) CronJob
	// SetDayOfMonth sets the day of the month of the cron job.
	SetDayOfMonth(day int) CronJob
	// SetMonth sets the month of the cron job.
	SetMonth(month int) CronJob
	// SetDayOfWeek sets the day of the week of the cron job.
	SetDayOfWeek(day Weekday) CronJob
	// Command sets the command to be executed by the cron job.
	Command(command string) CronJob
	// Compile compiles the cron job into a cron expression string.
	Compile() string
	// Exists checks if the cron job already exists.
	Exists() (bool, error)
	// Install installs the cron job. returns false if cronjob exists.
	Install() (bool, error)
	// Uninstall uninstalls the cron job.
	Uninstall() error
}

// NewTZ creates a new timezone for a cron job.
func NewTZ() *CronTZ {
	return new(CronTZ)
}

// NewCronJob creates a new cron job.
func NewCronJob(command string, tz *CronTZ) CronJob {
	driver := new(cronDriver)
	driver.command = command
	driver.tz = tz
	driver.set("*", "*", "*", "*", "*")
	return driver
}

// SetCronTZ sets the timezone of the cron daemon to the specified timezone.
func SetCronTZ(tz string) error {
	if lines, err := allCrons(); err != nil {
		return err
	} else {
		var result strings.Builder
		result.WriteString("TZ=" + tz + "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "TZ=") {
				result.WriteString(line + "\n")
			}
		}
		cmd := `echo "` + result.String() + `" | crontab -`
		return exec.Command("sudo", "bash", "-c", cmd).Run()
	}
}
