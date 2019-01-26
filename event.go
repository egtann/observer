package observer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type EventSet struct {
	RoleTimings map[string]time.Duration
	HostTimings map[string]time.Duration
	MsgTimings  []*Timing
}

type Event struct {
	Role      string  `json:"role"`
	Host      string  `json:"host"`
	RequestID string  `json:"req_id"`
	Time      decTime `json:"ts"`
	Msg       string  `json:"msg"`
}

type Timing struct {
	Msg      string
	Duration time.Duration
}

// RequestDetail describes at a higher level the route the request took across
// hosts/roles and the total time.
type RequestDetail struct {
	Duration time.Duration
	RolePath []*Timing
}

type decTime time.Time

func (t decTime) MarshalJSON() ([]byte, error) {
	nsec := time.Time(t).UnixNano()
	sec := float64(nsec) / float64(time.Second)
	return []byte(fmt.Sprint(sec)), nil
}

func (t *decTime) UnmarshalJSON(b []byte) error {
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	sec := int64(f)
	nsec := int64(f*float64(time.Millisecond)) % sec
	*t = decTime(time.Unix(sec, nsec))
	return nil
}

func (t decTime) String() string {
	return time.Time(t).Format(time.RFC3339Nano)
}

// ParseEvents in a reader line by line.
func ParseEvents(rdr io.Reader) ([]*Event, error) {
	r := bufio.NewReader(rdr)
	events := []*Event{}

	// Since log files tend to be large, bufio.NewScanner can throw
	// bufio.ErrTooLong, so we read line by line instead.
	var err error
	for byt := []byte{}; err != io.EOF && err != context.Canceled; byt, err = r.ReadBytes('\n') {
		evt := &Event{}
		if err := json.Unmarshal(byt, evt); err != nil {
			continue
		}
		events = append(events, evt)
	}
	sort.Slice(events, func(i, j int) bool {
		t1 := time.Time(events[i].Time)
		t2 := time.Time(events[j].Time)
		return t1.Before(t2)
	})
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

// LastFile in a directory.
func LastFile(dir, ext string) (string, error) {
	// Get files in migration dir
	files := []os.FileInfo{}
	tmp, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", errors.Wrap(err, "read dir")
	}
	for _, fi := range tmp {
		// Skip directories and hidden files
		if fi.IsDir() || strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		if ext != "" {
			// Skip any files that don't match the extension
			if filepath.Ext(fi.Name()) != "."+ext {
				continue
			}
		}
		files = append(files, fi)
	}
	if len(files) == 0 {
		return "", errors.New("no logs found (might be the wrong -log)")
	}

	// Sort the files by name, ensuring that something like 1.log, 2.log,
	// 10.log is correct
	regexNum := regexp.MustCompile(`^\d+`)
	sort.Slice(files, func(i, j int) bool {
		fiName1 := regexNum.FindString(files[i].Name())
		fiName2 := regexNum.FindString(files[j].Name())
		fiNum1, err := strconv.ParseUint(fiName1, 10, 64)
		if err != nil {
			err = errors.Wrapf(err, "parse uint in file %s", files[i].Name())
			panic(err)
		}
		fiNum2, err := strconv.ParseUint(fiName2, 10, 64)
		if err != nil {
			err = errors.Wrapf(err, "parse uint in file %s", files[i].Name())
			panic(err)
		}
		if fiNum1 == fiNum2 {
			err = fmt.Errorf("cannot have duplicate timestamp: %d", fiNum1)
			panic(err)
		}
		return fiNum1 < fiNum2
	})
	name := filepath.Join(dir, files[len(files)-1].Name())
	return name, nil
}

func NewEventSet(events []*Event) *EventSet {
	set := &EventSet{
		RoleTimings: map[string]time.Duration{},
		HostTimings: map[string]time.Duration{},
		MsgTimings:  []*Timing{},
	}
	roleTimes := map[string][]time.Time{}
	hostTimes := map[string][]time.Time{}
	for i := 0; i < len(events)-1; i++ {
		ev := events[i]
		if ev.Role != "" {
			roleTimes[ev.Role] = append(roleTimes[ev.Role], time.Time(ev.Time))
		}
		if ev.Host != "" {
			hostTimes[ev.Host] = append(hostTimes[ev.Host], time.Time(ev.Time))
		}
	}
	for role, times := range roleTimes {
		for i := 0; i < len(times)-1; i++ {
			t1 := times[i]
			t2 := times[i+1]
			set.RoleTimings[role] += t2.Sub(t1)
		}
	}
	for host, times := range hostTimes {
		for i := 0; i < len(times)-1; i++ {
			t1 := times[i]
			t2 := times[i+1]
			set.HostTimings[host] += t2.Sub(t1)
		}
	}
	for i := 0; i < len(events)-1; i++ {
		t1 := time.Time(events[i].Time)
		t2 := time.Time(events[i+1].Time)
		timing := &Timing{
			Msg:      events[i].Msg,
			Duration: t2.Sub(t1),
		}
		set.MsgTimings = append(set.MsgTimings, timing)
	}
	return set
}

func (dt decTime) Sub(dt2 decTime) time.Duration {
	t := time.Time(dt)
	t2 := time.Time(dt2)
	return t.Sub(t2)
}

func (dt decTime) Before(dt2 decTime) bool {
	t := time.Time(dt)
	t2 := time.Time(dt2)
	return t.Before(t2)
}

func FilterByRequestID(evs []*Event, id string) []*Event {
	out := []*Event{}
	for _, ev := range evs {
		if ev.RequestID == id {
			out = append(out, ev)
		}
	}
	return out
}

func RequestDetailFromEvents(evs []*Event) *RequestDetail {
	if len(evs) <= 1 {
		return nil
	}
	rdt1 := time.Time(evs[0].Time)
	rdt2 := time.Time(evs[len(evs)-1].Time)
	rd := &RequestDetail{Duration: rdt2.Sub(rdt1)}
	var curRoleIdx int
	for i, ev := range evs {
		if i == 0 {
			curRoleIdx = i
			continue
		}
		if evs[curRoleIdx].Role == ev.Role {
			continue
		}
		t1 := time.Time(evs[curRoleIdx].Time)
		t2 := time.Time(evs[i-1].Time)
		timing := &Timing{
			Msg:      evs[curRoleIdx].Role,
			Duration: t2.Sub(t1),
		}
		rd.RolePath = append(rd.RolePath, timing)
		curRoleIdx = i
	}
	t1 := time.Time(evs[curRoleIdx].Time)
	t2 := time.Time(evs[len(evs)-1].Time)
	timing := &Timing{
		Msg:      evs[curRoleIdx].Role,
		Duration: t2.Sub(t1),
	}
	rd.RolePath = append(rd.RolePath, timing)
	return rd
}
