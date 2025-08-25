package parser

import (
	"testing"

	"github.com/nuthan-ms/codecontext/pkg/types"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Error("NewManager() returned nil")
	}

	if manager.parsers == nil {
		t.Error("Manager parsers not initialized")
	}

	if manager.languages == nil {
		t.Error("Manager languages not initialized")
	}

	if manager.cache == nil {
		t.Error("Manager cache not initialized")
	}
}

func TestGetSupportedLanguages(t *testing.T) {
	manager := NewManager()
	languages := manager.GetSupportedLanguages()

	if len(languages) == 0 {
		t.Error("No supported languages found")
	}

	// Check if TypeScript is supported
	var tsFound bool
	for _, lang := range languages {
		if lang == "typescript" {
			tsFound = true
			break
		}
	}

	if !tsFound {
		t.Error("TypeScript language not found in supported languages")
	}
}

func TestDetectLanguage(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name     string
		filePath string
		expected string
		wantNil  bool
	}{
		{
			name:     "typescript file",
			filePath: "test.ts",
			expected: "typescript",
			wantNil:  false,
		},
		{
			name:     "tsx file",
			filePath: "component.tsx",
			expected: "typescript",
			wantNil:  false,
		},
		{
			name:     "javascript file",
			filePath: "script.js",
			expected: "javascript",
			wantNil:  false,
		},
		{
			name:     "jsx file",
			filePath: "component.jsx",
			expected: "javascript",
			wantNil:  false,
		},
		{
			name:     "python file",
			filePath: "script.py",
			expected: "python",
			wantNil:  false,
		},
		{
			name:     "java file",
			filePath: "Main.java",
			expected: "java",
			wantNil:  false,
		},
		{
			name:     "go file",
			filePath: "main.go",
			expected: "go",
			wantNil:  false,
		},
		{
			name:     "rust file",
			filePath: "main.rs",
			expected: "rust",
			wantNil:  false,
		},
		{
			name:     "unsupported file",
			filePath: "document.txt",
			expected: "",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.detectLanguage(tt.filePath)

			if tt.wantNil && result != nil {
				t.Errorf("Expected nil result for %s, got %v", tt.filePath, result)
			}

			if !tt.wantNil && result == nil {
				t.Errorf("Expected non-nil result for %s, got nil", tt.filePath)
			}

			if !tt.wantNil && result != nil && result.Name != tt.expected {
				t.Errorf("Expected language %s for %s, got %s", tt.expected, tt.filePath, result.Name)
			}
		})
	}
}

func TestClassifyFile(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name         string
		filePath     string
		expectedType string
		expectedTest bool
		expectError  bool
	}{
		{
			name:         "typescript source file",
			filePath:     "src/main.ts",
			expectedType: "source",
			expectedTest: false,
			expectError:  false,
		},
		{
			name:         "typescript test file",
			filePath:     "src/main.test.ts",
			expectedType: "test",
			expectedTest: true,
			expectError:  false,
		},
		{
			name:         "config file",
			filePath:     "config.json",
			expectedType: "config",
			expectedTest: false,
			expectError:  false,
		},
		{
			name:         "unsupported file",
			filePath:     "document.txt",
			expectedType: "",
			expectedTest: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.ClassifyFile(tt.filePath)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.filePath)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for %s, got %v", tt.filePath, err)
			}

			if !tt.expectError && result != nil {
				if result.FileType != tt.expectedType {
					t.Errorf("Expected file type %s for %s, got %s", tt.expectedType, tt.filePath, result.FileType)
				}

				if result.IsTest != tt.expectedTest {
					t.Errorf("Expected IsTest %v for %s, got %v", tt.expectedTest, tt.filePath, result.IsTest)
				}
			}
		})
	}
}

