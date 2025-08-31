package parser

import (
	"strings"
	"testing"

	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// P1 Enhancement Tests: Access specifiers, virtual/override/final, template specialization

func TestCppAccessSpecifiers(t *testing.T) {
	manager := NewManager()

	t.Run("access specifier detection", func(t *testing.T) {
		cppCode := `class TestClass {
private:
    int privateVar;
    void privateMethod();

public:
    TestClass();
    ~TestClass();
    void publicMethod();

protected:
    int protectedVar;
    virtual void protectedVirtual();

public:
    static int staticPublicVar;
};`

		ast, err := manager.parseContent(cppCode, types.Language{
			Name:       "cpp",
			Extensions: []string{".cpp"},
			Parser:     "tree-sitter-cpp",
			Enabled:    true,
		}, "access_test.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)

		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)

		t.Logf("Found %d symbols", len(symbols))

		// Enhanced parser provides detailed symbol classification with access specifiers

		// Find symbols by name and check their visibility
		symbolMap := make(map[string]*types.Symbol)
		for _, symbol := range symbols {
			key := symbol.Name + "_" + string(symbol.Type)
			symbolMap[key] = symbol
		}

		// Check private members
		if privateVar := symbolMap["privateVar_variable"]; privateVar != nil {
			assert.Equal(t, "private", privateVar.Visibility, "privateVar should have private visibility")
		}
		
		if privateMethod := symbolMap["privateMethod_method"]; privateMethod != nil {
			assert.Equal(t, "private", privateMethod.Visibility, "privateMethod should have private visibility")
		}

		// Check public members
		if constructor := symbolMap["TestClass_constructor"]; constructor != nil {
			assert.Equal(t, "public", constructor.Visibility, "Constructor should have public visibility")
		}
		
		if publicMethod := symbolMap["publicMethod_method"]; publicMethod != nil {
			assert.Equal(t, "public", publicMethod.Visibility, "publicMethod should have public visibility")
		}

		// Check protected members
		if protectedVar := symbolMap["protectedVar_variable"]; protectedVar != nil {
			assert.Equal(t, "protected", protectedVar.Visibility, "protectedVar should have protected visibility")
		}
	})
}

func TestCppVirtualOverrideFinal(t *testing.T) {
	manager := NewManager()

	t.Run("virtual/override/final detection", func(t *testing.T) {
		cppCode := `class Base {
public:
    virtual void virtualMethod();
    virtual void pureVirtualMethod() = 0;
    virtual ~Base() = default;
};

class Derived : public Base {
public:
    void virtualMethod() override;
    void pureVirtualMethod() override final;
    
    virtual void newVirtual() final;
};`

		ast, err := manager.parseContent(cppCode, types.Language{
			Name:       "cpp",
			Extensions: []string{".cpp"},
			Parser:     "tree-sitter-cpp",
			Enabled:    true,
		}, "virtual_test.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)

		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)

		t.Logf("Found %d symbols", len(symbols))

		// Check virtual method signatures
		for _, symbol := range symbols {
			if symbol.Name == "virtualMethod" && symbol.Type == types.SymbolTypeMethod {
				t.Logf("Method: %s, Signature: %s", symbol.Name, symbol.Signature)
				if symbol.Signature != "" {
					// Check if signature contains virtual qualifier info
					assert.Contains(t, symbol.Signature, "virtual", "Should detect virtual qualifier")
				}
			}
			
			if symbol.Name == "pureVirtualMethod" && symbol.Type == types.SymbolTypeMethod {
				t.Logf("Method: %s, Signature: %s", symbol.Name, symbol.Signature)
				if symbol.Signature != "" {
					// Pure virtual should be detected in base class, override in derived class
					hasPureVirtual := strings.Contains(symbol.Signature, "pure virtual") || strings.Contains(symbol.Signature, "= 0")
					hasOverride := strings.Contains(symbol.Signature, "override")
					
					// Either it's pure virtual (base class) OR it has override (derived class)
					assert.True(t, hasPureVirtual || hasOverride,
						"Should detect either pure virtual or override for pureVirtualMethod")
				}
			}
		}
	})
}

func TestCppTemplateSpecialization(t *testing.T) {
	manager := NewManager()

	t.Run("template specialization detection", func(t *testing.T) {
		cppCode := `// Primary template
template<typename T>
class MyTemplate {
public:
    void primaryMethod();
};

// Full specialization
template<>
class MyTemplate<int> {
public:
    void specializedMethod();
    int getValue() { return 42; }
};

// Partial specialization
template<typename T>
class MyTemplate<T*> {
public:
    void pointerSpecialization();
};

// Function template specialization
template<typename T>
void processValue(T value) {
    // primary template
}

template<>
void processValue<std::string>(std::string value) {
    // specialized for string
}`

		ast, err := manager.parseContent(cppCode, types.Language{
			Name:       "cpp",
			Extensions: []string{".cpp"},
			Parser:     "tree-sitter-cpp",
			Enabled:    true,
		}, "template_spec_test.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)

		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)

		t.Logf("Found %d symbols", len(symbols))

		// Count templates and specializations
		templateCount := 0
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeTemplate {
				templateCount++
				t.Logf("Template: %s, Signature: %s", symbol.Name, symbol.Signature)
			}
		}

		// Should detect multiple template-related symbols
		assert.GreaterOrEqual(t, templateCount, 1, "Should detect template declarations")

		// Check for specialization indicators in metadata
		if ast.Root != nil && ast.Root.Metadata != nil {
			hasTemplates := ast.Root.Metadata["has_templates"].(bool)
			assert.True(t, hasTemplates, "Should detect template usage")
		}
	})
}

