package diff

import (
	"strings"
	"testing"
)

func TestParseRange(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantStart int
		wantEnd   int
		wantErr   bool
	}{
		{
			name:      "valid range",
			input:     "10:20",
			wantStart: 10,
			wantEnd:   20,
			wantErr:   false,
		},
		{
			name:      "single line",
			input:     "5:5",
			wantStart: 5,
			wantEnd:   5,
			wantErr:   false,
		},
		{
			name:      "start of file",
			input:     "1:10",
			wantStart: 1,
			wantEnd:   10,
			wantErr:   false,
		},
		{
			name:    "invalid format",
			input:   "10-20",
			wantErr: true,
		},
		{
			name:    "invalid numbers",
			input:   "abc:def",
			wantErr: true,
		},
		{
			name:    "zero line",
			input:   "0:10",
			wantErr: true,
		},
		{
			name:    "negative line",
			input:   "-1:10",
			wantErr: true,
		},
		{
			name:    "start > end",
			input:   "20:10",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRange(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseRange() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseRange() unexpected error: %v", err)
				return
			}

			if got.Start != tt.wantStart {
				t.Errorf("ParseRange() Start = %d, want %d", got.Start, tt.wantStart)
			}

			if got.End != tt.wantEnd {
				t.Errorf("ParseRange() End = %d, want %d", got.End, tt.wantEnd)
			}
		})
	}
}

func TestDiffer_ComputeDiff(t *testing.T) {
	differ := NewDiffer()

	tests := []struct {
		name           string
		oldContent     string
		newContent     string
		expectedChanges int
	}{
		{
			name:            "no changes",
			oldContent:      "line1\nline2\nline3",
			newContent:      "line1\nline2\nline3",
			expectedChanges: 0,
		},
		{
			name:            "one line modified",
			oldContent:      "line1\nline2\nline3",
			newContent:      "line1\nmodified\nline3",
			expectedChanges: 1,
		},
		{
			name:            "one line added",
			oldContent:      "line1\nline2",
			newContent:      "line1\nline2\nline3",
			expectedChanges: 1,
		},
		{
			name:            "one line deleted",
			oldContent:      "line1\nline2\nline3",
			newContent:      "line1\nline2",
			expectedChanges: 1,
		},
		{
			name:            "multiple changes",
			oldContent:      "line1\nline2\nline3",
			newContent:      "modified1\nline2\nmodified3\nline4",
			expectedChanges: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := differ.ComputeDiff("test.txt", tt.oldContent, tt.newContent)

			if len(diff.Changes) != tt.expectedChanges {
				t.Errorf("Expected %d changes, got %d", tt.expectedChanges, len(diff.Changes))
			}

			if diff.FilePath != "test.txt" {
				t.Errorf("Expected FilePath 'test.txt', got '%s'", diff.FilePath)
			}

			if diff.OldContent != tt.oldContent {
				t.Errorf("OldContent mismatch")
			}

			if diff.NewContent != tt.newContent {
				t.Errorf("NewContent mismatch")
			}
		})
	}
}

func TestDiffer_ApplyEdit(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		editRange   EditRange
		wantContent string
		wantErr     bool
	}{
		{
			name:    "edit single line",
			content: "line1\nline2\nline3\nline4",
			editRange: EditRange{
				Start: 2,
				End:   2,
				Text:  "modified line 2",
			},
			wantContent: "line1\nmodified line 2\nline3\nline4",
			wantErr:     false,
		},
		{
			name:    "edit multiple lines",
			content: "line1\nline2\nline3\nline4\nline5",
			editRange: EditRange{
				Start: 2,
				End:   4,
				Text:  "new line\nanother new line",
			},
			wantContent: "line1\nnew line\nanother new line\nline5",
			wantErr:     false,
		},
		{
			name:    "delete lines (empty text)",
			content: "line1\nline2\nline3\nline4",
			editRange: EditRange{
				Start: 2,
				End:   3,
				Text:  "",
			},
			wantContent: "line1\nline4",
			wantErr:     false,
		},
		{
			name:    "edit first line",
			content: "line1\nline2\nline3",
			editRange: EditRange{
				Start: 1,
				End:   1,
				Text:  "new first line",
			},
			wantContent: "new first line\nline2\nline3",
			wantErr:     false,
		},
		{
			name:    "edit last line",
			content: "line1\nline2\nline3",
			editRange: EditRange{
				Start: 3,
				End:   3,
				Text:  "new last line",
			},
			wantContent: "line1\nline2\nnew last line",
			wantErr:     false,
		},
		{
			name:    "start line out of range",
			content: "line1\nline2",
			editRange: EditRange{
				Start: 5,
				End:   5,
				Text:  "test",
			},
			wantErr: true,
		},
		{
			name:    "end line out of range",
			content: "line1\nline2",
			editRange: EditRange{
				Start: 1,
				End:   10,
				Text:  "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			differ := NewDiffer()

			newContent, diff, err := differ.ApplyEdit("test.txt", tt.content, tt.editRange)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ApplyEdit() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ApplyEdit() unexpected error: %v", err)
				return
			}

			if newContent != tt.wantContent {
				t.Errorf("ApplyEdit() content mismatch\nGot:\n%s\n\nWant:\n%s", newContent, tt.wantContent)
			}

			if diff == nil {
				t.Error("ApplyEdit() diff is nil")
				return
			}

			if diff.OldContent != tt.content {
				t.Error("Diff OldContent mismatch")
			}

			if diff.NewContent != newContent {
				t.Error("Diff NewContent mismatch")
			}
		})
	}
}

