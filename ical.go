package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/apognu/gocal"
)

type icalImpl struct {
	cal *gocal.Gocal

	years []int
}

func (ic *icalImpl) Years() []int {
	return ic.years
}

func (ic *icalImpl) Months(year int) []time.Month {
	months := map[time.Month]struct{}{}
	for _, e := range ic.cal.Events {
		if e.Start.Year() == year {
			months[e.Start.Month()] = struct{}{}
		}
	}

	retval := make([]time.Month, 0, len(months))
	for m := range months {
		retval = append(retval, m)
	}

	return retval
}

func (ic *icalImpl) Days(year int, month time.Month) []int {
	days := map[int]struct{}{}
	for _, e := range ic.cal.Events {
		if e.Start.Year() == year && e.Start.Month() == month {
			days[e.Start.Day()] = struct{}{}
		}
	}

	// todo: replace with go 1.23 maps.Keys()
	retval := make([]int, 0, len(days))
	for d := range days {
		retval = append(retval, d)
	}

	return retval
}

func (ic *icalImpl) Entries(day time.Time) []CalendarEntry {
	var retval []CalendarEntry
	for _, e := range ic.cal.Events {
		if e.Start.Year() == day.Year() && e.Start.Month() == day.Month() && e.Start.Day() == day.Day() {
			retval = append(retval, CalendarEntry{
				Start:       *e.Start,
				End:         *e.End,
				Summary:     e.Summary,
				Description: e.Description,
				Notes:       e.Comment,
			})
		}
	}
	return retval
}

func OpenICal(icalFile string) (CalendarInterface, error) {
	f, err := os.Open(icalFile)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %s", icalFile, err)
	}

	c := gocal.NewParser(f)
	if err := c.Parse(); err != nil {
		f.Close()
		return nil, err
	}

	yearsSeen := map[int]struct{}{}
	for _, e := range c.Events {
		yearsSeen[e.Start.Year()] = struct{}{}
	}

	var years []int
	for y := range yearsSeen {
		years = append(years, y)
	}
	sort.Ints(years)

	return &icalImpl{cal: c, years: years}, nil
}
