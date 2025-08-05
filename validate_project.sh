#!/bin/bash

# Project validation script - can be used in CI/CD pipelines
# This performs basic validation checks on the project structure and files

echo "🔍 Validating project structure and files..."

EXIT_CODE=0

# Check required files exist
required_files=("main.go" "go.mod" "README.md" "Makefile")
for file in "${required_files[@]}"; do
    if [[ -f "$file" ]]; then
        echo "✅ $file exists"
    else
        echo "❌ Required file missing: $file"
        EXIT_CODE=1
    fi
done

# Check Go module is valid
if go mod verify; then
    echo "✅ Go module verification passed"
else
    echo "❌ Go module verification failed"
    EXIT_CODE=1
fi

# Check if main.go compiles
if go build -o /tmp/test_build main.go; then
    echo "✅ main.go compiles successfully"
    rm -f /tmp/test_build
else
    echo "❌ main.go compilation failed"
    EXIT_CODE=1
fi

# Check README.md is not just placeholder
if [[ -f "README.md" ]]; then
    readme_content=$(cat README.md | tr -d '\n\r\t ' | tr '[:upper:]' '[:lower:]')
    if [[ "$readme_content" == "ardfgdsg" ]]; then
        echo "⚠️  README.md contains only placeholder content"
        # Don't fail for this, just warn
    elif [[ ${#readme_content} -lt 10 ]]; then
        echo "⚠️  README.md content is very short"
    else
        echo "✅ README.md has meaningful content"
    fi
fi

# Check test files exist
if ls *_test.go 1> /dev/null 2>&1; then
    echo "✅ Test files found"
    test_count=$(ls *_test.go | wc -l)
    echo "   📊 Number of test files: $test_count"
else
    echo "❌ No test files found (*_test.go)"
    EXIT_CODE=1
fi

exit $EXIT_CODE