# UX/UI Regression Test Suite

## Overview

Comprehensive UX/UI regression test suite for Ollama Code, covering all major functionalities and user interaction patterns.

## Test Structure

```
tests/ux/
├── 01_conversation_flow/      # Conversational UX (greetings, courtesy)
├── 02_intent_detection/       # Intent classification accuracy
├── 03_file_operations/        # File read/write/modify
├── 04_code_generation/        # Code generation across languages
├── 05_web_search/            # Web search functionality
├── 06_git_operations/        # Git commands (future)
├── 07_error_handling/        # Graceful error handling
├── logs/                     # Test execution logs
└── run_all_ux_tests.sh       # Master test runner
```

## Test Categories

### 1. Conversation Flow (01_conversation_flow/)

Tests natural language interaction patterns:
- **Greetings**: "oi", "olá", "hello"
- **Thanks**: "obrigado", "valeu", "thanks"
- **Confirmations**: "ok", "certo", "entendi"
- **State**: "estou bem", "tudo certo"
- **Farewells**: "tchau", "até logo", "bye"

**Purpose**: Ensure courtesy messages are correctly classified as `question` intent, NOT `web_search`.

### 2. Intent Detection (02_intent_detection/)

Tests all 8 intent types:
1. `read_file` - Read/analyze/explain files
2. `write_file` - Create/modify/refactor code
3. `execute_command` - Run shell commands
4. `search_code` - Search within project
5. `analyze_project` - Project structure analysis
6. `git_operation` - Git commands
7. `web_search` - Search online information
8. `question` - Conceptual questions + courtesy

**Purpose**: Validate intent classification accuracy across all categories.

### 3. File Operations (03_file_operations/)

Tests file system operations:
- Creating files (txt, py, html, css, js, go)
- Reading existing files
- Modifying file contents
- Validation of file creation

**Purpose**: Ensure file operations work correctly and safely.

### 4. Code Generation (04_code_generation/)

Tests code generation across languages:
- **Python**: Scripts, servers, algorithms
- **JavaScript/HTML/CSS**: Web pages, styling, validation
- **Go**: Programs, APIs, file handlers
- **Multi-file**: Coordinated file generation

**Purpose**: Validate code generation quality and multi-file support.

### 5. Web Search (05_web_search/)

Tests web search capabilities:
- Real-time information (weather, news)
- Documentation search (React, Go, Python)
- General information queries
- Source citation formatting

**Purpose**: Ensure web search works and formats results properly.

### 6. Git Operations (06_git_operations/)

Tests Git integration (future):
- Git status
- Git commit
- Git diff
- Branch operations

**Purpose**: Validate Git command execution.

### 7. Error Handling (07_error_handling/)

Tests graceful error handling:
- File not found scenarios
- Invalid/empty input
- Ambiguous requests
- No crashes or panics

**Purpose**: Ensure user-friendly error messages and no crashes.

## Running Tests

### Run All Tests

```bash
cd tests/ux
chmod +x run_all_ux_tests.sh
./run_all_ux_tests.sh
```

### Run Individual Test Suite

```bash
cd tests/ux/01_conversation_flow
chmod +x test_greetings.sh
./test_greetings.sh
```

### Run Specific Category

```bash
bash tests/ux/02_intent_detection/test_all_intents.sh
```

## Test Output

Tests generate:
- **Console output**: Real-time pass/fail for each test
- **Individual logs**: `logs/<category>_<test_name>.log`
- **Consolidated report**: `logs/ux_test_report_<timestamp>.log`

### Example Output

```
==========================================
OLLAMA CODE - UX/UI REGRESSION TEST SUITE
==========================================

Running: 01_conversation_flow/test_greetings
----------------------------------------
[GREET-01] Simple greeting 'oi'... PASS
[GREET-02] Simple greeting 'olá'... PASS
[THANKS-01] Simple thanks... PASS
✓ test_greetings PASSED

==========================================
FINAL SUMMARY
==========================================

Test Suites:
  Passed: 6
  Failed: 0
  Total: 6

Individual Tests:
  Passed: 52
  Failed: 0
  Total: 52

Success Rate: 100.0%

✓✓✓ ALL UX TESTS PASSED! ✓✓✓
```

## Test Metrics

Target metrics:
- **Intent Detection Accuracy**: ≥ 95%
- **File Operations Success**: 100%
- **Code Generation Quality**: ≥ 90%
- **Error Handling**: 100% (no crashes)

## Adding New Tests

### 1. Create Test Script

```bash
# Create in appropriate category
touch tests/ux/0X_category/test_new_feature.sh
chmod +x tests/ux/0X_category/test_new_feature.sh
```

### 2. Use Template

```bash
#!/usr/bin/env bash
set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

OLLAMA_CODE="../../../build/ollama-code.exe"
PASSED=0
FAILED=0

test_feature() {
    local test_id="$1"
    local message="$2"
    local validation="$3"
    local description="$4"

    echo -n "[$test_id] $description... "

    output=$(echo "$message" | timeout 30s $OLLAMA_CODE ask --mode autonomous 2>&1 || true)

    if eval "$validation"; then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC}"
        ((FAILED++))
    fi
}

# Add test cases
test_feature "TEST-01" "test message" "validation command" "Description"

# Summary
echo "Passed: $PASSED"
echo "Failed: $FAILED"
[ $FAILED -eq 0 ] && exit 0 || exit 1
```

### 3. Add to Master Script

Edit `run_all_ux_tests.sh` and add to `TESTS` array:

```bash
TESTS=(
    ...
    "0X_category/test_new_feature.sh"
)
```

## Continuous Integration

These tests can be integrated into CI/CD:

```yaml
# .github/workflows/ux-tests.yml
name: UX Regression Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Build
        run: go build -o build/ollama-code ./cmd/ollama-code
      - name: Run UX Tests
        run: |
          cd tests/ux
          ./run_all_ux_tests.sh
```

## Troubleshooting

### Common Issues

1. **Timeout errors**: Increase timeout in individual test (default 30s)
2. **Model not responding**: Check Ollama service is running
3. **File permissions**: Ensure scripts are executable (`chmod +x`)
4. **Windows compatibility**: Tests use bash, run via Git Bash or WSL

### Debug Mode

Add `-x` to test script for verbose output:

```bash
bash -x tests/ux/01_conversation_flow/test_greetings.sh
```

## Contributing

When adding features:
1. Add corresponding UX tests
2. Ensure tests pass before PR
3. Update this README if adding new category
4. Maintain ≥95% success rate

## License

Same as Ollama Code project.
