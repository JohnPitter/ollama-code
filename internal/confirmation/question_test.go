package confirmation

import (
	"testing"
)

func TestQuestion_Validate(t *testing.T) {
	tests := []struct {
		name    string
		question Question
		wantErr error
	}{
		{
			name: "valid question",
			question: Question{
				Question: "Which library should we use?",
				Header:   "Library",
				Options: []Option{
					{Label: "React", Description: "Popular UI library"},
					{Label: "Vue", Description: "Progressive framework"},
				},
				MultiSelect: false,
			},
			wantErr: nil,
		},
		{
			name: "empty question",
			question: Question{
				Question: "",
				Options: []Option{
					{Label: "Option 1"},
					{Label: "Option 2"},
				},
			},
			wantErr: ErrEmptyQuestion,
		},
		{
			name: "header too long",
			question: Question{
				Question: "Which one?",
				Header:   "ThisHeaderIsTooLong",
				Options: []Option{
					{Label: "Option 1"},
					{Label: "Option 2"},
				},
			},
			wantErr: ErrHeaderTooLong,
		},
		{
			name: "too few options",
			question: Question{
				Question: "Which one?",
				Options: []Option{
					{Label: "Only one"},
				},
			},
			wantErr: ErrInvalidOptionCount,
		},
		{
			name: "too many options",
			question: Question{
				Question: "Which one?",
				Options: []Option{
					{Label: "Option 1"},
					{Label: "Option 2"},
					{Label: "Option 3"},
					{Label: "Option 4"},
					{Label: "Option 5"},
				},
			},
			wantErr: ErrInvalidOptionCount,
		},
		{
			name: "empty option label",
			question: Question{
				Question: "Which one?",
				Options: []Option{
					{Label: "Option 1"},
					{Label: ""}, // Empty label
				},
			},
			wantErr: ErrEmptyOptionLabel,
		},
		{
			name: "max valid options (4)",
			question: Question{
				Question: "Choose features to enable",
				Header:   "Features",
				Options: []Option{
					{Label: "Auth", Description: "Authentication"},
					{Label: "DB", Description: "Database"},
					{Label: "Cache", Description: "Caching"},
					{Label: "Logs", Description: "Logging"},
				},
				MultiSelect: true,
			},
			wantErr: nil,
		},
		{
			name: "min valid options (2)",
			question: Question{
				Question: "Continue?",
				Options: []Option{
					{Label: "Yes", Description: "Proceed"},
					{Label: "No", Description: "Cancel"},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.question.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQuestionSet_Validate(t *testing.T) {
	validQuestion := Question{
		Question: "Which one?",
		Options: []Option{
			{Label: "Option 1"},
			{Label: "Option 2"},
		},
	}

	tests := []struct {
		name    string
		qs      QuestionSet
		wantErr error
	}{
		{
			name: "valid question set",
			qs: QuestionSet{
				Questions: []Question{validQuestion},
			},
			wantErr: nil,
		},
		{
			name: "empty question set",
			qs: QuestionSet{
				Questions: []Question{},
			},
			wantErr: ErrInvalidQuestionCount,
		},
		{
			name: "too many questions (5)",
			qs: QuestionSet{
				Questions: []Question{
					validQuestion,
					validQuestion,
					validQuestion,
					validQuestion,
					validQuestion,
				},
			},
			wantErr: ErrInvalidQuestionCount,
		},
		{
			name: "max valid questions (4)",
			qs: QuestionSet{
				Questions: []Question{
					validQuestion,
					validQuestion,
					validQuestion,
					validQuestion,
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid question in set",
			qs: QuestionSet{
				Questions: []Question{
					{
						Question: "", // Invalid
						Options: []Option{
							{Label: "A"},
							{Label: "B"},
						},
					},
				},
			},
			wantErr: ErrEmptyQuestion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.qs.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAnswer_Structure(t *testing.T) {
	// Test single select answer
	singleAnswer := &Answer{
		Question:      "Which library?",
		SelectedLabel: "React",
	}

	if singleAnswer.SelectedLabel != "React" {
		t.Errorf("Expected SelectedLabel 'React', got '%s'", singleAnswer.SelectedLabel)
	}

	// Test multi select answer
	multiAnswer := &Answer{
		Question:       "Which features?",
		SelectedLabels: []string{"Auth", "DB", "Cache"},
	}

	if len(multiAnswer.SelectedLabels) != 3 {
		t.Errorf("Expected 3 selected labels, got %d", len(multiAnswer.SelectedLabels))
	}

	// Test custom input answer
	customAnswer := &Answer{
		Question:    "Other framework?",
		CustomInput: "Svelte",
	}

	if customAnswer.CustomInput != "Svelte" {
		t.Errorf("Expected CustomInput 'Svelte', got '%s'", customAnswer.CustomInput)
	}
}

func TestOption_Structure(t *testing.T) {
	option := Option{
		Label:       "TypeScript",
		Description: "Typed superset of JavaScript",
	}

	if option.Label != "TypeScript" {
		t.Errorf("Expected Label 'TypeScript', got '%s'", option.Label)
	}

	if option.Description != "Typed superset of JavaScript" {
		t.Errorf("Expected description, got '%s'", option.Description)
	}
}
