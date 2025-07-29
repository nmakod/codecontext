package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// ProgressBar represents a CLI progress bar
type ProgressBar struct {
	writer       io.Writer
	total        int64
	current      int64
	width        int
	description  string
	showPercent  bool
	showRate     bool
	showETA      bool
	startTime    time.Time
	lastUpdate   time.Time
	updateRate   time.Duration
	mutex        sync.RWMutex
	finished     bool
	template     string
	customSuffix string
}

// ProgressConfig holds configuration for progress bars
type ProgressConfig struct {
	Width       int           `json:"width"`        // Progress bar width
	ShowPercent bool          `json:"show_percent"` // Show percentage
	ShowRate    bool          `json:"show_rate"`    // Show processing rate
	ShowETA     bool          `json:"show_eta"`     // Show estimated time to completion
	UpdateRate  time.Duration `json:"update_rate"`  // How often to update display
	Template    string        `json:"template"`     // Custom template
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int64, description string, config *ProgressConfig) *ProgressBar {
	if config == nil {
		config = &ProgressConfig{
			Width:       50,
			ShowPercent: true,
			ShowRate:    true,
			ShowETA:     true,
			UpdateRate:  100 * time.Millisecond,
		}
	}

	return &ProgressBar{
		writer:      os.Stderr,
		total:       total,
		width:       config.Width,
		description: description,
		showPercent: config.ShowPercent,
		showRate:    config.ShowRate,
		showETA:     config.ShowETA,
		startTime:   time.Now(),
		lastUpdate:  time.Now(),
		updateRate:  config.UpdateRate,
		template:    config.Template,
	}
}

// SetWriter sets the output writer for the progress bar
func (pb *ProgressBar) SetWriter(writer io.Writer) {
	pb.mutex.Lock()
	pb.writer = writer
	pb.mutex.Unlock()
}

// SetTotal updates the total count
func (pb *ProgressBar) SetTotal(total int64) {
	pb.mutex.Lock()
	pb.total = total
	pb.mutex.Unlock()
}

// SetDescription updates the description
func (pb *ProgressBar) SetDescription(description string) {
	pb.mutex.Lock()
	pb.description = description
	pb.mutex.Unlock()
}

// SetCustomSuffix sets a custom suffix to display
func (pb *ProgressBar) SetCustomSuffix(suffix string) {
	pb.mutex.Lock()
	pb.customSuffix = suffix
	pb.mutex.Unlock()
}

// Increment increments the progress by 1
func (pb *ProgressBar) Increment() {
	pb.Add(1)
}

// Add adds the specified amount to the progress
func (pb *ProgressBar) Add(amount int64) {
	pb.mutex.Lock()
	pb.current += amount
	if pb.current > pb.total {
		pb.current = pb.total
	}
	pb.mutex.Unlock()

	pb.maybeRender()
}

// SetCurrent sets the current progress value
func (pb *ProgressBar) SetCurrent(current int64) {
	pb.mutex.Lock()
	pb.current = current
	if pb.current > pb.total {
		pb.current = pb.total
	}
	pb.mutex.Unlock()

	pb.maybeRender()
}

// Finish completes the progress bar
func (pb *ProgressBar) Finish() {
	pb.mutex.Lock()
	pb.current = pb.total
	pb.finished = true
	pb.mutex.Unlock()

	pb.render()
	fmt.Fprintln(pb.writer) // Add newline
}

// Reset resets the progress bar
func (pb *ProgressBar) Reset() {
	pb.mutex.Lock()
	pb.current = 0
	pb.finished = false
	pb.startTime = time.Now()
	pb.lastUpdate = time.Now()
	pb.mutex.Unlock()
}

// GetProgress returns current progress information
func (pb *ProgressBar) GetProgress() (current, total int64, percent float64) {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()

	current = pb.current
	total = pb.total
	if total > 0 {
		percent = float64(current) / float64(total) * 100
	}
	return
}

// IsFinished returns whether the progress bar is finished
func (pb *ProgressBar) IsFinished() bool {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()
	return pb.finished
}

// Private methods

func (pb *ProgressBar) maybeRender() {
	pb.mutex.RLock()
	shouldUpdate := time.Since(pb.lastUpdate) >= pb.updateRate || pb.current == pb.total
	pb.mutex.RUnlock()

	if shouldUpdate {
		pb.render()
	}
}

func (pb *ProgressBar) render() {
	pb.mutex.RLock()
	defer pb.mutex.RUnlock()

	if pb.template != "" {
		pb.renderCustom()
	} else {
		pb.renderDefault()
	}
}

