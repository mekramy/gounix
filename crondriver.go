package gounix

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type cronDriver struct {
	command string
	tz      *CronTZ

	reboot  bool
	minute  string
	hour    string
	day     string
	month   string
	weekday string
}

// .---------------- minute (0 - 59)
// |  .------------- hour (0 - 23)
// |  |  .---------- day of month (1 - 31)
// |  |  |  .------- month (1 - 12) OR jan, feb, mar, apr ...
// |  |  |  |  .---- day of week (0 - 6) (Sunday=0) OR sun, mon, tue, wed, thu, fri, sat
// |  |  |  |  |
// m h dom mon dow command
func (driver *cronDriver) set(minute, hour, day, mon, wd string) CronJob {
	driver.minute = minute
	driver.hour = hour
	driver.day = day
	driver.month = mon
	driver.weekday = wd
	return driver
}

// tzHour get time zone hour interval.
func (driver *cronDriver) tzHour() time.Duration {
	if driver.tz != nil {
		return time.Duration(driver.tz.hour) * time.Hour
	}
	return 0
}

// tzMinute get time zone minute interval.
func (driver *cronDriver) tzMinute() time.Duration {
	if driver.tz != nil {
		return time.Duration(driver.tz.minute) * time.Minute
	}
	return 0
}

// weekend get time zone weekend.
func (driver *cronDriver) weekend() Weekday {
	if driver.tz != nil && driver.tz.weekend.IsValid() {
		return driver.tz.weekend
	}
	return Sunday
}

// interval calculate cron expression
// hour and minute based of timezone.
func (driver *cronDriver) interval() string {
	def := driver.minute + " " +
		driver.hour + " " +
		driver.day + " " +
		driver.month + " " +
		driver.weekday

	// Return default interval if minute or hour not specified
	if driver.minute == "*" || strings.Contains(driver.minute, "*/") ||
		driver.hour == "*" || strings.Contains(driver.hour, "*/") {
		return def
	}

	// Parse cron time
	tz, err := time.Parse("15:4", driver.hour+":"+driver.minute)
	if err != nil {
		return def
	}

	// Calculate the time in the specified timezone
	duration := -(driver.tzHour() + driver.tzMinute())
	timeInTz := tz.Add(duration)

	return timeInTz.Format("4 ") +
		timeInTz.Format("15 ") +
		driver.day + " " +
		driver.month + " " +
		driver.weekday
}

func (driver *cronDriver) AtReboot() CronJob {
	driver.reboot = true
	return driver
}

func (driver *cronDriver) Yearly() CronJob {
	return driver.set("0", "0", "1", "1", "*")
}

func (driver *cronDriver) Monthly() CronJob {
	return driver.set("0", "0", "1", "*", "*")
}

func (driver *cronDriver) Weekly(wd Weekday) CronJob {
	if !wd.IsValid() {
		return driver.set("0", "0", "*", "*", strconv.Itoa(driver.weekend().Real()))
	} else {
		return driver.set("0", "0", "*", "*", strconv.Itoa(wd.Real()))
	}
}

func (driver *cronDriver) Daily() CronJob {
	return driver.set("0", "0", "*", "*", "*")
}

func (driver *cronDriver) EveryXHours(hours int) CronJob {
	driver.hour = "*/" + strconv.Itoa(hours)
	return driver
}

func (driver *cronDriver) EveryXMinutes(minutes int) CronJob {
	driver.minute = "*/" + strconv.Itoa(minutes)
	return driver
}

func (driver *cronDriver) SetMinute(minute int) CronJob {
	if minute >= 0 && minute <= 59 {
		driver.minute = strconv.Itoa(minute)
	}
	return driver
}

func (driver *cronDriver) SetHour(hour int) CronJob {
	if hour >= 0 && hour <= 23 {
		driver.hour = strconv.Itoa(hour)
	}
	return driver
}

func (driver *cronDriver) SetDayOfMonth(day int) CronJob {
	if day >= 1 && day <= 31 {
		driver.day = strconv.Itoa(day)
	}
	return driver
}

func (driver *cronDriver) SetMonth(month int) CronJob {
	if month >= 1 && month <= 12 {
		driver.month = strconv.Itoa(month)
	}
	return driver
}

func (driver *cronDriver) SetDayOfWeek(day Weekday) CronJob {
	if day.IsValid() {
		driver.weekday = strconv.Itoa(day.Real())
	}
	return driver
}

func (driver *cronDriver) Command(command string) CronJob {
	driver.command = command
	return driver
}

func (driver *cronDriver) Compile() string {
	if driver.reboot {
		return "@reboot " + driver.command
	} else {
		return driver.interval() + " " + driver.command
	}
}

func (driver *cronDriver) Exists() (bool, error) {
	// Read cron jobs
	lines, err := allCrons()
	if err != nil {
		return false, err
	}

	// Search command in cron jobs
	for _, line := range lines {
		ok, cmd := parseCommand(line)
		if ok && cmd == driver.command {
			return true, nil
		}
	}

	// Not found
	return false, nil
}

func (driver *cronDriver) Install() (bool, error) {
	var exists bool
	var result strings.Builder

	// Read cron jobs
	lines, err := allCrons()
	if err != nil {
		return false, err
	}

	// Try to find and update cron job
	for _, line := range lines {
		// Find and update
		ok, cmd := parseCommand(line)
		if ok && cmd == driver.command {
			exists = true
			result.WriteString(driver.Compile() + "\n")
			continue
		}

		// Other valid commands
		if strings.TrimSpace(line) != "" {
			result.WriteString(line + "\n")
		}
	}

	// Append if not exists
	if !exists {
		result.WriteString(driver.Compile() + "\n")
	}

	// Update cron jobs
	cmd := `echo "` + result.String() + `" | crontab -`
	err = cmdError(exec.Command("sudo", "bash", "-c", cmd).Run())
	if err != nil {
		return false, err
	}

	// Restart cron service
	err = cmdError(exec.Command("sudo", "systemctl", "restart", "cron").Run())
	if err != nil {
		return false, err
	}

	return true, nil
}

func (driver *cronDriver) Uninstall() error {
	var result strings.Builder

	// Read cron jobs
	lines, err := allCrons()
	if err != nil {
		return err
	}

	// Exclude cron from jobs list
	for _, line := range lines {
		// Find and exclude
		ok, cmd := parseCommand(line)
		if ok && cmd == driver.command {
			continue
		}

		// Other valid commands
		if strings.TrimSpace(line) != "" {
			result.WriteString(line + "\n")
		}
	}

	// Update cron jobs
	cmd := `echo "` + result.String() + `" | crontab -`
	err = cmdError(exec.Command("sudo", "bash", "-c", cmd).Run())
	if err != nil {
		return err
	}

	// Restart cron service
	return cmdError(exec.Command("sudo", "systemctl", "restart", "cron").Run())
}
