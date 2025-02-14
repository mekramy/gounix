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
func (c *cronDriver) set(minute, hour, day, mon, wd string) CronJob {
	c.minute = minute
	c.hour = hour
	c.day = day
	c.month = mon
	c.weekday = wd
	return c
}

// tzHour get time zone hour interval.
func (c *cronDriver) tzHour() time.Duration {
	if c.tz != nil {
		return time.Duration(c.tz.hour) * time.Hour
	}
	return 0
}

// tzMinute get time zone minute interval.
func (c *cronDriver) tzMinute() time.Duration {
	if c.tz != nil {
		return time.Duration(c.tz.minute) * time.Minute
	}
	return 0
}

// weekend get time zone weekend.
func (c *cronDriver) weekend() Weekday {
	if c.tz != nil && c.tz.weekend.IsValid() {
		return c.tz.weekend
	}
	return Sunday
}

// interval calculate cron expression
// hour and minute based of timezone.
func (c *cronDriver) interval() string {
	def := c.minute + " " +
		c.hour + " " +
		c.day + " " +
		c.month + " " +
		c.weekday

	// Return default interval if minute or hour not specified
	if c.minute == "*" || strings.Contains(c.minute, "*/") ||
		c.hour == "*" || strings.Contains(c.hour, "*/") {
		return def
	}

	// Parse cron time
	tz, err := time.Parse("15:4", c.hour+":"+c.minute)
	if err != nil {
		return def
	}

	// Calculate the time in the specified timezone
	duration := -(c.tzHour() + c.tzMinute())
	timeInTz := tz.Add(duration)

	return timeInTz.Format("4 ") +
		timeInTz.Format("15 ") +
		c.day + " " +
		c.month + " " +
		c.weekday
}

func (c *cronDriver) AtReboot() CronJob {
	c.reboot = true
	return c
}

func (c *cronDriver) Yearly() CronJob {
	return c.set("0", "0", "1", "1", "*")
}

func (c *cronDriver) Monthly() CronJob {
	return c.set("0", "0", "1", "*", "*")
}

func (c *cronDriver) Weekly(wd Weekday) CronJob {
	if !wd.IsValid() {
		return c.set("0", "0", "*", "*", strconv.Itoa(c.weekend().Real()))
	} else {
		return c.set("0", "0", "*", "*", strconv.Itoa(wd.Real()))
	}
}

func (c *cronDriver) Daily() CronJob {
	return c.set("0", "0", "*", "*", "*")
}

func (c *cronDriver) EveryXHours(hours int) CronJob {
	c.hour = "*/" + strconv.Itoa(hours)
	return c
}

func (c *cronDriver) EveryXMinutes(minutes int) CronJob {
	c.minute = "*/" + strconv.Itoa(minutes)
	return c
}

func (c *cronDriver) SetMinute(minute int) CronJob {
	if minute >= 0 && minute <= 59 {
		c.minute = strconv.Itoa(minute)
	}
	return c
}

func (c *cronDriver) SetHour(hour int) CronJob {
	if hour >= 0 && hour <= 23 {
		c.hour = strconv.Itoa(hour)
	}
	return c
}

func (c *cronDriver) SetDayOfMonth(day int) CronJob {
	if day >= 1 && day <= 31 {
		c.day = strconv.Itoa(day)
	}
	return c
}

func (c *cronDriver) SetMonth(month int) CronJob {
	if month >= 1 && month <= 12 {
		c.month = strconv.Itoa(month)
	}
	return c
}

func (c *cronDriver) SetDayOfWeek(day Weekday) CronJob {
	if day.IsValid() {
		c.weekday = strconv.Itoa(day.Real())
	}
	return c
}

func (c *cronDriver) Command(command string) CronJob {
	c.command = command
	return c
}

func (c *cronDriver) Compile() string {
	if c.reboot {
		return "@reboot " + c.command
	} else {
		return c.interval() + " " + c.command
	}
}

func (c *cronDriver) Exists() (bool, error) {
	// Read cron jobs
	lines, err := allCrons()
	if err != nil {
		return false, err
	}

	// Search command in cron jobs
	for _, line := range lines {
		ok, cmd := parseCommand(line)
		if ok && cmd == c.command {
			return true, nil
		}
	}

	// Not found
	return false, nil
}

func (c *cronDriver) Install() (bool, error) {
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
		if ok && cmd == c.command {
			exists = true
			result.WriteString(c.Compile() + "\n")
			continue
		}

		// Other valid commands
		if strings.TrimSpace(line) != "" {
			result.WriteString(line + "\n")
		}
	}

	// Append if not exists
	if !exists {
		result.WriteString(c.Compile() + "\n")
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

func (c *cronDriver) Uninstall() error {
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
		if ok && cmd == c.command {
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
