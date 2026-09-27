package build

import "math"

// FlashSeconds is how long a reminder flashes after firing.
const FlashSeconds = 1.5

// ReminderState is an active reminder at a given time.
type ReminderState struct {
	Text      string
	Countdown int  // whole seconds until the next fire; -1 when there is none before until
	Flash     bool // fired less than FlashSeconds ago
}

// ReminderStatus returns the reminders active at game time t (t <= until),
// in file order. Reminders fire at every, 2*every, ... up to until.
func ReminderStatus(rs []Reminder, t float64) []ReminderState {
	var out []ReminderState
	for _, r := range rs {
		if t > float64(r.Until) {
			continue
		}
		every := float64(r.EverySeconds)
		st := ReminderState{Text: r.Text, Countdown: -1}
		last := math.Max(0, math.Floor(t/every)*every)
		if last >= every && t-last < FlashSeconds {
			st.Flash = true
		}
		next := last + every
		if next <= float64(r.Until) {
			st.Countdown = Countdown(next - t)
		}
		out = append(out, st)
	}
	return out
}
