package diff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2E_CompleteEditWorkflow testa o fluxo completo de edição
func TestE2E_CompleteEditWorkflow(t *testing.T) {
	// Setup: criar arquivo temporário
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.go")

	initialContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`

	// Escrever conteúdo inicial
	err := os.WriteFile(filePath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Criar differ e previewer
	differ := NewDiffer()
	previewer := NewPreviewer()

	// Step 1: Editar linha 5 (mudar mensagem)
	editRange1 := EditRange{
		Start: 6,
		End:   6,
		Text:  `	fmt.Println("Hello, Ollama Code!")`,
	}

	newContent1, diff1, err := differ.ApplyEdit(filePath, initialContent, editRange1)
	if err != nil {
		t.Fatalf("Step 1: ApplyEdit failed: %v", err)
	}

	if !strings.Contains(newContent1, "Hello, Ollama Code!") {
		t.Errorf("Step 1: Content not updated correctly")
	}

	// Verificar preview
	preview1 := previewer.Preview(diff1)
	if !strings.Contains(preview1, "MODIFICADA") {
		t.Errorf("Step 1: Preview should contain 'MODIFICADA'")
	}

	// Escrever no arquivo
	err = os.WriteFile(filePath, []byte(newContent1), 0644)
	if err != nil {
		t.Fatalf("Step 1: Failed to write file: %v", err)
	}

	// Step 2: Adicionar nova linha (import)
	editRange2 := EditRange{
		Start: 3,
		End:   3,
		Text:  `import "time"`,
	}

	newContent2, _, err := differ.ApplyEdit(filePath, newContent1, editRange2)
	if err != nil {
		t.Fatalf("Step 2: ApplyEdit failed: %v", err)
	}

	if !strings.Contains(newContent2, `import "time"`) {
		t.Errorf("Step 2: Content not updated correctly")
	}

	// Escrever no arquivo
	err = os.WriteFile(filePath, []byte(newContent2), 0644)
	if err != nil {
		t.Fatalf("Step 2: Failed to write file: %v", err)
	}

	// Step 3: Verificar histórico
	historySlice := differ.GetHistory(filePath)

	if len(historySlice) != 2 {
		t.Errorf("Step 3: Expected 2 history entries, got %d", len(historySlice))
	}

	// Step 4: Rollback primeira edição (volta para newContent1)
	rolledContent1, err := differ.Rollback(filePath)
	if err != nil {
		t.Fatalf("Step 4: Rollback failed: %v", err)
	}

	if rolledContent1 != newContent1 {
		t.Errorf("Step 4: Rollback content mismatch")
	}

	// Escrever no arquivo após rollback
	err = os.WriteFile(filePath, []byte(rolledContent1), 0644)
	if err != nil {
		t.Fatalf("Step 4: Failed to write file after rollback: %v", err)
	}

	// Step 5: Rollback segunda edição (volta para initialContent)
	rolledContent2, err := differ.Rollback(filePath)
	if err != nil {
		t.Fatalf("Step 5: Rollback failed: %v", err)
	}

	if rolledContent2 != initialContent {
		t.Errorf("Step 5: Rollback content mismatch, expected initial content")
	}

	// Verificar que arquivo voltou ao original
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Step 5: Failed to read file: %v", err)
	}

	if string(fileContent) != rolledContent1 {
		t.Errorf("Step 5: File content doesn't match rolled back content")
	}

	// Step 6: Tentar rollback sem histórico (deve falhar)
	_, err = differ.Rollback(filePath)
	if err == nil {
		t.Error("Step 6: Rollback without history should fail")
	}
}

// TestE2E_PreviewBeforeApply testa preview antes de aplicar edição
func TestE2E_PreviewBeforeApply(t *testing.T) {
	previewer := NewPreviewer()

	content := `line1
line2
line3
line4
line5`

	// Preview de range edit
	editRange := EditRange{
		Start: 2,
		End:   3,
		Text:  "new line",
	}

	preview := previewer.PreviewRange("test.txt", content, editRange)

	// Verificar que preview mostra contexto antes
	if !strings.Contains(preview, "line1") {
		t.Error("Preview should show context before")
	}

	// Verificar que preview mostra linhas a remover
	if !strings.Contains(preview, "line2") || !strings.Contains(preview, "line3") {
		t.Error("Preview should show lines to remove")
	}

	// Verificar que preview mostra novo texto
	if !strings.Contains(preview, "new line") {
		t.Error("Preview should show new text")
	}

	// Verificar que preview mostra contexto depois
	if !strings.Contains(preview, "line4") {
		t.Error("Preview should show context after")
	}
}

// TestE2E_MultipleFilesEditing testa edição de múltiplos arquivos
func TestE2E_MultipleFilesEditing(t *testing.T) {
	tmpDir := t.TempDir()

	// Criar 3 arquivos diferentes
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	file3 := filepath.Join(tmpDir, "file3.txt")

	content1 := "content1\nline2\nline3"
	content2 := "content2\nline2\nline3"
	content3 := "content3\nline2\nline3"

	os.WriteFile(file1, []byte(content1), 0644)
	os.WriteFile(file2, []byte(content2), 0644)
	os.WriteFile(file3, []byte(content3), 0644)

	differ := NewDiffer()

	// Editar cada arquivo
	editRange := EditRange{Start: 1, End: 1, Text: "MODIFIED"}

	_, _, err := differ.ApplyEdit(file1, content1, editRange)
	if err != nil {
		t.Fatalf("Edit file1 failed: %v", err)
	}

	_, _, err = differ.ApplyEdit(file2, content2, editRange)
	if err != nil {
		t.Fatalf("Edit file2 failed: %v", err)
	}

	_, _, err = differ.ApplyEdit(file3, content3, editRange)
	if err != nil {
		t.Fatalf("Edit file3 failed: %v", err)
	}

	// Verificar histórico total
	historySlice := differ.GetHistory("")

	if len(historySlice) != 3 {
		t.Errorf("Expected 3 history entries, got %d", len(historySlice))
	}

	// Verificar histórico por arquivo
	file1Slice := differ.GetHistory(file1)

	if len(file1Slice) != 1 {
		t.Errorf("Expected 1 history entry for file1, got %d", len(file1Slice))
	}

	// Rollback arquivo específico
	rolled, err := differ.Rollback(file2)
	if err != nil {
		t.Fatalf("Rollback file2 failed: %v", err)
	}

	if rolled != content2 {
		t.Error("Rollback didn't restore file2 content")
	}

	// Histórico total deve ter 2 agora
	history2Slice := differ.GetHistory("")

	if len(history2Slice) != 2 {
		t.Errorf("Expected 2 history entries after rollback, got %d", len(history2Slice))
	}
}

// TestE2E_CompactPreviewForLogs testa preview compacto
func TestE2E_CompactPreviewForLogs(t *testing.T) {
	differ := NewDiffer()
	previewer := NewPreviewer()

	oldContent := "line1\nline2\nline3"
	newContent := "line1\nmodified\nline3\nline4"

	diff := differ.ComputeDiff("test.txt", oldContent, newContent)

	compact := previewer.CompactPreview(diff)

	// Deve conter filename
	if !strings.Contains(compact, "test.txt") {
		t.Error("Compact preview should contain filename")
	}

	// Deve conter estatísticas (+1 ~1 -0)
	if !strings.Contains(compact, "+1") {
		t.Error("Compact preview should show additions")
	}

	if !strings.Contains(compact, "~1") {
		t.Error("Compact preview should show modifications")
	}
}
