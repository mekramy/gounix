package gounix

// Weekday represents a day of the week for cron job.
type Weekday int

const (
	Auto      Weekday = 0 // only use on Weekly method to get weekend from global timezone
	Sunday    Weekday = 1
	Monday    Weekday = 2
	Tuesday   Weekday = 3
	Wednesday Weekday = 4
	Thursday  Weekday = 5
	Friday    Weekday = 6
	Saturday  Weekday = 7
)

func (wd Weekday) IsValid() bool {
	return wd >= Sunday && wd <= Saturday
}

func (wd Weekday) Real() int {
	return int(wd) - 1
}

// CronTZ represents a timezone for a cron job.
type CronTZ struct {
	hour    int
	minute  int
	weekend Weekday
}

func (tz *CronTZ) Hour(hour int) *CronTZ {
	tz.hour = hour
	return tz
}

func (tz *CronTZ) Minute(minute int) *CronTZ {
	tz.minute = minute
	return tz
}

func (tz *CronTZ) Weekend(weekend Weekday) *CronTZ {
	tz.weekend = weekend
	return tz
}
