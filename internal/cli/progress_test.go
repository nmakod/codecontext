package cli

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// safeWriter is a thread-safe writer wrapper for testing
type safeWriter struct {
	writer io.Writer
	mutex  *sync.Mutex
}

func (sw *safeWriter) Write(p []byte) (n int, err error) {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	return sw.writer.Write(p)
}

func TestNewProgressBar(t *testing.T) {
	config := &ProgressConfig{
		Width:       30,
		ShowPercent: true,
		ShowRate:    true,
		ShowETA:     true,
		UpdateRate:  50 * time.Millisecond,
	}

	pb := NewProgressBar(100, "Test Progress", config)

	if pb == nil {
		t.Fatal("ProgressBar should not be nil")
	}

	if pb.total != 100 {
		t.Errorf("Expected total 100, got %d", pb.total)
	}

	if pb.description != "Test Progress" {
		t.Errorf("Expected description 'Test Progress', got %s", pb.description)
	}

	if pb.width != 30 {
		t.Errorf("Expected width 30, got %d", pb.width)
	}
}

func TestProgressBar_BasicOperations(t *testing.T) {
	var buf bytes.Buffer

	pb := NewProgressBar(10, "Test", nil)
	pb.SetWriter(&buf)

	// Test initial state
	current, total, percent := pb.GetProgress()
	if current != 0 || total != 10 || percent != 0 {
		t.Errorf("Expected (0, 10, 0), got (%d, %d, %.1f)", current, total, percent)
	}

	// Test increment
	pb.Increment()
	current, total, percent = pb.GetProgress()
	if current != 1 || percent != 10 {
		t.Errorf("Expected current=1, percent=10, got current=%d, percent=%.1f", current, percent)
	}

	// Test add
	pb.Add(3)
	current, _, _ = pb.GetProgress()
	if current != 4 {
		t.Errorf("Expected current=4, got %d", current)
	}

	// Test set current
	pb.SetCurrent(8)
	current, _, _ = pb.GetProgress()
	if current != 8 {
		t.Errorf("Expected current=8, got %d", current)
	}

	// Test finish
	if pb.IsFinished() {
		t.Error("Progress bar should not be finished yet")
	}

	pb.Finish()
	if !pb.IsFinished() {
		t.Error("Progress bar should be finished")
	}

	current, _, percent = pb.GetProgress()
	if current != 10 || percent != 100 {
		t.Errorf("Expected current=10, percent=100, got current=%d, percent=%.1f", current, percent)
	}
}

func TestProgressBar_SetMethods(t *testing.T) {
	pb := NewProgressBar(50, "Initial", nil)

	// Test SetTotal
	pb.SetTotal(100)
	_, total, _ := pb.GetProgress()
	if total != 100 {
		t.Errorf("Expected total=100, got %d", total)
	}

	// Test SetDescription
	pb.SetDescription("Updated Description")
	if pb.description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got %s", pb.description)
	}

	// Test SetCustomSuffix
	pb.SetCustomSuffix("(custom)")
	if pb.customSuffix != "(custom)" {
		t.Errorf("Expected custom suffix '(custom)', got %s", pb.customSuffix)
	}
}

func TestProgressBar_Reset(t *testing.T) {
	pb := NewProgressBar(10, "Test", nil)

	// Make some progress
	pb.SetCurrent(5)
	current, _, _ := pb.GetProgress()
	if current != 5 {
		t.Errorf("Expected current=5, got %d", current)
	}

	// Reset
	pb.Reset()
	current, _, _ = pb.GetProgress()
	if current != 0 {
		t.Errorf("Expected current=0 after reset, got %d", current)
	}

	if pb.IsFinished() {
		t.Error("Progress bar should not be finished after reset")
	}
}

func TestProgressBar_Overflow(t *testing.T) {
	pb := NewProgressBar(10, "Test", nil)

	// Try to set current beyond total
	pb.SetCurrent(15)
	current, total, _ := pb.GetProgress()
	if current != total {
		t.Errorf("Current should be clamped to total, got current=%d, total=%d", current, total)
	}

	// Try to add beyond total
	pb.Reset()
	pb.Add(15)
	current, total, _ = pb.GetProgress()
	if current != total {
		t.Errorf("Current should be clamped to total after add, got current=%d, total=%d", current, total)
	}
}

