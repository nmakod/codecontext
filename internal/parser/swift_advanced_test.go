package parser

import (
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwiftAdvancedFeatures(t *testing.T) {
	manager := NewManager()
	
	// Test actor parsing
	t.Run("actors", func(t *testing.T) {
		swiftCode := `import Foundation

actor BankAccount {
    private var balance: Double = 0
    
    func deposit(_ amount: Double) {
        balance += amount
    }
    
    func withdraw(_ amount: Double) -> Bool {
        if balance >= amount {
            balance -= amount
            return true
        }
        return false
    }
    
    func getBalance() -> Double {
        return balance
    }
}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "bank.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find actor and methods
		assert.GreaterOrEqual(t, len(symbols), 4)
		
		var actorSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "BankAccount" {
				actorSymbol = symbol
				break
			}
		}
		
		require.NotNil(t, actorSymbol, "Should find BankAccount actor")
		assert.Equal(t, "BankAccount", actorSymbol.Name)
		assert.Equal(t, types.SymbolTypeClass, actorSymbol.Type) // Actors map to class type
	})
	
	// Test typealias parsing
	t.Run("typealias", func(t *testing.T) {
		swiftCode := `import Foundation

typealias StringDictionary = Dictionary<String, String>
typealias CompletionHandler = (Bool) -> Void
typealias GenericHandler<T> = (T) -> Void

class MyClass {
    var data: StringDictionary = [:]
    var onComplete: CompletionHandler?
}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "types.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find typealias declarations
		var typealiasSymbols []*types.Symbol
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeType {
				typealiasSymbols = append(typealiasSymbols, symbol)
			}
		}
		
		assert.GreaterOrEqual(t, len(typealiasSymbols), 2, "Should find at least 2 typealias declarations")
		
		// Check for specific typealias names
		var foundNames []string
		for _, symbol := range typealiasSymbols {
			foundNames = append(foundNames, symbol.Name)
		}
		assert.Contains(t, foundNames, "StringDictionary")
		assert.Contains(t, foundNames, "CompletionHandler")
	})
	
	// Test computed properties
	t.Run("computed properties", func(t *testing.T) {
		swiftCode := `class Rectangle {
    var width: Double = 0
    var height: Double = 0
    
    var area: Double {
        get {
            return width * height
        }
        set {
            let ratio = width / height
            width = sqrt(newValue * ratio)
            height = sqrt(newValue / ratio)
        }
    }
    
    var perimeter: Double {
        return 2 * (width + height)
    }
}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "rectangle.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find both stored and computed properties
		var propertySymbols []*types.Symbol
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeProperty {
				propertySymbols = append(propertySymbols, symbol)
			}
		}
		
		assert.GreaterOrEqual(t, len(propertySymbols), 4, "Should find all properties")
		
		// Check for specific properties
		var foundProperties []string
		for _, symbol := range propertySymbols {
			foundProperties = append(foundProperties, symbol.Name)
		}
		assert.Contains(t, foundProperties, "width")
		assert.Contains(t, foundProperties, "height") 
		assert.Contains(t, foundProperties, "area")
		assert.Contains(t, foundProperties, "perimeter")
	})
	
	// Test property wrappers
	t.Run("property wrappers", func(t *testing.T) {
		swiftCode := `import SwiftUI

struct ContentView: View {
    @State private var isToggled = false
    @StateObject private var viewModel = ViewModel()
    @ObservedObject var dataManager: DataManager
    @Published var items: [String] = []
    @AppStorage("username") var username: String = ""
    @Environment(\.colorScheme) var colorScheme
    
    var body: some View {
        VStack {
            Toggle("Toggle me", isOn: $isToggled)
            Text("Hello, \(username)")
        }
    }
}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "contentview.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find properties with wrappers
		var propertySymbols []*types.Symbol
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeProperty {
				propertySymbols = append(propertySymbols, symbol)
			}
		}
		
		assert.GreaterOrEqual(t, len(propertySymbols), 5, "Should find all wrapped properties")
		
		// Check for specific wrapped properties
		var foundProperties []string
		for _, symbol := range propertySymbols {
			foundProperties = append(foundProperties, symbol.Name)
		}
		assert.Contains(t, foundProperties, "isToggled")
		assert.Contains(t, foundProperties, "viewModel")
		assert.Contains(t, foundProperties, "dataManager")
		assert.Contains(t, foundProperties, "items")
		assert.Contains(t, foundProperties, "username")
	})
	
	// Test closures
	t.Run("closures", func(t *testing.T) {
		swiftCode := `import Foundation

class DataProcessor {
    var onComplete: ((Bool) -> Void)?
    var transform: @escaping (String) -> String = { $0.uppercased() }
    
    func processData(completion: @escaping (Result<String, Error>) -> Void) {
        let processQueue = DispatchQueue.global()
        
        processQueue.async {
            let result = self.data.map { item in
                return item.uppercased()
            }
            
            DispatchQueue.main.async {
                completion(.success(result.joined()))
            }
        }
    }
    
    func sortedData() -> [String] {
        return data.sorted { $0.count < $1.count }
    }
}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "processor.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// Check closure metadata
		assert.True(t, ast.Root.Metadata["has_closures"].(bool), "Should detect closures")
		assert.Greater(t, ast.Root.Metadata["closure_count"].(int), 0, "Should count closures")
	})
	
	// Test async/await
	t.Run("async await", func(t *testing.T) {
		swiftCode := `import Foundation

actor DataManager {
    private var cache: [String: Data] = [:]
    
    func fetchData(from url: URL) async throws -> Data {
        if let cached = cache[url.absoluteString] {
            return cached
        }
        
        let data = try await URLSession.shared.data(from: url).0
        cache[url.absoluteString] = data
        return data
    }
    
    var isReady: Bool {
        get async {
            return !cache.isEmpty
        }
    }
    
    func processAll() async throws {
        let urls = [URL(string: "https://api.example.com")!]
        
        for url in urls {
            let data = try await fetchData(from: url)
            // Process data
        }
    }
}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "datamanager.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// Check async/await metadata
		assert.True(t, ast.Root.Metadata["has_async_await"].(bool), "Should detect async/await")
		assert.Greater(t, ast.Root.Metadata["async_function_count"].(int), 0, "Should count async functions")
		assert.Greater(t, ast.Root.Metadata["await_call_count"].(int), 0, "Should count await calls")
	})
	
	// Test optionals
	t.Run("optionals", func(t *testing.T) {
		swiftCode := `import Foundation

class UserManager {
    var currentUser: User?
    
    func getUserName() -> String? {
        return currentUser?.name?.uppercased()
    }
    
    func processUser() {
        guard let user = currentUser else { return }
        
        if let name = user.name {
            print("Hello, \(name)")
        }
        
        let displayName = user.displayName ?? "Anonymous"
        let forcedValue = user.id!
        
        user.email?.isEmpty ?? true
    }
}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "usermanager.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// Check optional metadata
		assert.True(t, ast.Root.Metadata["has_optionals"].(bool), "Should detect optionals")
		assert.Greater(t, ast.Root.Metadata["optional_chaining_count"].(int), 0, "Should count optional chaining")
		assert.Greater(t, ast.Root.Metadata["optional_binding_count"].(int), 0, "Should count optional binding")
		assert.Greater(t, ast.Root.Metadata["nil_coalescing_count"].(int), 0, "Should count nil coalescing")
	})
	
	// Test guard/defer
	t.Run("guard and defer", func(t *testing.T) {
		swiftCode := `import Foundation

class FileManager {
    func processFile(path: String) throws {
        guard !path.isEmpty else {
            throw FileError.invalidPath
        }
        
        let file = try openFile(path)
        defer {
            closeFile(file)
        }
        
        guard file.isReadable else {
            throw FileError.notReadable
        }
        
        defer {
            logOperation("File processed: \(path)")
        }
        
        // Process file
    }
}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "filemanager.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// Check control flow metadata
		assert.True(t, ast.Root.Metadata["has_control_flow"].(bool), "Should detect control flow")
		assert.Equal(t, 2, ast.Root.Metadata["guard_statement_count"].(int), "Should count guard statements")
		assert.Equal(t, 2, ast.Root.Metadata["defer_statement_count"].(int), "Should count defer statements")
	})
}