package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type EventSet struct {
	Events      []*Event
	RoleTimings map[string][]*Timing
	HostTimings map[string][]*Timing
	MsgTimings  map[string][]*Timing
}

type Event struct {
	Role      string  `json:"role"`
	Host      string  `json:"host"`
	RequestID uint64  `json:"request_id"`
	Time      decTime `json:"timestamp"`
	Msg       string  `json:"msg"`
}

type Timing struct {
	Event    *Event
	Duration time.Duration
}

type decTime time.Time

func (t *decTime) UnmarshalJSON(b []byte) error {
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	fs := fmt.Sprint(f)
	parts := strings.SplitN(fs, ".", 2)
	sec, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return errors.Wrap(err, "parse part 1")
	}
	var nsec int64
	if len(parts) > 1 {
		nsec, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return errors.Wrap(err, "parse part 1")
		}
	}
	*t = decTime(time.Unix(sec, nsec))
	return nil
}

// ParseEvents in a reader line by line.
func ParseEvents(rdr io.Reader) ([]*Event, error) {
	events := []*Event{}
	scn := bufio.NewScanner(rdr)
	for scn.Scan() {
		line := scn.Text()
		evt := &Event{}
		if err := json.Unmarshal([]byte(line), evt); err != nil {
			return nil, errors.Wrap(err, "unmarshal")
		}
		events = append(events, evt)
	}
	if err := scn.Err(); err != nil {
		return nil, errors.Wrap(err, "scan")
	}
	return events, nil
}

// ParseEventsFile is a convenience wrapper around ParseEvents to work with a
// file.
func ParseEventsFile(pth string) ([]*Event, error) {
	fi, err := os.Open(pth)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	return ParseEvents(fi)
}

func NewEventSet(events []*Event) *EventSet {
	set := &EventSet{Events: events}
	for i := 0; i < len(events)-1; i++ {
		ev := events[i]
		t := &Timing{
			Event:    ev,
			Duration: events[i+1].Time.Sub(ev.Time),
		}
		set.RoleTimings[ev.Role] = append(set.RoleTimings[ev.Role], t)
		set.HostTimings[ev.Host] = append(set.HostTimings[ev.Host], t)
		set.MsgTimings[ev.Msg] = append(set.MsgTimings[ev.Msg], t)
	}
	return set
}

func SumTimings(ts []*Timing) time.Duration {
	var dur time.Duration
	for _, t := range ts {
		dur += t.Duration
	}
	return dur
}

func (dt decTime) Sub(dt2 decTime) time.Duration {
	t := time.Time(dt)
	t2 := time.Time(dt2)
	return t.Sub(t2)
}