func (pb *ProgressBar) renderDefault() {
	// Calculate percentage
	percent := float64(0)
	if pb.total > 0 {
		percent = float64(pb.current) / float64(pb.total) * 100
	}

	// Create progress bar
	filled := int(float64(pb.width) * percent / 100)
	if filled > pb.width {
		filled = pb.width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.width-filled)

	// Build output string
	output := fmt.Sprintf("\r%s [%s]", pb.description, bar)

	if pb.showPercent {
		output += fmt.Sprintf(" %.1f%%", percent)
	}

	output += fmt.Sprintf(" %d/%d", pb.current, pb.total)

	if pb.showRate {
		rate := pb.calculateRate()
		output += fmt.Sprintf(" (%.1f/s)", rate)
	}

	if pb.showETA && !pb.finished {
		eta := pb.calculateETA()
		if eta > 0 {
			output += fmt.Sprintf(" ETA: %s", formatDuration(eta))
		}
	}

	if pb.customSuffix != "" {
		output += " " + pb.customSuffix
	}

	fmt.Fprint(pb.writer, output)
	pb.lastUpdate = time.Now()
}

func (pb *ProgressBar) renderCustom() {
	// Custom template rendering would go here
	// For now, fall back to default
	pb.renderDefault()
}

func (pb *ProgressBar) calculateRate() float64 {
	elapsed := time.Since(pb.startTime).Seconds()
	if elapsed > 0 {
		return float64(pb.current) / elapsed
	}
	return 0
}

func (pb *ProgressBar) calculateETA() time.Duration {
	if pb.current == 0 {
		return 0
	}

	elapsed := time.Since(pb.startTime)
	rate := float64(pb.current) / elapsed.Seconds()
	remaining := pb.total - pb.current

	if rate > 0 {
		return time.Duration(float64(remaining)/rate) * time.Second
	}

	return 0
}

// MultiProgressBar manages multiple progress bars
type MultiProgressBar struct {
	bars      []*ProgressBar
	writer    io.Writer
	mutex     sync.RWMutex
	height    int
	active    bool
	lastLines int
}

// NewMultiProgressBar creates a new multi-progress bar manager
func NewMultiProgressBar() *MultiProgressBar {
	return &MultiProgressBar{
		bars:   make([]*ProgressBar, 0),
		writer: os.Stderr,
		height: 0,
	}
}

// AddBar adds a progress bar to the multi-progress display
func (mpb *MultiProgressBar) AddBar(total int64, description string, config *ProgressConfig) *ProgressBar {
	bar := NewProgressBar(total, description, config)

	mpb.mutex.Lock()
	mpb.bars = append(mpb.bars, bar)
	mpb.height++
	mpb.mutex.Unlock()

	// Redirect bar output to prevent individual rendering
	bar.SetWriter(io.Discard)

	return bar
}

// Start begins the multi-progress bar display
func (mpb *MultiProgressBar) Start() {
	mpb.mutex.Lock()
	mpb.active = true
	mpb.mutex.Unlock()

	go mpb.renderLoop()
}

// Stop stops the multi-progress bar display
func (mpb *MultiProgressBar) Stop() {
	mpb.mutex.Lock()
	mpb.active = false
	mpb.mutex.Unlock()

	// Final render and cleanup
	mpb.renderAll()
	fmt.Fprintln(mpb.writer)
}

// Private methods for MultiProgressBar

func (mpb *MultiProgressBar) renderLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		mpb.mutex.RLock()
		active := mpb.active
		mpb.mutex.RUnlock()

		if !active {
			break
		}

		select {
		case <-ticker.C:
			mpb.renderAll()
		}
	}
}

func (mpb *MultiProgressBar) renderAll() {
	mpb.mutex.Lock()
	defer mpb.mutex.Unlock()

	// Clear previous lines
	if mpb.lastLines > 0 {
		for i := 0; i < mpb.lastLines; i++ {
			fmt.Fprint(mpb.writer, "\033[A\033[K") // Move up and clear line
		}
	}

	lines := 0
	for _, bar := range mpb.bars {
		if !bar.IsFinished() || bar.IsFinished() {
			// Render each bar
			bar.mutex.RLock()
			percent := float64(0)
			if bar.total > 0 {
				percent = float64(bar.current) / float64(bar.total) * 100
			}

			filled := int(float64(bar.width) * percent / 100)
			if filled > bar.width {
				filled = bar.width
			}

			barStr := strings.Repeat("█", filled) + strings.Repeat("░", bar.width-filled)

			output := fmt.Sprintf("%s [%s] %.1f%% %d/%d",
				bar.description, barStr, percent, bar.current, bar.total)

			if bar.showRate {
				rate := bar.calculateRate()
				output += fmt.Sprintf(" (%.1f/s)", rate)
			}

			if bar.customSuffix != "" {
				output += " " + bar.customSuffix
			}

			bar.mutex.RUnlock()

			fmt.Fprintln(mpb.writer, output)
			lines++
		}
	}

	mpb.lastLines = lines
}