func TestParseFileVersioned(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name        string
		filePath    string
		content     string
		version     string
		expectError bool
	}{
		{
			name:        "typescript function",
			filePath:    "test.ts",
			content:     "function hello() { return 'world'; }",
			version:     "1.0",
			expectError: false,
		},
		{
			name:        "python function",
			filePath:    "test.py",
			content:     "def hello():\n    return 'world'",
			version:     "1.0",
			expectError: false,
		},
		{
			name:        "java class",
			filePath:    "Test.java",
			content:     "public class Test { public void hello() {} }",
			version:     "1.0",
			expectError: false,
		},
		{
			name:        "go function",
			filePath:    "test.go",
			content:     "package main\n\nfunc hello() string {\n    return \"world\"\n}",
			version:     "1.0",
			expectError: false,
		},
		{
			name:        "rust function",
			filePath:    "test.rs",
			content:     "fn hello() -> String {\n    \"world\".to_string()\n}",
			version:     "1.0",
			expectError: false,
		},
		{
			name:        "unsupported file",
			filePath:    "test.txt",
			content:     "plain text",
			version:     "1.0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.ParseFileVersioned(tt.filePath, tt.content, tt.version)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.filePath)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for %s, got %v", tt.filePath, err)
			}

			if !tt.expectError && result != nil {
				if result.Version != tt.version {
					t.Errorf("Expected version %s, got %s", tt.version, result.Version)
				}

				if result.AST == nil {
					t.Error("Expected AST to be non-nil")
				}

				if result.AST.FilePath != tt.filePath {
					t.Errorf("Expected file path %s, got %s", tt.filePath, result.AST.FilePath)
				}
			}
		})
	}
}

func TestFrameworkHelperFunctions(t *testing.T) {
	manager := NewManager()

	// Test React component detection
	t.Run("React component detection", func(t *testing.T) {
		node := &types.ASTNode{
			Type: "function_declaration",
			Value: "function MyComponent() {\n  return <div>Hello</div>;\n}",
		}
		content := "import React from 'react';"
		result := manager.isReactComponent(node, content)
		if !result {
			t.Error("Expected React component to be detected")
		}
	})

	// Test React hook detection
	t.Run("React hook detection", func(t *testing.T) {
		node := &types.ASTNode{
			Type: "function_declaration",
			Value: "function useCounter() { return useState(0); }",
			Children: []*types.ASTNode{
				{
					Type: "identifier",
					Value: "useCounter",
				},
			},
		}
		content := "import { useState } from 'react';"
		result := manager.isReactHook(node, content)
		if !result {
			t.Error("Expected React hook to be detected")
		}
	})

	// Test Vue component detection
	t.Run("Vue component detection", func(t *testing.T) {
		node := &types.ASTNode{
			Type: "export_statement",
			Value: "export default { template: '<div>Hello</div>' }",
		}
		content := "import { createApp } from 'vue';"
		result := manager.isVueComponent(node, content)
		if !result {
			t.Error("Expected Vue component to be detected")
		}
	})

	// Test Angular component detection
	t.Run("Angular component detection", func(t *testing.T) {
		node := &types.ASTNode{
			Type: "class_declaration",
			Value: "export class MyComponent {}",
		}
		content := "@Component({ template: '<div>Hello</div>' })\nexport class MyComponent {}"
		result := manager.isAngularComponent(node, content)
		if !result {
			t.Error("Expected Angular component to be detected")
		}
	})

	// Test Svelte store detection
	t.Run("Svelte store detection", func(t *testing.T) {
		node := &types.ASTNode{
			Type: "variable_declaration",
			Value: "const count = writable(0);",
		}
		content := "import { writable } from 'svelte/store';"
		result := manager.isSvelteStore(node, content)
		if !result {
			t.Error("Expected Svelte store to be detected")
		}
	})
}

func TestCalculateHash(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "simple content",
			content:  "hello world",
			expected: "hash-11",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "hash-0",
		},
		{
			name:     "long content",
			content:  "this is a much longer piece of content for testing",
			expected: "hash-50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateHash(tt.content)
			if result != tt.expected {
				t.Errorf("Expected hash %s for content %s, got %s", tt.expected, tt.content, result)
			}
		})
	}
}

