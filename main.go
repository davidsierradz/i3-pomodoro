package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

const (
	stateFile = "/tmp/pomodoro-state"

	iconRun   = ""
	iconPause = ""

	defaultTimerLength      = 25 * time.Minute
	defaultShortBreakLength = 5 * time.Minute
	defaultLongBreakLength  = 15 * time.Minute

	clockPomodoro    = 0
	clockPomodoro2   = iota
	clockPomodoro3   = iota
	clockPomodoro4   = iota
	clockShortBreak  = iota
	clockShortBreak2 = iota
	clockShortBreak3 = iota
	clockLongBreak   = iota
)

// State is the state of the pomodoro timer.
type State struct {
	Running   bool
	Paused    bool
	Duration  time.Duration
	LastTime  time.Time
	Now       time.Time
	ClockType int
}

func (s *State) clockTextShort() string {
	clockText := "P1"

	switch s.ClockType {
	case clockPomodoro2:
		clockText = "P2"
	case clockPomodoro3:
		clockText = "P3"
	case clockPomodoro4:
		clockText = "P4"
	case clockShortBreak:
		clockText = "SB"
	case clockShortBreak2:
		clockText = "SB2"
	case clockShortBreak3:
		clockText = "SB3"
	case clockLongBreak:
		clockText = "LB"
	}

	return clockText
}

func (s *State) clockText() string {
	clockText := "POMODORO"

	switch s.ClockType {
	case clockPomodoro2:
		clockText = "POMODORO 2"
	case clockPomodoro3:
		clockText = "POMODORO 3"
	case clockPomodoro4:
		clockText = "POMODORO 4"
	case clockShortBreak:
		clockText = "SHORT BREAK"
	case clockShortBreak2:
		clockText = "SHORT BREAK 2"
	case clockShortBreak3:
		clockText = "SHORT BREAK 3"
	case clockLongBreak:
		clockText = "LONG BREAK"
	}

	return clockText
}

func (s *State) cycleClock() {
	switch s.ClockType {
	case clockPomodoro:
		s.ClockType = clockShortBreak
	case clockPomodoro2:
		s.ClockType = clockShortBreak2
	case clockPomodoro3:
		s.ClockType = clockShortBreak3
	case clockPomodoro4:
		s.ClockType = clockLongBreak
	case clockShortBreak:
		s.ClockType = clockPomodoro2
	case clockShortBreak2:
		s.ClockType = clockPomodoro3
	case clockShortBreak3:
		s.ClockType = clockPomodoro4
	case clockLongBreak:
		s.ClockType = clockPomodoro
	}
}

func (s *State) finish() {
	s.cycleClock()

	soundText := ""
	switch s.ClockType {
	case clockShortBreak:
		soundText = "/home/neuromante/.local/share/sounds/chinese-gong-daniel_simon_2.wav"
	case clockShortBreak2:
		soundText = "/home/neuromante/.local/share/sounds/chinese-gong-daniel_simon_2.wav"
	case clockShortBreak3:
		soundText = "/home/neuromante/.local/share/sounds/chinese-gong-daniel_simon_2.wav"
	case clockPomodoro:
		soundText = "/home/neuromante/.local/share/sounds/476871_8152631-lq-2.ogg"
	case clockPomodoro2:
		soundText = "/home/neuromante/.local/share/sounds/411089_5121236-lq-2.ogg"
	case clockPomodoro3:
		soundText = "/home/neuromante/.local/share/sounds/411089_5121236-lq-2.ogg"
	case clockPomodoro4:
		soundText = "/home/neuromante/.local/share/sounds/411089_5121236-lq-2.ogg"
	case clockLongBreak:
		soundText = "/home/neuromante/.local/share/sounds/chinese-gong-daniel_simon_2.wav"
	}

	exec.Command("dunstify", "--appname", "Pomodoro", "--urgency", "2", fmt.Sprintf("%s!", s.clockText())).Start()
	exec.Command("mpv", "--no-terminal", "--no-video", soundText).Start()
}

func (s *State) output() {
	icon := iconPause
	if s.Running && !s.Paused {
		icon = iconRun
	}

	pomodoro := fmt.Sprintf("%s %s %s", s.clockTextShort(), icon, s.Duration.Round(time.Second))
	fmt.Println(pomodoro)
	fmt.Println(pomodoro)
}

func (s *State) write() error {
	content, err := json.Marshal(s)
	if err != nil {
		return errors.Wrap(err, "while marshaling json")
	}

	err = ioutil.WriteFile(stateFile, content, 0600)
	if err != nil {
		return errors.Wrap(err, "while writing state file")
	}
	return nil
}

func loadState() (*State, error) {
	s := &State{}
	content, err := ioutil.ReadFile(stateFile)
	if os.IsNotExist(err) {
		return s, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "while reading state file")
	}

	if err := json.Unmarshal(content, &s); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling json")
	}

	return s, nil
}

func (s *State) reset() {
	s.Running = false
	s.Paused = false
	s.LastTime = s.Now

	switch s.ClockType {
	case clockPomodoro, clockPomodoro2, clockPomodoro3, clockPomodoro4:
		s.Duration = defaultTimerLength
	case clockLongBreak:
		s.Duration = defaultLongBreakLength
	case clockShortBreak, clockShortBreak2, clockShortBreak3:
		s.Duration = defaultShortBreakLength
	}

	s.write()
}

func main() {
	s, err := loadState()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s.Now = time.Now()
	if s.LastTime.IsZero() {
		s.reset()
	}

	switch os.Getenv("BLOCK_BUTTON") {
	case "1":
		if s.Running {
			s.Paused = !s.Paused
		} else {
			s.Running = true
		}
		s.LastTime = s.Now
	case "2":
		s.reset()
	case "3":
		s.cycleClock()
		s.reset()
	}

	if s.Running && !s.Paused {
		s.Duration -= s.Now.Sub(s.LastTime)
	}

	s.LastTime = s.Now

	if s.Duration < 0 {
		s.output()
		s.finish()
		s.reset()
	} else {
		s.output()
	}

	if s.Running {
		s.write()
	}
}
