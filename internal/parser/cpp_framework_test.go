package parser

import (
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Phase 3: Framework Detection Tests
func TestCppFrameworkDetection(t *testing.T) {
	manager := NewManager()
	
	// Test Qt framework detection
	t.Run("Qt framework", func(t *testing.T) {
		qtCode := `#include <QApplication>
#include <QMainWindow>
#include <QPushButton>
#include <QObject>

class MainWindow : public QMainWindow {
    Q_OBJECT
    
public:
    explicit MainWindow(QWidget *parent = nullptr);
    
public slots:
    void handleButtonClicked();
    
Q_SIGNALS:
    void buttonPressed();
    
private:
    QPushButton *button;
};`
		
		ast, err := manager.parseContent(qtCode, types.Language{
			Name: "cpp",
			Extensions: []string{".cpp"},
			Parser: "tree-sitter-cpp",
			Enabled: true,
		}, "mainwindow.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// Check Qt framework detection
		require.NotNil(t, ast.Root.Metadata)
		assert.True(t, ast.Root.Metadata["has_qt"].(bool), "Should detect Qt framework")
		assert.True(t, ast.Root.Metadata["has_includes"].(bool), "Should detect includes")
		assert.True(t, ast.Root.Metadata["has_classes"].(bool), "Should detect classes")
	})
	
	// Test STL detection
	t.Run("STL library", func(t *testing.T) {
		stlCode := `#include <vector>
#include <string>
#include <algorithm>
#include <memory>
#include <iostream>

class DataProcessor {
public:
    void processData() {
        std::vector<int> numbers = {1, 2, 3, 4, 5};
        std::string text = "Hello World";
        
        auto result = std::find_if(numbers.begin(), numbers.end(), 
            [](int n) { return n > 3; });
            
        std::unique_ptr<std::string> ptr = std::make_unique<std::string>(text);
        std::cout << *ptr << std::endl;
    }
};`
		
		ast, err := manager.parseContent(stlCode, types.Language{
			Name: "cpp",
			Extensions: []string{".cpp"},
			Parser: "tree-sitter-cpp",
			Enabled: true,
		}, "processor.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// Check STL detection
		require.NotNil(t, ast.Root.Metadata)
		assert.True(t, ast.Root.Metadata["has_stl"].(bool), "Should detect STL")
		assert.True(t, ast.Root.Metadata["has_smart_pointers"].(bool), "Should detect smart pointers")
		assert.True(t, ast.Root.Metadata["has_lambdas"].(bool), "Should detect lambdas")
	})
	
	// Test Boost library detection
	t.Run("Boost library", func(t *testing.T) {
		boostCode := `#include <boost/algorithm/string.hpp>
#include <boost/shared_ptr.hpp>
#include <boost/foreach.hpp>

class BoostExample {
public:
    void useBoost() {
        std::string text = "Hello,World,Test";
        std::vector<std::string> tokens;
        
        boost::split(tokens, text, boost::is_any_of(","));
        boost::shared_ptr<int> ptr(new int(42));
        
        BOOST_FOREACH(const std::string& token, tokens) {
            std::cout << token << std::endl;
        }
    }
};`
		
		ast, err := manager.parseContent(boostCode, types.Language{
			Name: "cpp",
			Extensions: []string{".cpp"},
			Parser: "tree-sitter-cpp",
			Enabled: true,
		}, "boost_example.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// Check Boost detection
		require.NotNil(t, ast.Root.Metadata)
		assert.True(t, ast.Root.Metadata["has_boost"].(bool), "Should detect Boost")
	})
	
	// Test Unreal Engine detection
	t.Run("Unreal Engine", func(t *testing.T) {
		unrealCode := `#include "CoreMinimal.h"
#include "GameFramework/Actor.h"
#include "MyActor.generated.h"

UCLASS()
class MYGAME_API AMyActor : public AActor {
    GENERATED_BODY()
    
public:    
    AMyActor();

protected:
    virtual void BeginPlay() override;

    UPROPERTY(VisibleAnywhere, BlueprintReadOnly, Category = Camera)
    class UCameraComponent* CameraComponent;
    
    UFUNCTION(BlueprintCallable, Category = "Gameplay")
    void DoSomething();

public:    
    virtual void Tick(float DeltaTime) override;
};`
		
		ast, err := manager.parseContent(unrealCode, types.Language{
			Name: "cpp",
			Extensions: []string{".cpp"},
			Parser: "tree-sitter-cpp",
			Enabled: true,
		}, "MyActor.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// Check Unreal Engine detection
		require.NotNil(t, ast.Root.Metadata)
		assert.True(t, ast.Root.Metadata["has_unreal"].(bool), "Should detect Unreal Engine")
	})
}

// Phase 3: P2 Features and Framework Coverage Test
func TestCppP2AndFrameworkCoverage(t *testing.T) {
	manager := NewManager()
	
	// Test P2 (modern C++) features
	t.Run("P2 features", func(t *testing.T) {
		cpp20Code := `#include <concepts>
#include <coroutine>

template<typename T>
concept Numeric = std::integral<T> || std::floating_point<T>;

template<Numeric T>
class Calculator {
public:
    // C++17: structured bindings
    auto getDimensions() -> std::pair<int, int> {
        auto [width, height] = std::make_pair(100, 200);
        return {width, height};
    }
    
    // C++17: if constexpr
    template<typename U>
    void process(U value) {
        if constexpr (std::is_integral_v<U>) {
            processInteger(value);
        } else {
            processFloat(value);
        }
    }
    
    // C++20: coroutines
    std::generator<int> generateNumbers() {
        for (int i = 0; i < 10; ++i) {
            co_yield i;
        }
    }
    
private:
    void processInteger(auto value) { /* impl */ }
    void processFloat(auto value) { /* impl */ }
};`
		
		ast, err := manager.parseContent(cpp20Code, types.Language{
			Name: "cpp",
			Extensions: []string{".cpp"},
			Parser: "tree-sitter-cpp",
			Enabled: true,
		}, "modern_cpp20.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// P2 features to detect
		p2Features := map[string]bool{
			"has_concepts":           false,
			"has_structured_binding": false,
			"has_if_constexpr":      false,
			"has_coroutines":        false,
			"has_modules":           false,
		}
		
		// Check P2 feature detection against AST metadata
		require.NotNil(t, ast.Root.Metadata)
		for feature := range p2Features {
			if val, exists := ast.Root.Metadata[feature]; exists && val.(bool) {
				p2Features[feature] = true
			}
		}
		
		// Calculate P2 coverage
		detected := 0
		total := len(p2Features)
		for feature, isDetected := range p2Features {
			if isDetected {
				detected++
			} else {
				t.Logf("Missing P2 feature: %s", feature)
			}
		}
		
		coverage := float64(detected) / float64(total) * 100
		t.Logf("P2 C++ Feature Coverage: %.1f%% (%d/%d)", coverage, detected, total)
		
		// Phase 3 target: 70% P2 feature coverage
		assert.GreaterOrEqual(t, coverage, 70.0, "Should achieve 70%+ P2 feature coverage")
	})
	
	// Test framework coverage
	t.Run("framework coverage", func(t *testing.T) {
		frameworkCode := `#include <QApplication>
#include <boost/algorithm/string.hpp>
#include <opencv2/opencv.hpp>
#include <vector>
#include <memory>
#include "UnrealEngine.h"

// Qt usage
QWidget* widget = new QWidget();

// Boost usage  
boost::shared_ptr<int> boostPtr;

// OpenCV usage
cv::Mat image = cv::imread("image.jpg");

// STL usage
std::vector<int> numbers;
std::unique_ptr<std::string> text;

// Unreal usage (if present)
UCLASS()
class MyClass {};`
		
		ast, err := manager.parseContent(frameworkCode, types.Language{
			Name: "cpp",
			Extensions: []string{".cpp"},
			Parser: "tree-sitter-cpp",
			Enabled: true,
		}, "frameworks.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// Framework features to detect
		frameworkFeatures := map[string]bool{
			"has_qt":     false,
			"has_boost":  false,
			"has_opencv": false,
			"has_unreal": false,
			"has_stl":    false,
		}
		
		// Check framework detection against AST metadata
		require.NotNil(t, ast.Root.Metadata)
		for feature := range frameworkFeatures {
			if val, exists := ast.Root.Metadata[feature]; exists && val.(bool) {
				frameworkFeatures[feature] = true
			}
		}
		
		// Calculate framework coverage
		detected := 0
		total := len(frameworkFeatures)
		for feature, isDetected := range frameworkFeatures {
			if isDetected {
				detected++
			} else {
				t.Logf("Missing framework: %s", feature)
			}
		}
		
		coverage := float64(detected) / float64(total) * 100
		t.Logf("Framework Detection Coverage: %.1f%% (%d/%d)", coverage, detected, total)
		
		// Phase 3 target: 80% framework detection coverage
		assert.GreaterOrEqual(t, coverage, 80.0, "Should achieve 80%+ framework detection coverage")
	})
}