package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jakeChal/gotune/pkg/audio"
	"github.com/jakeChal/gotune/pkg/config"
	"github.com/jakeChal/gotune/pkg/dsp"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	sampleRate        = 48000
	bufferSize        = 4096
	silenceThreshold  = 0.001
	minDisplayClarity = 0.3
)

var (
	pipeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // Gray
	inTuneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
	closeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	sharpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // Red (too high)
	flatStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue (too low)
	noteStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))

	// New styles
	listeningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Italic(true)
	helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Faint(true)
	containerStyle = lipgloss.NewStyle().Padding(1, 2)
	perfectStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) // Bright green
)

// current state
type model struct {
	frequency float64
	noteName  string
	cents     float64
	hasPitch  bool
	profile   config.InstrumentProfile
	audioChan chan pitchMsg
}

// event carrying new data (to be copied into model)
type pitchMsg struct {
	frequency float64
	noteName  string
	cents     float64
	hasPitch  bool
}

func listenForAudio(audioChan chan pitchMsg) tea.Cmd {
	return func() tea.Msg {
		return <-audioChan // Block until data arrives
	}
}

func initialModel(profile config.InstrumentProfile, audioChan chan pitchMsg) model {
	return model{
		profile:   profile,
		audioChan: audioChan,
	}
}

func renderHeader(profile config.InstrumentProfile) string {
	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		BorderForeground(lipgloss.Color("14")).
		Align(lipgloss.Center)

	text := fmt.Sprintf("%s Tuner\n%.0f - %.0f Hz",
		profile.Name, profile.MinFreq, profile.MaxFreq)

	return headerStyle.Render(text)
}

func renderFooter() string {
	return helpStyle.Render("q: quit  •  ctrl+c: exit")
}

func (m model) Init() tea.Cmd {
	return listenForAudio(m.audioChan)
}

func (m model) View() string {
	header := renderHeader(m.profile)

	var body string
	if !m.hasPitch {
		listening := listeningStyle.Render("♪ Listening for notes...")
		body = lipgloss.Place(50, 1, lipgloss.Center, lipgloss.Center, listening)
	} else {
		body = formatTuningDisplay(m.frequency, m.noteName, m.cents)
	}

	footer := renderFooter()

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", body, "", footer)
	return containerStyle.Render(content)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case pitchMsg:
		m.frequency = msg.frequency
		m.noteName = msg.noteName
		m.cents = msg.cents
		m.hasPitch = msg.hasPitch
		return m, listenForAudio(m.audioChan)
	}

	return m, nil
}

func processAudio(ai *audio.AudioInput, audioChan chan pitchMsg, profile config.InstrumentProfile) {
	threshold := 0.1
	bufferAccum := make([]float64, 0, bufferSize)

	for frame := range ai.Frames {
		frame64 := dsp.Float32ToFloat64(frame)
		bufferAccum = append(bufferAccum, frame64...)

		if len(bufferAccum) >= bufferSize {
			energy := dsp.CalculateRMS(bufferAccum[:bufferSize])
			if energy < silenceThreshold {
				bufferAccum = bufferAccum[bufferSize:]
				continue
			}

			result := dsp.DetectPitch(bufferAccum[:bufferSize], sampleRate, threshold)

			// display
			if result.HasPitch &&
				result.Clarity >= minDisplayClarity &&
				result.Frequency >= profile.MinFreq &&
				result.Frequency <= profile.MaxFreq {
				noteName, _, centsOff := dsp.PitchToNote(result.Frequency)

				newData := pitchMsg{
					frequency: result.Frequency,
					noteName:  noteName,
					cents:     centsOff,
					hasPitch:  true,
				}
				audioChan <- newData
			}
			// Keep overflow samples for better continuity
			bufferAccum = bufferAccum[bufferSize:]

		}
	}
}