func TestProgressBar_Rendering(t *testing.T) {
	var buf bytes.Buffer

	config := &ProgressConfig{
		Width:       10,
		ShowPercent: true,
		ShowRate:    false,
		ShowETA:     false,
		UpdateRate:  time.Nanosecond, // Force immediate updates
	}

	pb := NewProgressBar(10, "Test", config)
	pb.SetWriter(&buf)

	// Test rendering at 50%
	pb.SetCurrent(5)

	output := buf.String()
	if !strings.Contains(output, "Test") {
		t.Error("Output should contain description")
	}

	if !strings.Contains(output, "50.0%") {
		t.Error("Output should contain percentage")
	}

	if !strings.Contains(output, "5/10") {
		t.Error("Output should contain current/total")
	}
}

func TestNewMultiProgressBar(t *testing.T) {
	mpb := NewMultiProgressBar()

	if mpb == nil {
		t.Fatal("MultiProgressBar should not be nil")
	}

	if len(mpb.bars) != 0 {
		t.Error("Initial bars slice should be empty")
	}
}

func TestMultiProgressBar_AddBar(t *testing.T) {
	mpb := NewMultiProgressBar()

	bar1 := mpb.AddBar(100, "Task 1", nil)
	if bar1 == nil {
		t.Fatal("Added bar should not be nil")
	}

	bar2 := mpb.AddBar(50, "Task 2", nil)
	if bar2 == nil {
		t.Fatal("Added bar should not be nil")
	}

	if len(mpb.bars) != 2 {
		t.Errorf("Expected 2 bars, got %d", len(mpb.bars))
	}

	if mpb.height != 2 {
		t.Errorf("Expected height 2, got %d", mpb.height)
	}
}

func TestMultiProgressBar_StartStop(t *testing.T) {
	mpb := NewMultiProgressBar()

	bar := mpb.AddBar(10, "Test Task", nil)

	// Start
	mpb.Start()
	if !mpb.active {
		t.Error("MultiProgressBar should be active after start")
	}

	// Make some progress
	bar.SetCurrent(5)

	// Give it time to render
	time.Sleep(150 * time.Millisecond)

	// Stop
	mpb.Stop()
	if mpb.active {
		t.Error("MultiProgressBar should not be active after stop")
	}
}

func TestNewSpinner(t *testing.T) {
	spinner := NewSpinner("Loading...")

	if spinner == nil {
		t.Fatal("Spinner should not be nil")
	}

	if spinner.message != "Loading..." {
		t.Errorf("Expected message 'Loading...', got %s", spinner.message)
	}

	if len(spinner.chars) == 0 {
		t.Error("Spinner should have characters")
	}
}

func TestSpinner_StartStop(t *testing.T) {
	var buf bytes.Buffer
	var bufMutex sync.Mutex

	spinner := NewSpinner("Test Message")
	
	// Create a thread-safe writer wrapper
	safeWriter := &safeWriter{writer: &buf, mutex: &bufMutex}
	spinner.writer = safeWriter

	// Start spinner
	spinner.Start()
	if !spinner.IsActive() {
		t.Error("Spinner should be active after start")
	}

	// Let it spin for a bit
	time.Sleep(250 * time.Millisecond)

	// Stop spinner
	spinner.Stop()
	if spinner.IsActive() {
		t.Error("Spinner should not be active after stop")
	}

	// Check that something was written
	bufMutex.Lock()
	output := buf.String()
	bufMutex.Unlock()
	
	if len(output) == 0 {
		t.Error("Spinner should have produced output")
	}
}

func TestSpinner_SetMessage(t *testing.T) {
	spinner := NewSpinner("Initial")

	if spinner.message != "Initial" {
		t.Errorf("Expected initial message 'Initial', got %s", spinner.message)
	}

	spinner.SetMessage("Updated")
	if spinner.message != "Updated" {
		t.Errorf("Expected updated message 'Updated', got %s", spinner.message)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m30s"},
		{3661 * time.Second, "1h1m"},
		{7200 * time.Second, "2h0m"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %s, expected %s", tt.duration, result, tt.expected)
		}
	}
}