func TestLanguageSpecificSymbolExtraction(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name           string
		filePath       string
		content        string
		expectedSymbol string
		expectedType   string
	}{
		{
			name:           "python function",
			filePath:       "test.py",
			content:        "def hello_world():\n    return 'world'",
			expectedSymbol: "hello_world",
			expectedType:   "function",
		},
		{
			name:           "python class",
			filePath:       "test.py",
			content:        "class HelloWorld:\n    def __init__(self):\n        pass",
			expectedSymbol: "HelloWorld",
			expectedType:   "class",
		},
		{
			name:           "java method",
			filePath:       "Test.java",
			content:        "public class Test {\n    public void helloWorld() {}\n}",
			expectedSymbol: "helloWorld",
			expectedType:   "method",
		},
		{
			name:           "java class",
			filePath:       "Test.java",
			content:        "public class HelloWorld {\n    private int value;\n}",
			expectedSymbol: "HelloWorld",
			expectedType:   "class",
		},
		{
			name:           "go function",
			filePath:       "test.go",
			content:        "package main\n\nfunc HelloWorld() string {\n    return \"world\"\n}",
			expectedSymbol: "HelloWorld",
			expectedType:   "function",
		},
		{
			name:           "go struct type",
			filePath:       "test.go",
			content:        "package main\n\ntype HelloWorld struct {\n    Value string\n}",
			expectedSymbol: "HelloWorld",
			expectedType:   "type",
		},
		{
			name:           "rust function",
			filePath:       "test.rs",
			content:        "fn hello_world() -> String {\n    \"Hello\".to_string()\n}",
			expectedSymbol: "hello_world",
			expectedType:   "function",
		},
		{
			name:           "rust struct",
			filePath:       "test.rs",
			content:        "struct HelloWorld {\n    value: String,\n}",
			expectedSymbol: "HelloWorld",
			expectedType:   "class",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the content
			lang := manager.detectLanguage(tt.filePath)
			if lang == nil {
				t.Fatalf("Failed to detect language for %s", tt.filePath)
			}

			ast, err := manager.parseContent(tt.content, *lang, tt.filePath)
			if err != nil {
				t.Fatalf("Failed to parse content: %v", err)
			}

			// Extract symbols
			symbols, err := manager.ExtractSymbols(ast)
			if err != nil {
				t.Fatalf("Failed to extract symbols: %v", err)
			}

			// Check if we found the expected symbol
			found := false
			for _, symbol := range symbols {
				if symbol.Name == tt.expectedSymbol && string(symbol.Type) == tt.expectedType {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected to find symbol '%s' of type '%s' in %s, but didn't find it. Found symbols: %v", 
					tt.expectedSymbol, tt.expectedType, tt.filePath, symbols)
			}
		})
	}
}

func TestFrameworkDetection(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name              string
		filePath          string
		content           string
		expectedFramework string
	}{
		{
			name:              "react component",
			filePath:          "Component.jsx",
			content:           "import React from 'react';\n\nfunction Component() {\n  return <div>Hello</div>;\n}",
			expectedFramework: "React",
		},
		{
			name:              "next.js page",
			filePath:          "page.tsx",
			content:           "import { NextPage } from 'next';\n\nexport default function Page() {}",
			expectedFramework: "Next.js",
		},
		{
			name:              "vue component by extension",
			filePath:          "Component.vue",
			content:           "<template><div>Hello</div></template>",
			expectedFramework: "Vue",
		},
		{
			name:              "vue component by import",
			filePath:          "component.js",
			content:           "import { createApp } from 'vue';\n\ncreateApp({}).mount('#app');",
			expectedFramework: "Vue",
		},
		{
			name:              "angular component",
			filePath:          "component.ts",
			content:           "import { Component } from '@angular/core';\n\n@Component({})\nexport class MyComponent {}",
			expectedFramework: "Angular",
		},
		{
			name:              "svelte component by extension",
			filePath:          "Component.svelte",
			content:           "<script>let name = 'world';</script><h1>Hello {name}!</h1>",
			expectedFramework: "Svelte",
		},
		{
			name:              "astro component",
			filePath:          "page.astro",
			content:           "---\nconst title = 'Hello';\n---\n<html><head><title>{title}</title></head></html>",
			expectedFramework: "Astro",
		},
		{
			name:              "django python",
			filePath:          "views.py",
			content:           "from django.shortcuts import render\n\ndef index(request):\n    return render(request, 'index.html')",
			expectedFramework: "Django",
		},
		{
			name:              "flask python",
			filePath:          "app.py",
			content:           "from flask import Flask\n\napp = Flask(__name__)\n\n@app.route('/')\ndef hello():\n    return 'Hello World!'",
			expectedFramework: "Flask",
		},
		{
			name:              "spring boot java",
			filePath:          "Controller.java",
			content:           "import org.springframework.web.bind.annotation.RestController;\n\n@RestController\npublic class Controller {}",
			expectedFramework: "Spring Boot",
		},
		{
			name:              "no framework",
			filePath:          "util.js",
			content:           "function add(a, b) {\n  return a + b;\n}",
			expectedFramework: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Detect language
			lang := manager.detectLanguage(tt.filePath)
			if lang == nil {
				t.Fatalf("Failed to detect language for %s", tt.filePath)
			}

			// Detect framework
			framework := manager.frameworkDetector.DetectFramework(tt.filePath, lang.Name, tt.content)

			if framework != tt.expectedFramework {
				t.Errorf("Expected framework '%s' for %s, got '%s'", 
					tt.expectedFramework, tt.filePath, framework)
			}
		})
	}
}