func TestDiffer_Rollback(t *testing.T) {
	differ := NewDiffer()

	// Aplicar primeira edição
	content1 := "line1\nline2\nline3"
	newContent1, _, _ := differ.ApplyEdit("test.txt", content1, EditRange{
		Start: 2,
		End:   2,
		Text:  "modified line 2",
	})

	// Aplicar segunda edição
	_, _, _ = differ.ApplyEdit("test.txt", newContent1, EditRange{
		Start: 1,
		End:   1,
		Text:  "modified line 1",
	})

	// Rollback - deve retornar ao newContent1
	rolled, err := differ.Rollback("test.txt")
	if err != nil {
		t.Fatalf("Rollback() error: %v", err)
	}

	if rolled != newContent1 {
		t.Errorf("Rollback() content mismatch\nGot:\n%s\n\nWant:\n%s", rolled, newContent1)
	}

	// Rollback novamente - deve retornar ao content1
	rolled2, err := differ.Rollback("test.txt")
	if err != nil {
		t.Fatalf("Second Rollback() error: %v", err)
	}

	if rolled2 != content1 {
		t.Errorf("Second Rollback() content mismatch\nGot:\n%s\n\nWant:\n%s", rolled2, content1)
	}

	// Rollback sem histórico - deve retornar erro
	_, err = differ.Rollback("test.txt")
	if err == nil {
		t.Error("Rollback() without history should return error")
	}
}

func TestDiffer_History(t *testing.T) {
	differ := NewDiffer()

	// Adicionar edições
	differ.ApplyEdit("file1.txt", "content1", EditRange{Start: 1, End: 1, Text: "new"})
	differ.ApplyEdit("file2.txt", "content2", EditRange{Start: 1, End: 1, Text: "new"})
	differ.ApplyEdit("file1.txt", "content3", EditRange{Start: 1, End: 1, Text: "new"})

	// Verificar histórico total
	allHistory := differ.GetHistory("")
	if len(allHistory) != 3 {
		t.Errorf("Expected 3 history entries, got %d", len(allHistory))
	}

	// Verificar histórico de file1.txt
	file1History := differ.GetHistory("file1.txt")
	if len(file1History) != 2 {
		t.Errorf("Expected 2 history entries for file1.txt, got %d", len(file1History))
	}

	// Verificar histórico de file2.txt
	file2History := differ.GetHistory("file2.txt")
	if len(file2History) != 1 {
		t.Errorf("Expected 1 history entry for file2.txt, got %d", len(file2History))
	}

	// Limpar histórico
	differ.ClearHistory()
	if len(differ.GetHistory("")) != 0 {
		t.Error("Expected empty history after clear")
	}
}

func TestDiffer_ChangeTypes(t *testing.T) {
	differ := NewDiffer()

	// Test add
	diff := differ.ComputeDiff("test.txt", "line1", "line1\nline2")
	if len(diff.Changes) != 1 || diff.Changes[0].Type != ChangeAdd {
		t.Error("Expected ChangeAdd")
	}

	// Test delete
	diff = differ.ComputeDiff("test.txt", "line1\nline2", "line1")
	if len(diff.Changes) != 1 || diff.Changes[0].Type != ChangeDelete {
		t.Error("Expected ChangeDelete")
	}

	// Test modify
	diff = differ.ComputeDiff("test.txt", "line1", "modified")
	if len(diff.Changes) != 1 || diff.Changes[0].Type != ChangeModify {
		t.Error("Expected ChangeModify")
	}
}

func TestPreviewer_Preview(t *testing.T) {
	previewer := NewPreviewer()
	differ := NewDiffer()

	diff := differ.ComputeDiff("test.txt", "line1\nline2", "line1\nmodified")

	preview := previewer.Preview(diff)

	if !strings.Contains(preview, "test.txt") {
		t.Error("Preview should contain filename")
	}

	if !strings.Contains(preview, "MODIFICADA") {
		t.Error("Preview should contain change type")
	}
}

func TestPreviewer_PreviewRange(t *testing.T) {
	previewer := NewPreviewer()

	content := "line1\nline2\nline3\nline4\nline5"
	editRange := EditRange{
		Start: 2,
		End:   3,
		Text:  "new line",
	}

	preview := previewer.PreviewRange("test.txt", content, editRange)

	if !strings.Contains(preview, "test.txt") {
		t.Error("Preview should contain filename")
	}

	if !strings.Contains(preview, "line2") {
		t.Error("Preview should show lines to remove")
	}

	if !strings.Contains(preview, "new line") {
		t.Error("Preview should show new text")
	}
}

func TestPreviewer_CompactPreview(t *testing.T) {
	previewer := NewPreviewer()
	differ := NewDiffer()

	// Diff com adição, modificação e deleção
	diff := differ.ComputeDiff("test.txt",
		"line1\nline2",
		"modified1\nline2\nline3")

	compact := previewer.CompactPreview(diff)

	if !strings.Contains(compact, "test.txt") {
		t.Error("Compact preview should contain filename")
	}

	// line1 -> modified1 = modify
	// line2 = sem mudança
	// line3 = add
	if !strings.Contains(compact, "+1") {
		t.Error("Compact preview should show additions")
	}

	if !strings.Contains(compact, "~1") {
		t.Error("Compact preview should show modifications")
	}
}
