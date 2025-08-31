package parser

import (
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Phase 2: TDD Red - Template and Modern C++ Features
func TestCppTemplates(t *testing.T) {
	manager := NewManager()
	
	// Test template class
	t.Run("template class", func(t *testing.T) {
		cppCode := `template<typename T, int N = 10>
class Container {
public:
    T data[N];
    
    template<typename U>
    void store(const U& item) {
        // implementation
    }
    
    auto get(int index) -> T& {
        return data[index];
    }
};`
		
		ast, err := manager.parseContent(cppCode, types.Language{
			Name: "cpp",
			Extensions: []string{".cpp"},
			Parser: "tree-sitter-cpp",
			Enabled: true,
		}, "container.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		t.Logf("Found %d symbols", len(symbols))
		
		// Should find template class and methods
		assert.GreaterOrEqual(t, len(symbols), 3)
		
		// Check template feature detection
		require.NotNil(t, ast.Root.Metadata)
		assert.True(t, ast.Root.Metadata["has_templates"].(bool), "Should detect templates")
		assert.True(t, ast.Root.Metadata["has_auto_keyword"].(bool), "Should detect auto keyword")
	})
	
	// Test modern C++ features
	t.Run("modern cpp features", func(t *testing.T) {
		cppCode := `#include <memory>
#include <vector>
#include <algorithm>

class ModernClass {
public:
    // C++11: auto keyword
    auto getValue() const -> int { return value_; }
    
    // C++11: lambda expressions
    void processItems() {
        auto lambda = [this](const auto& item) {
            return item * 2;
        };
        
        std::for_each(items_.begin(), items_.end(), lambda);
    }
    
    // C++11: range-based for loop
    void printAll() {
        for (const auto& item : items_) {
            std::cout << item << std::endl;
        }
    }
    
    // C++11: smart pointers
    std::unique_ptr<int> createValue() {
        return std::make_unique<int>(42);
    }
    
private:
    int value_ = 0;
    std::vector<int> items_;
};`
		
		ast, err := manager.parseContent(cppCode, types.Language{
			Name: "cpp",
			Extensions: []string{".cpp"},
			Parser: "tree-sitter-cpp",
			Enabled: true,
		}, "modern.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		_, err = manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Check modern C++ feature detection
		require.NotNil(t, ast.Root.Metadata)
		assert.True(t, ast.Root.Metadata["has_auto_keyword"].(bool), "Should detect auto")
		assert.True(t, ast.Root.Metadata["has_lambdas"].(bool), "Should detect lambdas")
		assert.True(t, ast.Root.Metadata["has_range_for"].(bool), "Should detect range-based for")
		assert.True(t, ast.Root.Metadata["has_smart_pointers"].(bool), "Should detect smart pointers")
	})
}

// Phase 2: P1 Feature Coverage Test
func TestCppP1FeatureCoverage(t *testing.T) {
	manager := NewManager()
	
	// Comprehensive P1 features code sample
	cppCode := `#include <memory>
#include <vector>
#include <functional>

template<typename T>
class Matrix {
private:
    std::vector<std::vector<T>> data_;
    
public:
    // Constructor
    Matrix(size_t rows, size_t cols) : data_(rows, std::vector<T>(cols)) {}
    
    // Destructor  
    ~Matrix() = default;
    
    // Auto return type deduction
    auto size() const -> std::pair<size_t, size_t> {
        return {data_.size(), data_.empty() ? 0 : data_[0].size()};
    }
    
    // Operator overloading
    T& operator()(size_t row, size_t col) {
        return data_[row][col];
    }
    
    // Lambda usage
    void transform(std::function<T(T)> func) {
        for (auto& row : data_) {
            for (auto& element : row) {
                element = func(element);
            }
        }
    }
    
    // Constexpr function
    constexpr static T zero() {
        return T{};
    }
};

// Smart pointer usage
std::unique_ptr<Matrix<double>> createMatrix() {
    return std::make_unique<Matrix<double>>(10, 10);
}`
	
	ast, err := manager.parseContent(cppCode, types.Language{
		Name: "cpp",
		Extensions: []string{".cpp"},
		Parser: "tree-sitter-cpp",
		Enabled: true,
	}, "matrix.cpp")
	require.NoError(t, err)
	require.NotNil(t, ast)
	
	// P1 features to detect
	p1Features := map[string]bool{
		"has_templates":         false,
		"has_auto_keyword":      false,
		"has_lambdas":          false,
		"has_range_for":        false,
		"has_smart_pointers":   false,
		"has_constexpr":        false,
		"has_operator_overload": false,
	}
	
	// Check feature detection against AST metadata
	require.NotNil(t, ast.Root.Metadata)
	for feature := range p1Features {
		if val, exists := ast.Root.Metadata[feature]; exists && val.(bool) {
			p1Features[feature] = true
		}
	}
	
	// Calculate P1 coverage
	detected := 0
	total := len(p1Features)
	for feature, isDetected := range p1Features {
		if isDetected {
			detected++
		} else {
			t.Logf("Missing P1 feature: %s", feature)
		}
	}
	
	coverage := float64(detected) / float64(total) * 100
	t.Logf("P1 C++ Feature Coverage: %.1f%% (%d/%d)", coverage, detected, total)
	
	// Phase 2 target: 85% P1 feature coverage
	assert.GreaterOrEqual(t, coverage, 85.0, "Should achieve 85%+ P1 feature coverage")
}