func formatTuningDisplay(freq float64, noteName string, cents float64) string {
	meter := makeMeter(cents)
	status := tuningStatus(cents)

	// Format the note name width FIRST, then style it
	formattedNote := fmt.Sprintf("%-4s", noteName)
	styledNote := noteStyle.Render(formattedNote)

	display := fmt.Sprintf("%s %s %s | %+6.1f¢ (%7.1f Hz)", styledNote, meter, status, cents, freq)

	// Add celebration when perfectly in tune (within 1 cent)
	absCents := cents
	if absCents < 0 {
		absCents = -absCents
	}
	if absCents <= 1.0 {
		display = perfectStyle.Render("★ ") + display + perfectStyle.Render(" ★")
	}

	return display
}

func makeMeter(cents float64) string {
	absCents := cents
	if absCents < 0 {
		absCents = -absCents
	}

	// Choose bracket color based on tuning state
	var bracketStyle lipgloss.Style
	if absCents <= 2 {
		bracketStyle = inTuneStyle // Green brackets when in tune
	} else if cents < 0 {
		bracketStyle = flatStyle // Blue brackets when flat
	} else {
		bracketStyle = sharpStyle // Red brackets when sharp
	}

	leftBracket := bracketStyle.Render("[")
	rightBracket := bracketStyle.Render("]")

	// cents: -50 to +50 range
	if cents < -20 {
		return leftBracket + flatStyle.Render("<<<<") + pipeStyle.Render("|") + "    " + rightBracket
	}
	if cents < -10 {
		return leftBracket + " " + flatStyle.Render("<<<") + pipeStyle.Render("|") + "    " + rightBracket
	}
	if cents < -5 {
		return leftBracket + "  " + flatStyle.Render("<<") + pipeStyle.Render("|") + "    " + rightBracket
	}
	if cents < -2 {
		return leftBracket + "   " + flatStyle.Render("<") + pipeStyle.Render("|") + "    " + rightBracket
	}
	if cents <= 2 {
		return leftBracket + "    " + pipeStyle.Render("|") + "    " + rightBracket // Centered, in tune
	}
	if cents <= 5 {
		return leftBracket + "    " + pipeStyle.Render("|") + sharpStyle.Render(">") + "   " + rightBracket
	}
	if cents <= 10 {
		return leftBracket + "    " + pipeStyle.Render("|") + sharpStyle.Render(">>") + "  " + rightBracket
	}
	if cents <= 20 {
		return leftBracket + "    " + pipeStyle.Render("|") + sharpStyle.Render(">>>") + " " + rightBracket
	}
	return leftBracket + "    " + pipeStyle.Render("|") + sharpStyle.Render(">>>>") + rightBracket
}

func tuningStatus(cents float64) string {
	absCents := cents
	if absCents < 0 {
		absCents = -absCents
	}

	if absCents <= 2 {
		return inTuneStyle.Render("✓ IN TUNE ")
	} else if absCents <= 5 {
		return closeStyle.Render(" ~ close  ")
	} else if absCents <= 10 {
		return closeStyle.Render("  adjust  ")
	} else if cents < 0 {
		return flatStyle.Render("  too low ")
	} else {
		return sharpStyle.Render(" too high ")
	}
}

func main() {
	// threshold := flag.Float64("t", 0.1, "The MPM algorithm's detection threshold [0 - 1.0]. Low values increase the sensitivity.")
	instrumentName := flag.String("i", "guitar", "Instrument profile (guitar, bouzouki).")
	flag.Parse()

	profile, ok := config.Profiles[*instrumentName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown instrument: %s\n", *instrumentName)
		fmt.Fprintf(os.Stderr, "Supported instruments: %v\n", config.ListInstruments())
		os.Exit(1)
	}

	audioChan := make(chan pitchMsg)
	ai, err := audio.NewAudioInput(sampleRate)
	if err != nil {
		panic(err)
	}

	if err := ai.Start(); err != nil {
		panic(err)
	}
	defer ai.Stop()
	go processAudio(ai, audioChan, profile)

	model := initialModel(profile, audioChan)

	p := tea.NewProgram(model)

	// if *threshold < 0 || *threshold > 1.0 {
	// 	fmt.Fprintf(os.Stderr, "threshold: %v is out of range!\n", *threshold)
	// 	os.Exit(1)
	// }

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