// Spinner represents a simple spinner for indeterminate progress
type Spinner struct {
	writer      io.Writer
	chars       []rune
	delay       time.Duration
	message     string
	active      bool
	mutex       sync.RWMutex
	stopChan    chan bool
	currentChar int
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		writer:   os.Stderr,
		chars:    []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'},
		delay:    100 * time.Millisecond,
		message:  message,
		stopChan: make(chan bool, 1),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mutex.Lock()
	s.active = true
	s.mutex.Unlock()

	go s.spin()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	s.mutex.Lock()
	if s.active {
		s.active = false
		select {
		case s.stopChan <- true:
		default:
		}
	}
	s.mutex.Unlock()

	// Clear the line
	fmt.Fprint(s.writer, "\r\033[K")
}

// IsActive returns whether the spinner is currently active
func (s *Spinner) IsActive() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.active
}

// SetMessage updates the spinner message
func (s *Spinner) SetMessage(message string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.message = message
	
	// Immediately redraw with new message to avoid text artifacts
	if s.active {
		fmt.Fprint(s.writer, "\r\033[K") // Clear the line
		char := s.chars[s.currentChar]
		fmt.Fprintf(s.writer, "%c %s", char, message)
	}
}

// Private methods for Spinner

func (s *Spinner) spin() {
	ticker := time.NewTicker(s.delay)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.mutex.RLock()
			if !s.active {
				s.mutex.RUnlock()
				return
			}

			char := s.chars[s.currentChar]
			message := s.message
			s.mutex.RUnlock()

			// Clear the line and redraw with current message
			fmt.Fprint(s.writer, "\r\033[K")
			fmt.Fprintf(s.writer, "%c %s", char, message)

			s.mutex.Lock()
			s.currentChar = (s.currentChar + 1) % len(s.chars)
			s.mutex.Unlock()
		}
	}
}

// Utility functions

// formatDuration formats a duration in human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) - 60*minutes
		if seconds == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) - 60*hours
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
}

// ProgressManager manages all progress indicators for the CLI
type ProgressManager struct {
	multiBar *MultiProgressBar
	spinner  *Spinner
	mode     string
	mutex    sync.RWMutex
}

// NewProgressManager creates a new progress manager
func NewProgressManager() *ProgressManager {
	return &ProgressManager{
		multiBar: NewMultiProgressBar(),
	}
}

// StartFileScanning starts progress indication for file scanning
func (pm *ProgressManager) StartFileScanning(totalFiles int64) *ProgressBar {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.mode = "scanning"
	pm.multiBar.Start()

	config := &ProgressConfig{
		Width:       40,
		ShowPercent: true,
		ShowRate:    true,
		ShowETA:     true,
		UpdateRate:  200 * time.Millisecond,
	}

	return pm.multiBar.AddBar(totalFiles, "Scanning files", config)
}

// StartParsing starts progress indication for parsing
func (pm *ProgressManager) StartParsing(totalFiles int64) *ProgressBar {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	config := &ProgressConfig{
		Width:       40,
		ShowPercent: true,
		ShowRate:    true,
		ShowETA:     true,
		UpdateRate:  200 * time.Millisecond,
	}

	return pm.multiBar.AddBar(totalFiles, "Parsing files", config)
}

// StartAnalyzing starts progress indication for analysis
func (pm *ProgressManager) StartAnalyzing(totalSymbols int64) *ProgressBar {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	config := &ProgressConfig{
		Width:       40,
		ShowPercent: true,
		ShowRate:    true,
		ShowETA:     true,
		UpdateRate:  200 * time.Millisecond,
	}

	return pm.multiBar.AddBar(totalSymbols, "Analyzing symbols", config)
}

// StartIndeterminate starts an indeterminate progress indicator
func (pm *ProgressManager) StartIndeterminate(message string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.mode = "indeterminate"
	pm.spinner = NewSpinner(message)
	pm.spinner.Start()
}

// UpdateIndeterminate updates the indeterminate progress message
func (pm *ProgressManager) UpdateIndeterminate(message string) {
	pm.mutex.RLock()
	spinner := pm.spinner
	pm.mutex.RUnlock()

	if spinner != nil {
		spinner.SetMessage(message)
	}
}

// Stop stops all progress indicators
func (pm *ProgressManager) Stop() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.multiBar != nil {
		pm.multiBar.Stop()
	}

	if pm.spinner != nil {
		pm.spinner.Stop()
	}
}
