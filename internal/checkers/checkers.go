package checkers

import (
	"embed"
	"encoding/base64"
	"fmt"
	"strings"
)

//go:embed src/*
var checkerFiles embed.FS

// PrepareCheckerSource prepends testlib.h, removes #include "testlib.h", and base64 encodes the result.
func PrepareCheckerSource(name string, customCode string) (string, error) {
	var checkerCode string
	if name == "custom" {
		checkerCode = customCode
	} else {
		if !strings.HasSuffix(name, ".cpp") {
			name = name + ".cpp"
		}
		checkerBytes, err := checkerFiles.ReadFile("src/" + name)
		if err != nil {
			return "", fmt.Errorf("checker not found: %s", name)
		}
		checkerCode = string(checkerBytes)
	}

	testlibBytes, err := checkerFiles.ReadFile("src/testlib.h")
	if err != nil {
		return "", fmt.Errorf("testlib.h not found")
	}
	testlibCode := string(testlibBytes)

	// Remove #include "testlib.h" to prevent compile errors on judge node
	checkerCode = strings.ReplaceAll(checkerCode, "#include \"testlib.h\"", "")
	checkerCode = strings.ReplaceAll(checkerCode, "#include <testlib.h>", "")

	finalCode := testlibCode + "\n\n" + checkerCode
	
	// Base64 encode for safe transmission
	return base64.StdEncoding.EncodeToString([]byte(finalCode)), nil
}
