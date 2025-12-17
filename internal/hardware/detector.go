package hardware

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Specs especificações de hardware detectadas
type Specs struct {
	// CPU
	CPUCores      int    `json:"cpu_cores"`
	CPUThreads    int    `json:"cpu_threads"`
	CPUModel      string `json:"cpu_model"`

	// Memory
	TotalRAM      int64  `json:"total_ram_mb"`
	AvailableRAM  int64  `json:"available_ram_mb"`

	// GPU
	HasNVIDIAGPU  bool   `json:"has_nvidia_gpu"`
	GPUModel      string `json:"gpu_model"`
	GPUMemory     int64  `json:"gpu_memory_mb"`
	GPUCount      int    `json:"gpu_count"`

	// Storage
	DiskSpace     int64  `json:"disk_space_gb"`

	// OS
	OS            string `json:"os"`
	Arch          string `json:"arch"`
}

// Detector detector de hardware
type Detector struct{}

// NewDetector cria novo detector
func NewDetector() *Detector {
	return &Detector{}
}

// Detect detecta especificações de hardware
func (d *Detector) Detect() (*Specs, error) {
	specs := &Specs{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// CPU
	specs.CPUCores = runtime.NumCPU()
	specs.CPUThreads = runtime.NumCPU()
	specs.CPUModel = d.detectCPUModel()

	// Memory
	specs.TotalRAM, specs.AvailableRAM = d.detectMemory()

	// GPU
	specs.HasNVIDIAGPU, specs.GPUModel, specs.GPUMemory, specs.GPUCount = d.detectGPU()

	// Disk
	specs.DiskSpace = d.detectDiskSpace()

	return specs, nil
}

// detectCPUModel detecta modelo da CPU
func (d *Detector) detectCPUModel() string {
	switch runtime.GOOS {
	case "linux":
		output, err := exec.Command("lscpu").Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Model name:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						return strings.TrimSpace(parts[1])
					}
				}
			}
		}

	case "darwin":
		output, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}

	case "windows":
		output, err := exec.Command("wmic", "cpu", "get", "name").Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 1 {
				return strings.TrimSpace(lines[1])
			}
		}
	}

	return "Unknown CPU"
}

// detectMemory detecta memória RAM
func (d *Detector) detectMemory() (total, available int64) {
	switch runtime.GOOS {
	case "linux":
		// Total
		output, err := exec.Command("grep", "MemTotal", "/proc/meminfo").Output()
		if err == nil {
			fields := strings.Fields(string(output))
			if len(fields) >= 2 {
				val, _ := strconv.ParseInt(fields[1], 10, 64)
				total = val / 1024 // Convert KB to MB
			}
		}

		// Available
		output, err = exec.Command("grep", "MemAvailable", "/proc/meminfo").Output()
		if err == nil {
			fields := strings.Fields(string(output))
			if len(fields) >= 2 {
				val, _ := strconv.ParseInt(fields[1], 10, 64)
				available = val / 1024
			}
		}

	case "darwin":
		// Total
		output, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err == nil {
			val, _ := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
			total = val / 1024 / 1024 // Convert bytes to MB
		}

		// Available - aproximado
		available = total * 70 / 100 // Assume 70% disponível

	case "windows":
		// Total
		output, err := exec.Command("wmic", "computersystem", "get", "TotalPhysicalMemory").Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 1 {
				val, _ := strconv.ParseInt(strings.TrimSpace(lines[1]), 10, 64)
				total = val / 1024 / 1024
			}
		}

		// Available
		output, err = exec.Command("wmic", "OS", "get", "FreePhysicalMemory").Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 1 {
				val, _ := strconv.ParseInt(strings.TrimSpace(lines[1]), 10, 64)
				available = val / 1024
			}
		}
	}

	return total, available
}

// detectGPU detecta GPU NVIDIA
func (d *Detector) detectGPU() (hasGPU bool, model string, memory int64, count int) {
	// Tentar nvidia-smi
	output, err := exec.Command("nvidia-smi", "--query-gpu=name,memory.total,count", "--format=csv,noheader").Output()
	if err != nil {
		return false, "No NVIDIA GPU", 0, 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return false, "No NVIDIA GPU", 0, 0
	}

	// Parse primeira GPU
	firstLine := lines[0]
	parts := strings.Split(firstLine, ",")

	if len(parts) >= 2 {
		model = strings.TrimSpace(parts[0])

		// Parse memory (formato: "16384 MiB")
		memStr := strings.TrimSpace(parts[1])
		memStr = strings.ReplaceAll(memStr, " MiB", "")
		memStr = strings.ReplaceAll(memStr, " MB", "")
		memory, _ = strconv.ParseInt(memStr, 10, 64)

		hasGPU = true
		count = len(lines)
	}

	return hasGPU, model, memory, count
}

// detectDiskSpace detecta espaço em disco
func (d *Detector) detectDiskSpace() int64 {
	switch runtime.GOOS {
	case "linux", "darwin":
		output, err := exec.Command("df", "-BG", "/").Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 1 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 4 {
					spaceStr := strings.TrimSuffix(fields[3], "G")
					space, _ := strconv.ParseInt(spaceStr, 10, 64)
					return space
				}
			}
		}

	case "windows":
		output, err := exec.Command("wmic", "logicaldisk", "get", "size,freespace").Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 1 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 1 {
					space, _ := strconv.ParseInt(fields[0], 10, 64)
					return space / 1024 / 1024 / 1024 // Bytes to GB
				}
			}
		}
	}

	return 0
}

// String retorna representação em string
func (s *Specs) String() string {
	return fmt.Sprintf(`Hardware Detected:
  CPU: %s (%d cores / %d threads)
  RAM: %d MB total, %d MB available
  GPU: %s (NVIDIA: %v, %d MB VRAM, %d GPU(s))
  Disk: %d GB available
  OS: %s/%s`,
		s.CPUModel, s.CPUCores, s.CPUThreads,
		s.TotalRAM, s.AvailableRAM,
		s.GPUModel, s.HasNVIDIAGPU, s.GPUMemory, s.GPUCount,
		s.DiskSpace,
		s.OS, s.Arch)
}

// GetPerformanceTier retorna tier de performance baseado no hardware
func (s *Specs) GetPerformanceTier() string {
	// High-end: 32GB+ RAM, NVIDIA GPU 16GB+, 8+ cores
	if s.TotalRAM >= 32768 && s.HasNVIDIAGPU && s.GPUMemory >= 16384 && s.CPUCores >= 8 {
		return "high-end"
	}

	// Mid-range: 16GB+ RAM, NVIDIA GPU 8GB+, 4+ cores
	if s.TotalRAM >= 16384 && s.HasNVIDIAGPU && s.GPUMemory >= 8192 && s.CPUCores >= 4 {
		return "mid-range"
	}

	// Entry: 8GB+ RAM, ou GPU menor, ou sem GPU
	if s.TotalRAM >= 8192 {
		return "entry"
	}

	// Low-end
	return "low-end"
}