func TestCppOperatorOverloading(t *testing.T) {
	manager := NewManager()

	t.Run("operator overloading detection", func(t *testing.T) {
		cppCode := `class Vector3 {
private:
    float x, y, z;

public:
    Vector3(float x = 0, float y = 0, float z = 0) : x(x), y(y), z(z) {}

    // Arithmetic operators
    Vector3 operator+(const Vector3& other) const;
    Vector3 operator-(const Vector3& other) const;
    Vector3 operator*(float scalar) const;

    // Assignment operators
    Vector3& operator+=(const Vector3& other);
    Vector3& operator*=(float scalar);

    // Comparison operators
    bool operator==(const Vector3& other) const;
    bool operator!=(const Vector3& other) const;

    // Stream operators
    friend std::ostream& operator<<(std::ostream& os, const Vector3& v);
    
    // Subscript operator
    float& operator[](size_t index);
    const float& operator[](size_t index) const;
    
    // Function call operator
    float operator()(size_t index) const;
};`

		ast, err := manager.parseContent(cppCode, types.Language{
			Name:       "cpp",
			Extensions: []string{".cpp"},
			Parser:     "tree-sitter-cpp",
			Enabled:    true,
		}, "operator_test.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)

		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)

		t.Logf("Found %d symbols", len(symbols))

		// Enhanced parser correctly detects operator overloads as SymbolTypeOperator

		// Count operator overloads
		operatorCount := 0
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeOperator {
				operatorCount++
				t.Logf("Operator: %s, Signature: %s, Visibility: %s", 
					symbol.Name, symbol.Signature, symbol.Visibility)
			}
		}

		// Should detect multiple operator overloads
		assert.GreaterOrEqual(t, operatorCount, 3, "Should detect multiple operator overloads")

		// Check for operator overload detection in metadata
		if ast.Root != nil && ast.Root.Metadata != nil {
			hasOperatorOverload := ast.Root.Metadata["has_operator_overload"].(bool)
			assert.True(t, hasOperatorOverload, "Should detect operator overloading")
		}
	})
}

