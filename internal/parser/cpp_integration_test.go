package parser

import (
	"fmt"
	"testing"
	"time"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Phase 4: Integration Testing and Final Validation
func TestCppIntegration(t *testing.T) {
	manager := NewManager()
	
	// Test realistic C++ project file
	t.Run("realistic cpp project", func(t *testing.T) {
		realisticCode := `// GameEngine.h
#pragma once

#include <memory>
#include <vector>
#include <string>
#include <unordered_map>
#include <functional>

namespace GameEngine {
    // Forward declarations
    class GameObject;
    class Component;
    
    // Modern C++ template with concepts (C++20)
    template<typename T>
    concept Renderable = requires(T t) {
        t.render();
        t.getPosition();
    };
    
    // Game Object System
    class GameObject {
    public:
        GameObject() = default;
        virtual ~GameObject() = default;
        
        // Template method with auto return type
        template<typename T>
        auto getComponent() -> std::shared_ptr<T> {
            static_assert(std::is_base_of_v<Component, T>);
            // Implementation...
            return nullptr;
        }
        
        // Modern C++ lambda and range-based for
        void updateComponents(float deltaTime) {
            for (auto& component : components_) {
                if (component->isActive()) {
                    component->update(deltaTime);
                }
            }
        }
        
        // Operator overloading
        GameObject& operator+=(std::shared_ptr<Component> component) {
            components_.emplace_back(component);
            return *this;
        }
        
        // Structured bindings usage (C++17)
        auto getBounds() const -> std::pair<float, float> {
            auto [width, height] = calculateBounds();
            return {width, height};
        }
        
    private:
        std::vector<std::shared_ptr<Component>> components_;
        std::unordered_map<std::string, std::any> properties_;
        
        auto calculateBounds() const -> std::pair<float, float> {
            return {100.0f, 200.0f};
        }
    };
    
    // Abstract base component
    class Component {
    public:
        Component() = default;
        virtual ~Component() = default;
        
        virtual void update(float deltaTime) = 0;
        virtual void render() = 0;
        
        bool isActive() const { return active_; }
        void setActive(bool active) { active_ = active; }
        
    private:
        bool active_ = true;
    };
    
    // Specific component implementation
    class RenderComponent : public Component {
    public:
        explicit RenderComponent(const std::string& texturePath) 
            : texturePath_(texturePath) {}
        
        void update(float deltaTime) override {
            // Update logic with constexpr calculations
            constexpr float maxDelta = 0.016f;  // 60 FPS
            if constexpr (true) {  // C++17 if constexpr
                deltaTime = std::min(deltaTime, maxDelta);
            }
            
            // Lambda usage for animation
            auto animate = [this, deltaTime](auto& property) {
                property += deltaTime * animationSpeed_;
            };
        }
        
        void render() override {
            // Rendering implementation
        }
        
    private:
        std::string texturePath_;
        float animationSpeed_ = 1.0f;
    };
    
    // Smart pointer factory pattern
    template<typename T, typename... Args>
    auto createGameObject(Args&&... args) -> std::unique_ptr<T> 
        requires std::is_base_of_v<GameObject, T>
    {
        return std::make_unique<T>(std::forward<Args>(args)...);
    }
}

// Usage example
int main() {
    using namespace GameEngine;
    
    auto gameObject = createGameObject<GameObject>();
    auto renderComponent = std::make_shared<RenderComponent>("texture.png");
    
    *gameObject += renderComponent;
    gameObject->updateComponents(0.016f);
    
    return 0;
}`
		
		ast, err := manager.parseContent(realisticCode, types.Language{
			Name: "cpp",
			Extensions: []string{".cpp"},
			Parser: "tree-sitter-cpp",
			Enabled: true,
		}, "GameEngine.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		t.Logf("Found %d symbols in realistic C++ project", len(symbols))
		assert.GreaterOrEqual(t, len(symbols), 15, "Should find many symbols in complex file")
		
		// Validate comprehensive feature detection
		require.NotNil(t, ast.Root.Metadata)
		
		// Core features
		assert.True(t, ast.Root.Metadata["has_classes"].(bool), "Should detect classes")
		assert.True(t, ast.Root.Metadata["has_namespaces"].(bool), "Should detect namespaces")
		assert.True(t, ast.Root.Metadata["has_templates"].(bool), "Should detect templates")
		assert.True(t, ast.Root.Metadata["has_includes"].(bool), "Should detect includes")
		
		// Modern C++ features
		assert.True(t, ast.Root.Metadata["has_auto_keyword"].(bool), "Should detect auto")
		assert.True(t, ast.Root.Metadata["has_lambdas"].(bool), "Should detect lambdas")
		assert.True(t, ast.Root.Metadata["has_smart_pointers"].(bool), "Should detect smart pointers")
		assert.True(t, ast.Root.Metadata["has_constexpr"].(bool), "Should detect constexpr")
		assert.True(t, ast.Root.Metadata["has_operator_overload"].(bool), "Should detect operator overloading")
		
		// Advanced features
		assert.True(t, ast.Root.Metadata["has_concepts"].(bool), "Should detect concepts")
		assert.True(t, ast.Root.Metadata["has_structured_binding"].(bool), "Should detect structured bindings")
		assert.True(t, ast.Root.Metadata["has_if_constexpr"].(bool), "Should detect if constexpr")
		
		// STL usage
		assert.True(t, ast.Root.Metadata["has_stl"].(bool), "Should detect STL usage")
	})
}

// Phase 4: Performance and Memory Validation
func TestCppPerformance(t *testing.T) {
	manager := NewManager()
	
	t.Run("parsing performance", func(t *testing.T) {
		// Large C++ file simulation
		largeCode := generateLargeCppFile()
		
		start := time.Now()
		ast, err := manager.parseContent(largeCode, types.Language{
			Name: "cpp",
			Extensions: []string{".cpp"},
			Parser: "tree-sitter-cpp",
			Enabled: true,
		}, "large_file.cpp")
		parseTime := time.Since(start)
		
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// Performance target: <50ms for large files (considering CI overhead)
		maxTime := 50 * time.Millisecond
		if testing.Short() {
			maxTime = 10 * time.Millisecond  // Stricter for unit tests
		}
		
		t.Logf("Parse time: %v", parseTime)
		assert.Less(t, parseTime, maxTime, "Parsing should be fast")
		
		// Memory validation - ensure AST is reasonable size
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		t.Logf("Extracted %d symbols", len(symbols))
		assert.GreaterOrEqual(t, len(symbols), 100, "Should extract many symbols from large file")
		assert.Less(t, len(symbols), 10000, "Symbol count should be reasonable")
	})
}

// Phase 4: Final Coverage Validation
func TestCppFinalCoverageValidation(t *testing.T) {
	manager := NewManager()
	
	comprehensiveCode := `// C++ Comprehensive Feature Test File
#include <iostream>
#include <vector>
#include <memory>
#include <algorithm>
#include <concepts>
#include <coroutine>
#include <QApplication>
#include <boost/algorithm/string.hpp>
#include <opencv2/opencv.hpp>
#include "UnrealEngine.h"

// Core Features: Namespace
namespace GameEngine {
    // Core Features: Struct
    struct GameConfig {
        int width = 1920;
        int height = 1080;
        bool fullscreen = true;
    };
    
    // P2 Features: Concepts (C++20)
    template<typename T>
    concept Renderable = requires(T t) {
        t.render();
        t.getPosition();
    };
    
    // Core Features: Base class for inheritance
    class GameObject {
    public:
        // Core Features: Constructor
        GameObject() = default;
        // Core Features: Destructor
        virtual ~GameObject() = default;
        
        // P1 Features: Template with auto return type
        template<typename T>
        auto getComponent() -> std::shared_ptr<T> {
            return std::make_shared<T>();
        }
        
        // P1 Features: Range-based for loop
        void updateComponents(float deltaTime) {
            for (const auto& component : components_) {
                component->update(deltaTime);
            }
        }
        
        // P1 Features: Operator overloading
        GameObject& operator+=(std::shared_ptr<int> component) {
            components_.emplace_back(component);
            return *this;
        }
        
        // P2 Features: Structured bindings usage (C++17)
        auto getBounds() const -> std::pair<float, float> {
            auto [width, height] = calculateBounds();
            return {width, height};
        }
        
        // Core Features: Function/method
        virtual void render() = 0;
        
    private:
        std::vector<std::shared_ptr<int>> components_;
        
        auto calculateBounds() const -> std::pair<float, float> {
            return {100.0f, 200.0f};
        }
    };
    
    // Core Features: Inheritance - derived class
    class RenderableObject : public GameObject {
    public:
        // Core Features: Constructor with inheritance
        explicit RenderableObject(const std::string& name) : name_(name) {}
        
        // P1 Features: Lambda expressions and constexpr
        void processData() {
            constexpr float maxValue = 100.0f;
            
            auto lambda = [maxValue](float value) {
                return std::min(value, maxValue);
            };
            
            // P2 Features: if constexpr (C++17)
            if constexpr (std::is_floating_point_v<float>) {
                position_ = lambda(position_);
            }
        }
        
        // Override virtual function
        void render() override {
            // Framework: STL usage
            std::cout << "Rendering: " << name_ << std::endl;
        }
        
        // P2 Features: Coroutine (C++20)
        std::generator<int> generateFrames() {
            for (int i = 0; i < 60; ++i) {
                co_yield i;
            }
        }
        
    private:
        std::string name_;
        float position_ = 0.0f;
    };
    
    // P1 Features: Smart pointer factory with perfect forwarding
    template<typename T, typename... Args>
    auto createObject(Args&&... args) -> std::unique_ptr<T> {
        return std::make_unique<T>(std::forward<Args>(args)...);
    }
}

// Framework Usage Examples
void frameworkExamples() {
    // Framework: Qt usage
    QApplication app(0, nullptr);
    QWidget* widget = new QWidget();
    
    // Framework: Boost usage
    std::string text = "Hello,World,Test";
    std::vector<std::string> tokens;
    boost::split(tokens, text, boost::is_any_of(","));
    boost::shared_ptr<int> boostPtr = boost::make_shared<int>(42);
    
    // Framework: OpenCV usage
    cv::Mat image = cv::imread("image.jpg");
    cv::Mat processed;
    cv::GaussianBlur(image, processed, cv::Size(15, 15), 0);
    
    // Framework: Unreal Engine usage
    UCLASS()
    class MyActor {};
    UPROPERTY(VisibleAnywhere)
    int32 Health = 100;
    
    // Framework: STL comprehensive usage
    std::vector<int> numbers = {1, 2, 3, 4, 5};
    std::unique_ptr<std::string> smartPtr = std::make_unique<std::string>("test");
    std::shared_ptr<GameEngine::GameObject> gameObj;
}

// P2 Features: Modules (C++20) - import statement
import std.core; // Module import

// Core Features: Global function
int main() {
    using namespace GameEngine;
    
    auto config = GameConfig{};
    auto gameObject = createObject<RenderableObject>("Player");
    gameObject->render();
    
    frameworkExamples();
    
    return 0;
}`
	
	ast, err := manager.parseContent(comprehensiveCode, types.Language{
		Name: "cpp",
		Extensions: []string{".cpp"},
		Parser: "tree-sitter-cpp",
		Enabled: true,
	}, "comprehensive.cpp")
	require.NoError(t, err)
	require.NotNil(t, ast)
	
	// Final coverage calculation for all phases
	allFeatures := map[string]bool{
		// Core features (Phase 1) - 8 features
		"has_classes":      false,
		"has_structs":      false, 
		"has_functions":    false,
		"has_namespaces":   false,
		"has_constructors": false,
		"has_destructors":  false,
		"has_inheritance":  false,
		"has_includes":     false,
		
		// P1 features (Phase 2) - 7 features  
		"has_templates":         false,
		"has_auto_keyword":      false,
		"has_lambdas":          false,
		"has_range_for":        false,
		"has_smart_pointers":   false,
		"has_constexpr":        false,
		"has_operator_overload": false,
		
		// P2 features (Phase 3) - 5 features
		"has_concepts":           false,
		"has_structured_binding": false,
		"has_if_constexpr":      false,
		"has_coroutines":        false,
		"has_modules":           false,
		
		// Framework features (Phase 3) - 5 features
		"has_qt":     false,
		"has_boost":  false,
		"has_opencv": false,
		"has_unreal": false,
		"has_stl":    false,
	}
	
	// Check comprehensive feature detection
	require.NotNil(t, ast.Root.Metadata)
	for feature := range allFeatures {
		if val, exists := ast.Root.Metadata[feature]; exists && val.(bool) {
			allFeatures[feature] = true
		}
	}
	
	// Calculate overall coverage by category
	coreDetected, p1Detected, p2Detected, frameworkDetected := 0, 0, 0, 0
	coreTotal, p1Total, p2Total, frameworkTotal := 8, 7, 5, 5
	
	coreFeatures := []string{"has_classes", "has_structs", "has_functions", "has_namespaces", "has_constructors", "has_destructors", "has_inheritance", "has_includes"}
	p1Features := []string{"has_templates", "has_auto_keyword", "has_lambdas", "has_range_for", "has_smart_pointers", "has_constexpr", "has_operator_overload"}
	p2Features := []string{"has_concepts", "has_structured_binding", "has_if_constexpr", "has_coroutines", "has_modules"}
	frameworkFeatures := []string{"has_qt", "has_boost", "has_opencv", "has_unreal", "has_stl"}
	
	for _, feature := range coreFeatures {
		if allFeatures[feature] { coreDetected++ }
	}
	for _, feature := range p1Features {
		if allFeatures[feature] { p1Detected++ }
	}
	for _, feature := range p2Features {
		if allFeatures[feature] { p2Detected++ }
	}
	for _, feature := range frameworkFeatures {
		if allFeatures[feature] { frameworkDetected++ }
	}
	
	// Calculate coverage percentages
	coreCoverage := float64(coreDetected) / float64(coreTotal) * 100
	p1Coverage := float64(p1Detected) / float64(p1Total) * 100
	p2Coverage := float64(p2Detected) / float64(p2Total) * 100
	frameworkCoverage := float64(frameworkDetected) / float64(frameworkTotal) * 100
	
	// Report final results
	t.Logf("=== C++ LANGUAGE SUPPORT FINAL RESULTS ===")
	t.Logf("Core Features Coverage: %.1f%% (%d/%d)", coreCoverage, coreDetected, coreTotal)
	t.Logf("P1 Features Coverage: %.1f%% (%d/%d)", p1Coverage, p1Detected, p1Total)
	t.Logf("P2 Features Coverage: %.1f%% (%d/%d)", p2Coverage, p2Detected, p2Total)
	t.Logf("Framework Coverage: %.1f%% (%d/%d)", frameworkCoverage, frameworkDetected, frameworkTotal)
	
	// Overall weighted score (Core=40%, P1=30%, P2=20%, Framework=10%)
	overallScore := (coreCoverage*0.4 + p1Coverage*0.3 + p2Coverage*0.2 + frameworkCoverage*0.1)
	t.Logf("Overall Weighted Score: %.1f%%", overallScore)
	
	// Phase 4 validation - all targets met
	assert.GreaterOrEqual(t, coreCoverage, 85.0, "Core features should be ≥85%")
	assert.GreaterOrEqual(t, p1Coverage, 85.0, "P1 features should be ≥85%")  
	assert.GreaterOrEqual(t, p2Coverage, 70.0, "P2 features should be ≥70%")
	assert.GreaterOrEqual(t, frameworkCoverage, 80.0, "Framework detection should be ≥80%")
	assert.GreaterOrEqual(t, overallScore, 80.0, "Overall score should be ≥80%")
}

// Helper function to generate large C++ file for performance testing
func generateLargeCppFile() string {
	baseClass := `#include <vector>
#include <memory>

class TestClass%d {
public:
    TestClass%d() = default;
    ~TestClass%d() = default;
    
    void method%d() {
        auto lambda = [](int x) { return x * 2; };
        std::vector<int> data = {1, 2, 3, 4, 5};
        
        for (const auto& item : data) {
            result_ += lambda(item);
        }
    }
    
    template<typename T>
    auto process(T value) -> T {
        return value * multiplier_;
    }
    
private:
    int result_ = 0;
    double multiplier_ = 1.5;
};

`
	
	result := ""
	for i := 0; i < 50; i++ {  // Generate 50 classes
		class := fmt.Sprintf(baseClass, i, i, i, i)
		result += class
	}
	
	return result
}