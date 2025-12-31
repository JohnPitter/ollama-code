package multimodel

import "testing"

// TestTaskType_IsValid testa validação de task types
func TestTaskType_IsValid(t *testing.T) {
	testCases := []struct {
		name     string
		taskType TaskType
		expected bool
	}{
		{"Intent is valid", TaskTypeIntent, true},
		{"Code is valid", TaskTypeCode, true},
		{"Search is valid", TaskTypeSearch, true},
		{"Analysis is valid", TaskTypeAnalysis, true},
		{"Default is valid", TaskTypeDefault, true},
		{"Invalid type", TaskType("invalid"), false},
		{"Empty type", TaskType(""), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.taskType.IsValid()
			if result != tc.expected {
				t.Errorf("Expected IsValid()=%v for %s, got %v", tc.expected, tc.taskType, result)
			}
		})
	}
}

// TestTaskType_String testa conversão para string
func TestTaskType_String(t *testing.T) {
	testCases := []struct {
		taskType TaskType
		expected string
	}{
		{TaskTypeIntent, "intent"},
		{TaskTypeCode, "code"},
		{TaskTypeSearch, "search"},
		{TaskTypeAnalysis, "analysis"},
		{TaskTypeDefault, "default"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.taskType.String()
			if result != tc.expected {
				t.Errorf("Expected String()=%s, got %s", tc.expected, result)
			}
		})
	}
}

// TestTaskType_Constants testa que constantes têm valores corretos
func TestTaskType_Constants(t *testing.T) {
	if string(TaskTypeIntent) != "intent" {
		t.Errorf("TaskTypeIntent should be 'intent', got '%s'", TaskTypeIntent)
	}

	if string(TaskTypeCode) != "code" {
		t.Errorf("TaskTypeCode should be 'code', got '%s'", TaskTypeCode)
	}

	if string(TaskTypeSearch) != "search" {
		t.Errorf("TaskTypeSearch should be 'search', got '%s'", TaskTypeSearch)
	}

	if string(TaskTypeAnalysis) != "analysis" {
		t.Errorf("TaskTypeAnalysis should be 'analysis', got '%s'", TaskTypeAnalysis)
	}

	if string(TaskTypeDefault) != "default" {
		t.Errorf("TaskTypeDefault should be 'default', got '%s'", TaskTypeDefault)
	}
}

// TestModelSpec_Fields testa campos do ModelSpec
func TestModelSpec_Fields(t *testing.T) {
	spec := ModelSpec{
		Name:        "test-model",
		MaxTokens:   1024,
		Temperature: 0.7,
		Description: "Test model",
	}

	if spec.Name != "test-model" {
		t.Errorf("Expected Name='test-model', got '%s'", spec.Name)
	}

	if spec.MaxTokens != 1024 {
		t.Errorf("Expected MaxTokens=1024, got %d", spec.MaxTokens)
	}

	if spec.Temperature != 0.7 {
		t.Errorf("Expected Temperature=0.7, got %f", spec.Temperature)
	}

	if spec.Description != "Test model" {
		t.Errorf("Expected Description='Test model', got '%s'", spec.Description)
	}
}

// TestTaskType_AllValid testa que todos os tipos definidos são válidos
func TestTaskType_AllValid(t *testing.T) {
	allTypes := []TaskType{
		TaskTypeIntent,
		TaskTypeCode,
		TaskTypeSearch,
		TaskTypeAnalysis,
		TaskTypeDefault,
	}

	for _, taskType := range allTypes {
		if !taskType.IsValid() {
			t.Errorf("TaskType %s should be valid", taskType)
		}
	}
}
