package parser

import (
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDartPartFilesDetection tests part file directive parsing and detection
func TestDartPartFilesDetection(t *testing.T) {
	manager := NewManager()
	
	t.Run("part directive", func(t *testing.T) {
		content := `// Main library file
library my_library;

part 'models.dart';
part 'services.dart';
part 'widgets/custom_widget.dart';

class MainClass {
  void mainMethod() {}
}`

		ast, err := manager.parseDartContent(content, "lib.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find part directive symbols
		var partSymbols []*types.Symbol
		var mainClass *types.Symbol
		
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeDirective {
				partSymbols = append(partSymbols, symbol)
			}
			if symbol.Name == "MainClass" {
				mainClass = symbol
			}
		}
		
		require.Len(t, partSymbols, 3, "Should find 3 part directives")
		require.NotNil(t, mainClass, "Should find MainClass")
		
		// Check part directive names
		partFiles := make(map[string]bool)
		for _, symbol := range partSymbols {
			partFiles[symbol.Name] = true
		}
		
		assert.True(t, partFiles["models.dart"], "Should find models.dart part")
		assert.True(t, partFiles["services.dart"], "Should find services.dart part")
		assert.True(t, partFiles["widgets/custom_widget.dart"], "Should find custom_widget.dart part")
		
		t.Logf("Found %d part directives and main class", len(partSymbols))
	})
	
	t.Run("part of directive with file path", func(t *testing.T) {
		content := `// Part file with file path reference
part of 'main.dart';

class PartModel {
  final String name;
  final int id;
  
  PartModel(this.name, this.id);
  
  Map<String, dynamic> toJson() {
    return {'name': name, 'id': id};
  }
}`

		ast, err := manager.parseDartContent(content, "models.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find part of directive and class
		var partOfSymbol *types.Symbol
		var modelClass *types.Symbol
		
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeDirective {
				partOfSymbol = symbol
			}
			if symbol.Name == "PartModel" {
				modelClass = symbol
			}
		}
		
		require.NotNil(t, partOfSymbol, "Should find part of directive")
		require.NotNil(t, modelClass, "Should find PartModel class")
		
		assert.Equal(t, "main.dart", partOfSymbol.Name, "Part of should reference main.dart")
		assert.Equal(t, types.SymbolTypeClass, modelClass.Type, "PartModel should be a class")
		
		t.Logf("Found part of directive: %s and model class: %s", partOfSymbol.Name, modelClass.Name)
	})
	
	t.Run("part of directive with library name", func(t *testing.T) {
		content := `// Part file with library name reference
part of my_library;

extension StringExtensions on String {
  String get capitalized {
    if (isEmpty) return this;
    return '${this[0].toUpperCase()}${substring(1)}';
  }
  
  bool get isValidEmail {
    return RegExp(r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$').hasMatch(this);
  }
}`

		ast, err := manager.parseDartContent(content, "string_extensions.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find part of directive and extension
		var partOfSymbol *types.Symbol
		var extension *types.Symbol
		
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeDirective {
				partOfSymbol = symbol
			}
			if symbol.Type == types.SymbolTypeExtension {
				extension = symbol
			}
		}
		
		require.NotNil(t, partOfSymbol, "Should find part of directive")
		require.NotNil(t, extension, "Should find StringExtensions extension")
		
		assert.Equal(t, "my_library", partOfSymbol.Name, "Part of should reference my_library")
		assert.Equal(t, "StringExtensions", extension.Name, "Should find StringExtensions")
		
		t.Logf("Found part of library: %s and extension: %s", partOfSymbol.Name, extension.Name)
	})
	
	t.Run("complex compilation unit", func(t *testing.T) {
		// Main library file
		mainContent := `library app_models;

part 'user.dart';
part 'product.dart'; 
part 'validators.dart';

abstract class BaseModel {
  Map<String, dynamic> toJson();
  
  void validate() {
    // Base validation logic
  }
}`

		// Test parsing main file
		ast, err := manager.parseDartContent(mainContent, "models.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find multiple part directives and base class
		partDirectives := 0
		var baseClass *types.Symbol
		
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeDirective {
				partDirectives++
			}
			if symbol.Name == "BaseModel" && symbol.Type == types.SymbolTypeClass {
				baseClass = symbol
			}
		}
		
		assert.Equal(t, 3, partDirectives, "Should find 3 part directives")
		require.NotNil(t, baseClass, "Should find BaseModel class")
		
		t.Logf("Complex compilation unit: %d part files, base class: %s", 
			partDirectives, baseClass.Name)
	})
	
	t.Run("invalid part directives", func(t *testing.T) {
		content := `// Test malformed part directives - should not crash
part 'incomplete
part of
part of '';
part '';

class ValidClass {
  void method() {}
}`

		// Should not crash on malformed directives
		ast, err := manager.parseDartContent(content, "malformed.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should still find the valid class
		var validClass *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "ValidClass" {
				validClass = symbol
				break
			}
		}
		
		require.NotNil(t, validClass, "Should find ValidClass despite malformed directives")
		assert.Equal(t, types.SymbolTypeClass, validClass.Type)
		
		t.Logf("Handled malformed directives gracefully, found class: %s", validClass.Name)
	})
}

// TestDartCompilationUnitIntegration tests complete compilation unit handling
func TestDartCompilationUnitIntegration(t *testing.T) {
	manager := NewManager()
	
	t.Run("full compilation unit workflow", func(t *testing.T) {
		// Simulate a complete Dart compilation unit
		mainFile := `library user_management;

import 'package:flutter/material.dart';

part 'user_model.dart';
part 'user_service.dart';

class UserManager {
  final UserService _service = UserService();
  
  Future<List<User>> getUsers() async {
    return await _service.fetchUsers();
  }
}`

		partFile1 := `part of user_management;

class User {
  final String id;
  final String name;
  final String email;
  
  User({required this.id, required this.name, required this.email});
  
  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name, 
    'email': email,
  };
}`

		partFile2 := `part of user_management;

class UserService {
  final String baseUrl = 'https://api.example.com';
  
  Future<List<User>> fetchUsers() async {
    // Implementation would fetch from API
    return [];
  }
  
  Future<User> createUser(User user) async {
    // Implementation would create via API
    return user;
  }
}`

		// Test main file
		mainAST, err := manager.parseDartContent(mainFile, "user_manager.dart")
		require.NoError(t, err)
		
		mainSymbols, err := manager.ExtractSymbols(mainAST)
		require.NoError(t, err)
		
		// Test part file 1  
		part1AST, err := manager.parseDartContent(partFile1, "user_model.dart")
		require.NoError(t, err)
		
		part1Symbols, err := manager.ExtractSymbols(part1AST)
		require.NoError(t, err)
		
		// Test part file 2
		part2AST, err := manager.parseDartContent(partFile2, "user_service.dart")
		require.NoError(t, err)
		
		part2Symbols, err := manager.ExtractSymbols(part2AST)
		require.NoError(t, err)
		
		// Validate symbol distribution
		mainClasses := countSymbolsByType(mainSymbols, types.SymbolTypeClass)
		part1Classes := countSymbolsByType(part1Symbols, types.SymbolTypeClass)
		part2Classes := countSymbolsByType(part2Symbols, types.SymbolTypeClass)
		
		partDirectives := countSymbolsByType(mainSymbols, types.SymbolTypeDirective)
		partOfDirectives := countSymbolsByType(part1Symbols, types.SymbolTypeDirective) +
							countSymbolsByType(part2Symbols, types.SymbolTypeDirective)
		
		assert.Equal(t, 1, mainClasses, "Main file should have 1 class")
		assert.Equal(t, 1, part1Classes, "Part file 1 should have 1 class")  
		assert.Equal(t, 1, part2Classes, "Part file 2 should have 1 class")
		assert.Equal(t, 2, partDirectives, "Main file should have 2 part directives")
		assert.Equal(t, 2, partOfDirectives, "Part files should have 2 part of directives")
		
		t.Logf("Compilation unit validation: %d main classes, %d part classes, %d directives", 
			mainClasses, part1Classes+part2Classes, partDirectives+partOfDirectives)
	})
}

// Helper function to count symbols by type
func countSymbolsByType(symbols []*types.Symbol, symbolType types.SymbolType) int {
	count := 0
	for _, symbol := range symbols {
		if symbol.Type == symbolType {
			count++
		}
	}
	return count
}