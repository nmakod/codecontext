package parser

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwiftFrameworkDetection(t *testing.T) {
	detector := NewFrameworkDetector("/test/project")
	
	t.Run("SwiftUI detection", func(t *testing.T) {
		swiftUICode := `import SwiftUI

struct ContentView: View {
    var body: some View {
        Text("Hello, World!")
    }
}`
		
		framework := detector.DetectFramework("ContentView.swift", "swift", swiftUICode)
		assert.Equal(t, "SwiftUI", framework, "Should detect SwiftUI framework")
	})
	
	t.Run("UIKit detection", func(t *testing.T) {
		uiKitCode := `import UIKit

class ViewController: UIViewController {
    override func viewDidLoad() {
        super.viewDidLoad()
        view.backgroundColor = .white
    }
}`
		
		framework := detector.DetectFramework("ViewController.swift", "swift", uiKitCode)
		assert.Equal(t, "UIKit", framework, "Should detect UIKit framework")
	})
	
	t.Run("Vapor detection", func(t *testing.T) {
		vaporCode := `import Vapor

struct UserController: RouteCollection {
    func boot(routes: RoutesBuilder) throws {
        routes.get("users", use: index)
    }
    
    func index(req: Request) throws -> String {
        return "Hello, Vapor!"
    }
}`
		
		framework := detector.DetectFramework("UserController.swift", "swift", vaporCode)
		assert.Equal(t, "Vapor", framework, "Should detect Vapor framework")
	})
	
	t.Run("Combine detection", func(t *testing.T) {
		combineCode := `import Combine
import Foundation

class DataManager: ObservableObject {
    @Published var items: [String] = []
    private var cancellables = Set<AnyCancellable>()
    
    func loadData() {
        URLSession.shared.dataTaskPublisher(for: url)
            .sink { completion in
                // Handle completion
            } receiveValue: { data in
                // Handle data
            }
            .store(in: &cancellables)
    }
}`
		
		framework := detector.DetectFramework("DataManager.swift", "swift", combineCode)
		assert.Equal(t, "Combine", framework, "Should detect Combine framework")
	})
	
	t.Run("Foundation only - no framework", func(t *testing.T) {
		foundationCode := `import Foundation

class Calculator {
    func add(_ a: Int, _ b: Int) -> Int {
        return a + b
    }
}`
		
		framework := detector.DetectFramework("Calculator.swift", "swift", foundationCode)
		assert.Equal(t, "", framework, "Should not detect framework for Foundation-only code")
	})
	
	t.Run("multiple frameworks - priority order", func(t *testing.T) {
		multiFrameworkCode := `import SwiftUI
import UIKit
import Combine

struct HybridView: View {
    var body: some View {
        Text("Hybrid")
    }
}`
		
		framework := detector.DetectFramework("HybridView.swift", "swift", multiFrameworkCode)
		// SwiftUI should take priority over UIKit
		assert.Equal(t, "SwiftUI", framework, "Should prioritize SwiftUI over other frameworks")
	})
}

func TestSwiftFileClassification(t *testing.T) {
	manager := NewManager()
	
	t.Run("swift source file", func(t *testing.T) {
		classification, err := manager.ClassifyFile("MyClass.swift")
		require.NoError(t, err)
		require.NotNil(t, classification)
		
		assert.Equal(t, "swift", classification.Language.Name)
		assert.Equal(t, "source", classification.FileType)
		assert.False(t, classification.IsTest)
		assert.False(t, classification.IsGenerated)
	})
	
	t.Run("swift test file", func(t *testing.T) {
		classification, err := manager.ClassifyFile("MyClassTests.swift")
		require.NoError(t, err)
		require.NotNil(t, classification)
		
		assert.Equal(t, "swift", classification.Language.Name)
		assert.Equal(t, "test", classification.FileType)
		assert.True(t, classification.IsTest)
		assert.False(t, classification.IsGenerated)
	})
}