func TestNewProgressManager(t *testing.T) {
	pm := NewProgressManager()

	if pm == nil {
		t.Fatal("ProgressManager should not be nil")
	}

	if pm.multiBar == nil {
		t.Error("MultiProgressBar should be initialized")
	}
}

func TestProgressManager_FileScanning(t *testing.T) {
	pm := NewProgressManager()

	bar := pm.StartFileScanning(100)
	if bar == nil {
		t.Fatal("File scanning progress bar should not be nil")
	}

	if pm.mode != "scanning" {
		t.Errorf("Expected mode 'scanning', got %s", pm.mode)
	}

	// Make some progress
	bar.SetCurrent(50)

	current, total, percent := bar.GetProgress()
	if current != 50 || total != 100 || percent != 50 {
		t.Errorf("Expected (50, 100, 50), got (%d, %d, %.1f)", current, total, percent)
	}

	pm.Stop()
}

func TestProgressManager_Parsing(t *testing.T) {
	pm := NewProgressManager()

	bar := pm.StartParsing(200)
	if bar == nil {
		t.Fatal("Parsing progress bar should not be nil")
	}

	// Test progress
	bar.Increment()
	current, _, _ := bar.GetProgress()
	if current != 1 {
		t.Errorf("Expected current=1, got %d", current)
	}

	pm.Stop()
}

func TestProgressManager_Analyzing(t *testing.T) {
	pm := NewProgressManager()

	bar := pm.StartAnalyzing(500)
	if bar == nil {
		t.Fatal("Analyzing progress bar should not be nil")
	}

	// Test progress
	bar.Add(100)
	current, _, _ := bar.GetProgress()
	if current != 100 {
		t.Errorf("Expected current=100, got %d", current)
	}

	pm.Stop()
}

func TestProgressManager_Indeterminate(t *testing.T) {
	pm := NewProgressManager()

	pm.StartIndeterminate("Processing...")
	if pm.mode != "indeterminate" {
		t.Errorf("Expected mode 'indeterminate', got %s", pm.mode)
	}

	if pm.spinner == nil {
		t.Error("Spinner should be initialized")
	}

	// Update message
	pm.UpdateIndeterminate("Still processing...")

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	pm.Stop()
}

func TestProgressManager_MultipleOperations(t *testing.T) {
	pm := NewProgressManager()

	// Start file scanning
	scanBar := pm.StartFileScanning(1000)
	scanBar.SetCurrent(500)

	// Add parsing
	parseBar := pm.StartParsing(800)
	parseBar.SetCurrent(200)

	// Add analyzing
	analyzeBar := pm.StartAnalyzing(500)
	analyzeBar.SetCurrent(100)

	// Let them render
	time.Sleep(300 * time.Millisecond)

	// Verify progress
	current, _, _ := scanBar.GetProgress()
	if current != 500 {
		t.Errorf("Scan bar: expected current=500, got %d", current)
	}

	current, _, _ = parseBar.GetProgress()
	if current != 200 {
		t.Errorf("Parse bar: expected current=200, got %d", current)
	}

	current, _, _ = analyzeBar.GetProgress()
	if current != 100 {
		t.Errorf("Analyze bar: expected current=100, got %d", current)
	}

	pm.Stop()
}

// Benchmark tests

func BenchmarkProgressBar_Increment(b *testing.B) {
	pb := NewProgressBar(int64(b.N), "Benchmark", &ProgressConfig{
		UpdateRate: time.Hour, // Prevent rendering during benchmark
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pb.Increment()
	}
}

func BenchmarkProgressBar_SetCurrent(b *testing.B) {
	pb := NewProgressBar(1000000, "Benchmark", &ProgressConfig{
		UpdateRate: time.Hour, // Prevent rendering during benchmark
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pb.SetCurrent(int64(i % 1000000))
	}
}

func BenchmarkSpinner_Render(b *testing.B) {
	var buf bytes.Buffer
	spinner := NewSpinner("Benchmark")
	spinner.writer = &buf

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spinner.spin()
	}
}
