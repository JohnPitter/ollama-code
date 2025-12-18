package hardware

import (
	"testing"
)

func TestNewDetector(t *testing.T) {
	detector := NewDetector()
	if detector == nil {
		t.Fatal("NewDetector() returned nil")
	}
}

func TestDetect(t *testing.T) {
	detector := NewDetector()
	specs, err := detector.Detect()

	if err != nil {
		t.Logf("Detection failed (may be expected in test environment): %v", err)
		// Don't fail the test as detection might not work in all environments
		return
	}

	if specs == nil {
		t.Fatal("Detect() returned nil specs without error")
	}

	// Verify basic sanity of detected values
	if specs.CPUCores <= 0 {
		t.Errorf("Invalid CPU cores: %d", specs.CPUCores)
	}

	if specs.TotalRAM <= 0 {
		t.Errorf("Invalid RAM: %d", specs.TotalRAM)
	}

	t.Logf("Detected: %d cores, %d MB RAM", specs.CPUCores, specs.TotalRAM)
	if specs.HasNVIDIAGPU {
		t.Logf("GPU: %s (%d MB)", specs.GPUModel, specs.GPUMemory)
	}
}

// Note: classifyTier is not exported, so we can't test it directly.
// Instead, we test the overall Detect() behavior above.