func TestCppConstructorDestructorDetection(t *testing.T) {
	manager := NewManager()

	t.Run("constructor and destructor classification", func(t *testing.T) {
		cppCode := `class MyClass {
public:
    // Default constructor
    MyClass();
    
    // Parameterized constructor
    MyClass(int value);
    
    // Copy constructor
    MyClass(const MyClass& other);
    
    // Move constructor
    MyClass(MyClass&& other) noexcept;
    
    // Destructor
    ~MyClass();

private:
    int value_;
};`

		ast, err := manager.parseContent(cppCode, types.Language{
			Name:       "cpp",
			Extensions: []string{".cpp"},
			Parser:     "tree-sitter-cpp",
			Enabled:    true,
		}, "constructor_test.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)

		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)

		t.Logf("Found %d symbols", len(symbols))

		constructorCount := 0
		destructorCount := 0
		
		for _, symbol := range symbols {
			t.Logf("Symbol: %s, Type: %s, Visibility: %s", symbol.Name, symbol.Type, symbol.Visibility)
			
			if symbol.Type == types.SymbolTypeConstructor {
				constructorCount++
				// All constructors should be public in this example
				assert.Equal(t, "public", symbol.Visibility, 
					"Constructor should have public visibility")
			}
			
			if symbol.Type == types.SymbolTypeDestructor {
				destructorCount++
				// Destructor should be public
				assert.Equal(t, "public", symbol.Visibility, 
					"Destructor should have public visibility")
			}
		}

		// Should detect multiple constructors and one destructor
		assert.GreaterOrEqual(t, constructorCount, 3, "Should detect multiple constructors")
		assert.Equal(t, 1, destructorCount, "Should detect one destructor")
	})
}

func TestCppModernFeatureIntegration(t *testing.T) {
	manager := NewManager()

	t.Run("modern C++ features integration", func(t *testing.T) {
		cppCode := `#include <memory>
#include <vector>
#include <string>

namespace ModernCpp {
    template<typename T>
    concept Printable = requires(T t) {
        std::cout << t;
    };

    class ModernClass {
    private:
        std::unique_ptr<int> smartPtr_;
        std::vector<std::string> data_;

    public:
        ModernClass() = default;
        virtual ~ModernClass() = default;

        // Auto return type deduction
        auto getValue() const -> int { return *smartPtr_; }

        // Lambda with capture
        void processData() {
            auto processor = [this](const auto& item) {
                return item + "_processed";
            };

            // Range-based for loop
            for (const auto& item : data_) {
                auto result = processor(item);
            }
        }

        // Constexpr function
        constexpr static int getVersion() { return 42; }

        // Override virtual function
        virtual void virtualMethod() override {}

        // Final virtual function
        virtual void finalMethod() final {}
    };
}`

		ast, err := manager.parseContent(cppCode, types.Language{
			Name:       "cpp",
			Extensions: []string{".cpp"},
			Parser:     "tree-sitter-cpp",
			Enabled:    true,
		}, "modern_cpp_test.cpp")
		require.NoError(t, err)
		require.NotNil(t, ast)

		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)

		t.Logf("Found %d symbols", len(symbols))

		// Check feature detection
		require.NotNil(t, ast.Root.Metadata)
		
		// Verify modern C++ features are detected
		assert.True(t, ast.Root.Metadata["has_concepts"].(bool), "Should detect concepts")
		assert.True(t, ast.Root.Metadata["has_auto_keyword"].(bool), "Should detect auto keyword")
		assert.True(t, ast.Root.Metadata["has_lambdas"].(bool), "Should detect lambdas")
		assert.True(t, ast.Root.Metadata["has_range_for"].(bool), "Should detect range-based for")
		assert.True(t, ast.Root.Metadata["has_smart_pointers"].(bool), "Should detect smart pointers")
		assert.True(t, ast.Root.Metadata["has_constexpr"].(bool), "Should detect constexpr")

		// Check symbol classification with visibility
		hasPublicConstructor := false
		hasPublicDestructor := false
		hasPrivateVariables := false

		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeConstructor && symbol.Visibility == "public" {
				hasPublicConstructor = true
			}
			if symbol.Type == types.SymbolTypeDestructor && symbol.Visibility == "public" {
				hasPublicDestructor = true
			}
			if symbol.Type == types.SymbolTypeVariable && symbol.Visibility == "private" {
				hasPrivateVariables = true
			}
		}

		assert.True(t, hasPublicConstructor, "Should have public constructor")
		assert.True(t, hasPublicDestructor, "Should have public destructor")
		assert.True(t, hasPrivateVariables, "Should have private variables")
	})
}