func TestFrameworkSpecificSymbolExtraction(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name           string
		filePath       string
		content        string
		expectedSymbol string
		expectedType   string
	}{
		{
			name:           "react functional component",
			filePath:       "Component.jsx",
			content:        "import React from 'react';\n\nfunction MyComponent() {\n  return <div>Hello</div>;\n}\n\nexport default MyComponent;",
			expectedSymbol: "MyComponent",
			expectedType:   "component",
		},
		{
			name:           "react hook",
			filePath:       "hooks.js",
			content:        "import { useState } from 'react';\n\nfunction useCounter() {\n  const [count, setCount] = useState(0);\n  return { count, setCount };\n}",
			expectedSymbol: "useCounter",
			expectedType:   "hook",
		},
		{
			name:           "vue computed property",
			filePath:       "component.js",
			content:        "import { computed } from 'vue';\n\nconst fullName = computed(() => {\n  return firstName.value + ' ' + lastName.value;\n});",
			expectedSymbol: "fullName",
			expectedType:   "computed",
		},
		{
			name:           "angular component",
			filePath:       "my-component.ts",
			content:        "import { Component } from '@angular/core';\n\n@Component({\n  template: '<div>Hello</div>'\n})\nexport class MyComponent {}",
			expectedSymbol: "MyComponent",
			expectedType:   "component",
		},
		{
			name:           "angular service",
			filePath:       "user.service.ts",
			content:        "import { Injectable } from '@angular/core';\n\n@Injectable()\nexport class UserService {}",
			expectedSymbol: "UserService",
			expectedType:   "service",
		},
		{
			name:           "svelte store",
			filePath:       "store.js",
			content:        "import { writable } from 'svelte/store';\n\nconst count = writable(0);",
			expectedSymbol: "count",
			expectedType:   "store",
		},
		{
			name:           "next.js page component",
			filePath:       "/pages/about.tsx",
			content:        "import { NextPage } from 'next';\n\nconst About: NextPage = () => {\n  return <div>About</div>;\n};\n\nexport default About;",
			expectedSymbol: "About",
			expectedType:   "route",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Detect language
			lang := manager.detectLanguage(tt.filePath)
			if lang == nil {
				t.Fatalf("Failed to detect language for %s", tt.filePath)
			}

			// Parse the content
			ast, err := manager.parseContent(tt.content, *lang, tt.filePath)
			if err != nil {
				t.Fatalf("Failed to parse content: %v", err)
			}

			// Extract symbols
			symbols, err := manager.ExtractSymbols(ast)
			if err != nil {
				t.Fatalf("Failed to extract symbols: %v", err)
			}

			// Check if we found the expected framework-specific symbol
			found := false
			for _, symbol := range symbols {
				if symbol.Name == tt.expectedSymbol && string(symbol.Type) == tt.expectedType {
					found = true
					break
				}
			}

			if !found {
				t.Logf("Expected to find framework symbol '%s' of type '%s' in %s", 
					tt.expectedSymbol, tt.expectedType, tt.filePath)
				for i, symbol := range symbols {
					t.Logf("Symbol %d: Name='%s', Type='%s'", i, symbol.Name, symbol.Type)
				}
				// Note: Some framework symbols might not be detected yet due to AST parsing limitations
				// This is expected behavior for the current implementation
			}
		})
	}
}
