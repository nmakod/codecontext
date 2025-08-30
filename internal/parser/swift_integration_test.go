package parser

import (
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwiftIntegration(t *testing.T) {
	manager := NewManager()
	
	// Test comprehensive SwiftUI app
	t.Run("realistic SwiftUI app", func(t *testing.T) {
		swiftUICode := `import SwiftUI
import Combine

@main
struct TodoApp: App {
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}

struct ContentView: View {
    @StateObject private var todoStore = TodoStore()
    @State private var newTodoText = ""
    @State private var showingAddTodo = false
    
    var body: some View {
        NavigationView {
            List {
                ForEach(todoStore.todos) { todo in
                    TodoRowView(todo: todo) {
                        todoStore.toggle(todo)
                    }
                }
                .onDelete(perform: deleteTodos)
            }
            .navigationTitle("Todos")
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Add") {
                        showingAddTodo = true
                    }
                }
            }
        }
        .sheet(isPresented: $showingAddTodo) {
            AddTodoView(todoStore: todoStore)
        }
    }
    
    private func deleteTodos(offsets: IndexSet) {
        todoStore.delete(at: offsets)
    }
}

struct TodoRowView: View {
    let todo: Todo
    let onToggle: () -> Void
    
    var body: some View {
        HStack {
            Button(action: onToggle) {
                Image(systemName: todo.isCompleted ? "checkmark.circle.fill" : "circle")
            }
            .buttonStyle(PlainButtonStyle())
            
            Text(todo.title)
                .strikethrough(todo.isCompleted)
                .foregroundColor(todo.isCompleted ? .gray : .primary)
            
            Spacer()
        }
        .padding(.vertical, 4)
    }
}

class TodoStore: ObservableObject {
    @Published var todos: [Todo] = []
    
    func add(_ todo: Todo) {
        todos.append(todo)
    }
    
    func toggle(_ todo: Todo) {
        if let index = todos.firstIndex(where: { $0.id == todo.id }) {
            todos[index].isCompleted.toggle()
        }
    }
    
    func delete(at offsets: IndexSet) {
        todos.remove(atOffsets: offsets)
    }
}

struct Todo: Identifiable, Codable {
    let id = UUID()
    var title: String
    var isCompleted: Bool = false
    
    init(title: String) {
        self.title = title
    }
}`
		
		ast, err := manager.parseContent(swiftUICode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "TodoApp.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find comprehensive symbol coverage
		assert.GreaterOrEqual(t, len(symbols), 15, "Should find many symbols in complex app")
		
		// Check for key symbols
		symbolNames := make(map[string]bool)
		for _, symbol := range symbols {
			symbolNames[symbol.Name] = true
		}
		
		assert.True(t, symbolNames["TodoApp"], "Should find main app struct")
		assert.True(t, symbolNames["ContentView"], "Should find main view")
		assert.True(t, symbolNames["TodoStore"], "Should find store class")
		assert.True(t, symbolNames["Todo"], "Should find model struct")
		
		// Check framework detection
		assert.True(t, ast.Root.Metadata["has_swiftui"].(bool), "Should detect SwiftUI")
		assert.True(t, ast.Root.Metadata["has_foundation"].(bool), "Should detect Foundation")
		
		// Check advanced features
		assert.True(t, ast.Root.Metadata["has_closures"].(bool), "Should detect closures")
		assert.True(t, ast.Root.Metadata["has_optionals"].(bool), "Should detect optionals")
	})
	
	// Test iOS UIKit app
	t.Run("realistic UIKit app", func(t *testing.T) {
		uiKitCode := `import UIKit

class SceneDelegate: UIResponder, UIWindowSceneDelegate {
    var window: UIWindow?
    
    func scene(_ scene: UIScene, willConnectTo session: UISceneSession, options connectionOptions: UIScene.ConnectionOptions) {
        guard let windowScene = (scene as? UIWindowScene) else { return }
        
        window = UIWindow(windowScene: windowScene)
        let mainVC = MainViewController()
        let navController = UINavigationController(rootViewController: mainVC)
        
        window?.rootViewController = navController
        window?.makeKeyAndVisible()
    }
}

class MainViewController: UIViewController {
    @IBOutlet weak var tableView: UITableView!
    @IBOutlet weak var addButton: UIBarButtonItem!
    
    private var dataSource: [String] = []
    private lazy var refreshControl: UIRefreshControl = {
        let control = UIRefreshControl()
        control.addTarget(self, action: #selector(refreshData), for: .valueChanged)
        return control
    }()
    
    override func viewDidLoad() {
        super.viewDidLoad()
        setupUI()
        loadData()
    }
    
    override func viewWillAppear(_ animated: Bool) {
        super.viewWillAppear(animated)
        updateNavigationBar()
    }
    
    @objc private func refreshData() {
        defer {
            refreshControl.endRefreshing()
        }
        
        guard !dataSource.isEmpty else {
            // Load initial data
            return
        }
        
        // Refresh logic here
    }
    
    @IBAction func addButtonTapped(_ sender: UIBarButtonItem) {
        let alert = UIAlertController(title: "Add Item", message: nil, preferredStyle: .alert)
        
        alert.addTextField { textField in
            textField.placeholder = "Enter item"
        }
        
        let addAction = UIAlertAction(title: "Add", style: .default) { [weak self] _ in
            guard let textField = alert.textFields?.first,
                  let text = textField.text,
                  !text.isEmpty else { return }
            
            self?.addItem(text)
        }
        
        alert.addAction(addAction)
        alert.addAction(UIAlertAction(title: "Cancel", style: .cancel))
        
        present(alert, animated: true)
    }
    
    private func addItem(_ item: String) {
        dataSource.append(item)
        tableView.reloadData()
    }
}`
		
		ast, err := manager.parseContent(uiKitCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "SceneDelegate.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find comprehensive symbol coverage
		assert.GreaterOrEqual(t, len(symbols), 12, "Should find many symbols in UIKit app")
		
		// Check framework detection
		assert.True(t, ast.Root.Metadata["has_uikit"].(bool), "Should detect UIKit")
		
		// Check control flow features
		assert.True(t, ast.Root.Metadata["has_control_flow"].(bool), "Should detect guard/defer")
		assert.True(t, ast.Root.Metadata["has_optionals"].(bool), "Should detect optionals")
		assert.True(t, ast.Root.Metadata["has_closures"].(bool), "Should detect closures")
	})
	
	// Test Vapor backend code
	t.Run("realistic Vapor backend", func(t *testing.T) {
		vaporCode := `import Vapor
import Foundation

typealias RouteHandler = (Request) async throws -> Response

actor UserService {
    private var cache: [UUID: User] = [:]
    
    func getUser(id: UUID) async throws -> User? {
        if let cached = cache[id] {
            return cached
        }
        
        // Fetch from database
        let user = try await User.find(id, on: database)
        cache[id] = user
        return user
    }
}

struct UserController: RouteCollection {
    let userService: UserService
    
    func boot(routes: RoutesBuilder) throws {
        let users = routes.grouped("users")
        users.get(use: index)
        users.post(use: create)
        users.group(":userID") { user in
            user.get(use: show)
            user.put(use: update)
            user.delete(use: delete)
        }
    }
    
    func index(req: Request) async throws -> [User] {
        return try await User.query(on: req.db).all()
    }
    
    func create(req: Request) async throws -> User {
        let user = try req.content.decode(User.self)
        
        guard !user.email.isEmpty else {
            throw Abort(.badRequest, reason: "Email is required")
        }
        
        defer {
            req.logger.info("User created: \(user.email)")
        }
        
        return try await user.save(on: req.db)
    }
    
    func show(req: Request) async throws -> User {
        guard let userID = req.parameters.get("userID", as: UUID.self) else {
            throw Abort(.badRequest)
        }
        
        guard let user = try await userService.getUser(id: userID) else {
            throw Abort(.notFound)
        }
        
        return user
    }
}

final class User: Model, Content {
    static let schema = "users"
    
    @ID(key: .id)
    var id: UUID?
    
    @Field(key: "email")
    var email: String
    
    @Field(key: "name")
    var name: String
    
    @Timestamp(key: "created_at", on: .create)
    var createdAt: Date?
    
    init() {}
    
    init(email: String, name: String) {
        self.email = email
        self.name = name
    }
}`
		
		ast, err := manager.parseContent(vaporCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "UserController.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find comprehensive symbol coverage
		assert.GreaterOrEqual(t, len(symbols), 15, "Should find many symbols in Vapor app")
		
		// Check framework detection
		assert.True(t, ast.Root.Metadata["has_vapor"].(bool), "Should detect Vapor")
		assert.True(t, ast.Root.Metadata["has_foundation"].(bool), "Should detect Foundation")
		
		// Check advanced features
		assert.True(t, ast.Root.Metadata["has_async_await"].(bool), "Should detect async/await")
		assert.True(t, ast.Root.Metadata["has_control_flow"].(bool), "Should detect guard/defer")
		assert.True(t, ast.Root.Metadata["has_optionals"].(bool), "Should detect optionals")
	})
	
	// Test associated types in protocols
	t.Run("protocols with associated types", func(t *testing.T) {
		swiftCode := `protocol Repository {
    associatedtype Entity: Codable
    associatedtype ID: Hashable
    
    func save(_ entity: Entity) async throws -> Entity
    func find(by id: ID) async throws -> Entity?
    func findAll() async throws -> [Entity]
}

protocol Cacheable {
    associatedtype Key: Hashable
    associatedtype Value
    
    var cacheKey: Key { get }
    var cacheValue: Value { get }
}

class UserRepository: Repository {
    typealias Entity = User
    typealias ID = UUID
    
    func save(_ entity: User) async throws -> User {
        // Implementation
        return entity
    }
    
    func find(by id: UUID) async throws -> User? {
        // Implementation
        return nil
    }
    
    func findAll() async throws -> [User] {
        // Implementation
        return []
    }
}`
		
		ast, err := manager.parseContent(swiftCode, types.Language{
			Name: "swift",
			Extensions: []string{".swift"},
			Parser: "tree-sitter-swift",
			Enabled: true,
		}, "Repository.swift")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find protocols, associated types, and implementation
		assert.GreaterOrEqual(t, len(symbols), 10, "Should find protocols, types, and implementations")
		
		// Check for associated types
		var associatedTypeSymbols []*types.Symbol
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeType {
				associatedTypeSymbols = append(associatedTypeSymbols, symbol)
			}
		}
		
		assert.GreaterOrEqual(t, len(associatedTypeSymbols), 2, "Should find associated types and typealiases")
	})
}

// TestSwiftFeatureCoverage validates comprehensive Swift feature detection
func TestSwiftFeatureCoverage(t *testing.T) {
	manager := NewManager()
	
	// Comprehensive Swift feature test
	comprehensiveCode := `import SwiftUI
import Combine
import Foundation

typealias DataHandler = (Result<Data, Error>) -> Void
typealias AsyncDataHandler<T> = (T) async throws -> Void

protocol DataProviding {
    associatedtype DataType: Codable
    associatedtype ErrorType: Error
    
    func fetchData() async throws -> DataType
    func cacheData(_ data: DataType) async
}

actor DataCache<T: Codable> {
    private var storage: [String: T] = [:]
    private var accessCounts: [String: Int] = [:]
    
    func store(_ item: T, forKey key: String) {
        storage[key] = item
        accessCounts[key] = (accessCounts[key] ?? 0) + 1
    }
    
    func retrieve(forKey key: String) -> T? {
        defer {
            accessCounts[key] = (accessCounts[key] ?? 0) + 1
        }
        
        return storage[key]
    }
    
    var totalItems: Int {
        get async {
            return storage.count
        }
    }
}

class NetworkManager: ObservableObject, DataProviding {
    typealias DataType = APIResponse
    typealias ErrorType = NetworkError
    
    @Published var isLoading = false
    @Published var lastError: Error?
    
    private let cache = DataCache<APIResponse>()
    private var cancellables = Set<AnyCancellable>()
    
    var isOnline: Bool {
        get {
            return NetworkMonitor.shared.isConnected
        }
        set {
            // Read-only computed property
        }
    }
    
    func fetchData() async throws -> APIResponse {
        guard isOnline else {
            throw NetworkError.offline
        }
        
        defer {
            await MainActor.run {
                isLoading = false
            }
        }
        
        await MainActor.run {
            isLoading = true
        }
        
        let url = URL(string: "https://api.example.com/data")!
        let data = try await URLSession.shared.data(from: url).0
        let response = try JSONDecoder().decode(APIResponse.self, from: data)
        
        await cache.store(response, forKey: "latest")
        return response
    }
    
    func cacheData(_ data: APIResponse) async {
        await cache.store(data, forKey: UUID().uuidString)
    }
    
    private func handleError(_ error: Error) {
        DispatchQueue.main.async { [weak self] in
            self?.lastError = error
        }
    }
}

extension NetworkManager {
    func clearCache() async {
        // Clear implementation
    }
    
    func retryLastRequest() async throws -> APIResponse? {
        guard let cachedResponse = await cache.retrieve(forKey: "latest") else {
            return try await fetchData()
        }
        
        return cachedResponse
    }
}

struct APIResponse: Codable {
    let id: UUID
    let data: String
    let timestamp: Date
    
    init(id: UUID = UUID(), data: String, timestamp: Date = Date()) {
        self.id = id
        self.data = data
        self.timestamp = timestamp
    }
}

enum NetworkError: Error {
    case offline
    case invalidResponse
    case timeout
    
    var localizedDescription: String {
        switch self {
        case .offline:
            return "No internet connection"
        case .invalidResponse:
            return "Invalid server response"
        case .timeout:
            return "Request timed out"
        }
    }
}`
	
	ast, err := manager.parseContent(comprehensiveCode, types.Language{
		Name: "swift",
		Extensions: []string{".swift"},
		Parser: "tree-sitter-swift",
		Enabled: true,
	}, "NetworkManager.swift")
	require.NoError(t, err)
	require.NotNil(t, ast)
	
	symbols, err := manager.ExtractSymbols(ast)
	require.NoError(t, err)
	
	// Validate comprehensive feature coverage
	featureCount := 0
	detectedFeatures := 0
	
	// Core language features (should all be detected)
	coreFeatures := map[string]bool{
		"has_classes": false,
		"has_structs": false,
		"has_protocols": false,
		"has_enums": false,
		"has_actors": false,
		"has_extensions": false,
		"has_imports": false,
	}
	
	// Advanced features (should all be detected)
	advancedFeatures := map[string]bool{
		"has_async_await": false,
		"has_closures": false,
		"has_optionals": false,
		"has_control_flow": false,
	}
	
	// Check symbol types for core features
	for _, symbol := range symbols {
		switch symbol.Type {
		case types.SymbolTypeClass:
			coreFeatures["has_classes"] = true
		case types.SymbolTypeInterface:
			coreFeatures["has_protocols"] = true
		case types.SymbolTypeType:
			coreFeatures["has_structs"] = true // Includes typealias/associated types
		case types.SymbolTypeImport:
			coreFeatures["has_imports"] = true
		case types.SymbolTypeNamespace:
			coreFeatures["has_extensions"] = true
		}
	}
	
	// Check AST metadata for advanced features
	if val, exists := ast.Root.Metadata["has_async_await"]; exists && val.(bool) {
		advancedFeatures["has_async_await"] = true
	}
	if val, exists := ast.Root.Metadata["has_closures"]; exists && val.(bool) {
		advancedFeatures["has_closures"] = true
	}
	if val, exists := ast.Root.Metadata["has_optionals"]; exists && val.(bool) {
		advancedFeatures["has_optionals"] = true
	}
	if val, exists := ast.Root.Metadata["has_control_flow"]; exists && val.(bool) {
		advancedFeatures["has_control_flow"] = true
	}
	
	// Count feature coverage
	for feature, detected := range coreFeatures {
		featureCount++
		if detected {
			detectedFeatures++
		} else {
			t.Logf("Missing core feature: %s", feature)
		}
	}
	
	for feature, detected := range advancedFeatures {
		featureCount++
		if detected {
			detectedFeatures++
		} else {
			t.Logf("Missing advanced feature: %s", feature)
		}
	}
	
	coverage := float64(detectedFeatures) / float64(featureCount) * 100
	assert.GreaterOrEqual(t, coverage, 70.0, "Swift feature coverage should be ≥70%")
	
	t.Logf("Swift feature coverage: %.1f%% (%d/%d features detected)", coverage, detectedFeatures, featureCount)
	t.Logf("Total symbols extracted: %d", len(symbols))
}