package parser

import (
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDartMixinDetection tests mixin parsing and detection
func TestDartMixinDetection(t *testing.T) {
	manager := NewManager()
	
	t.Run("basic mixin declaration", func(t *testing.T) {
		content := `mixin Flyable {
  void fly() {
    print('Flying!');
  }
  
  bool get canFly => true;
}`

		ast, err := manager.parseDartContent(content, "basic_mixin.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find mixin symbol
		var mixinSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "Flyable" && symbol.Type == types.SymbolTypeMixin {
				mixinSymbol = symbol
				break
			}
		}
		
		require.NotNil(t, mixinSymbol, "Should find mixin symbol")
		assert.Equal(t, types.SymbolTypeMixin, mixinSymbol.Type, "Should be mixin type")
		assert.Equal(t, "Flyable", mixinSymbol.Name, "Should have correct name")
		
		t.Logf("Found mixin: %s", mixinSymbol.Name)
	})
	
	t.Run("mixin with constraints (on clause)", func(t *testing.T) {
		content := `abstract class Animal {
  void makeSound();
}

class Mammal extends Animal {
  @override
  void makeSound() => print('Mammal sound');
}

mixin Walkable on Animal {
  void walk() {
    print('Walking...');
    makeSound(); // Can call Animal methods
  }
  
  int get legCount => 4;
}

class Dog extends Mammal with Walkable {
  @override
  void makeSound() => print('Woof!');
}`

		ast, err := manager.parseDartContent(content, "mixin_constraints.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find mixin with constraint
		var walkableMixin *types.Symbol
		var dogClass *types.Symbol
		
		for _, symbol := range symbols {
			switch {
			case symbol.Name == "Walkable" && symbol.Type == types.SymbolTypeMixin:
				walkableMixin = symbol
			case symbol.Name == "Dog" && (symbol.Type == types.SymbolTypeClass || symbol.Type == types.SymbolTypeWidget):
				dogClass = symbol
			}
		}
		
		require.NotNil(t, walkableMixin, "Should find Walkable mixin")
		require.NotNil(t, dogClass, "Should find Dog class")
		
		t.Logf("Found constrained mixin: %s", walkableMixin.Name)
		t.Logf("Found class with mixin: %s", dogClass.Name)
	})
	
	t.Run("multiple mixins usage", func(t *testing.T) {
		content := `mixin Flyable {
  void fly() => print('Flying');
}

mixin Swimmable {
  void swim() => print('Swimming');
}

mixin Walkable {
  void walk() => print('Walking');
}

class Animal {}

class Duck extends Animal with Flyable, Swimmable, Walkable {
  void quack() => print('Quack!');
}

class Fish extends Animal with Swimmable {
  void bubble() => print('Bubble...');
}

class Bird extends Animal with Flyable, Walkable {
  void chirp() => print('Chirp!');
}`

		ast, err := manager.parseDartContent(content, "multiple_mixins.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Count mixins and classes
		mixinCount := 0
		classCount := 0
		
		for _, symbol := range symbols {
			switch symbol.Type {
			case types.SymbolTypeMixin:
				mixinCount++
				t.Logf("Found mixin: %s", symbol.Name)
			case types.SymbolTypeClass:
				classCount++
				t.Logf("Found class: %s", symbol.Name)
			}
		}
		
		assert.GreaterOrEqual(t, mixinCount, 3, "Should find at least 3 mixins")
		assert.GreaterOrEqual(t, classCount, 4, "Should find at least 4 classes")
	})
}

// TestDartExtensionDetection tests extension method parsing and detection
func TestDartExtensionDetection(t *testing.T) {
	manager := NewManager()
	
	t.Run("basic extension on built-in type", func(t *testing.T) {
		content := `extension StringExtensions on String {
  bool get isValidEmail {
    return contains('@') && contains('.');
  }
  
  String get reversed {
    return split('').reversed.join('');
  }
  
  String capitalize() {
    if (isEmpty) return this;
    return '${this[0].toUpperCase()}${substring(1)}';
  }
}`

		ast, err := manager.parseDartContent(content, "string_extension.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find extension symbol
		var extensionSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "StringExtensions" && symbol.Type == types.SymbolTypeExtension {
				extensionSymbol = symbol
				break
			}
		}
		
		require.NotNil(t, extensionSymbol, "Should find extension symbol")
		assert.Equal(t, types.SymbolTypeExtension, extensionSymbol.Type, "Should be extension type")
		assert.Equal(t, "StringExtensions", extensionSymbol.Name, "Should have correct name")
		
		t.Logf("Found extension: %s", extensionSymbol.Name)
	})
	
	t.Run("extension on custom class", func(t *testing.T) {
		content := `class Point {
  final double x, y;
  Point(this.x, this.y);
}

extension PointExtensions on Point {
  double get distance => sqrt(x * x + y * y);
  
  Point operator +(Point other) {
    return Point(x + other.x, y + other.y);
  }
  
  Point scale(double factor) {
    return Point(x * factor, y * factor);
  }
  
  bool isNear(Point other, {double threshold = 1.0}) {
    final dx = x - other.x;
    final dy = y - other.y;
    return sqrt(dx * dx + dy * dy) <= threshold;
  }
}`

		ast, err := manager.parseDartContent(content, "point_extension.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find both class and extension
		var pointClass, pointExtension *types.Symbol
		
		for _, symbol := range symbols {
			switch {
			case symbol.Name == "Point" && symbol.Type == types.SymbolTypeClass:
				pointClass = symbol
			case symbol.Name == "PointExtensions" && symbol.Type == types.SymbolTypeExtension:
				pointExtension = symbol
			}
		}
		
		require.NotNil(t, pointClass, "Should find Point class")
		require.NotNil(t, pointExtension, "Should find PointExtensions extension")
		
		t.Logf("Found class: %s", pointClass.Name)
		t.Logf("Found extension: %s", pointExtension.Name)
	})
	
	t.Run("generic extension", func(t *testing.T) {
		content := `extension ListExtensions<T> on List<T> {
  T? get firstOrNull => isEmpty ? null : first;
  T? get lastOrNull => isEmpty ? null : last;
  
  List<T> get unique {
    final seen = <T>{};
    return where((element) => seen.add(element)).toList();
  }
  
  void addIfNotNull(T? item) {
    if (item != null) {
      add(item);
    }
  }
  
  List<R> mapIndexed<R>(R Function(int index, T item) mapper) {
    final result = <R>[];
    for (int i = 0; i < length; i++) {
      result.add(mapper(i, this[i]));
    }
    return result;
  }
}`

		ast, err := manager.parseDartContent(content, "generic_extension.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find generic extension
		var extensionSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "ListExtensions" && symbol.Type == types.SymbolTypeExtension {
				extensionSymbol = symbol
				break
			}
		}
		
		require.NotNil(t, extensionSymbol, "Should find generic extension")
		t.Logf("Found generic extension: %s", extensionSymbol.Name)
	})
	
	t.Run("unnamed extension", func(t *testing.T) {
		content := `extension on int {
  bool get isEven => this % 2 == 0;
  bool get isOdd => !isEven;
  
  int get squared => this * this;
  
  String get ordinal {
    if (this >= 11 && this <= 13) return '${this}th';
    switch (this % 10) {
      case 1: return '${this}st';
      case 2: return '${this}nd'; 
      case 3: return '${this}rd';
      default: return '${this}th';
    }
  }
}`

		ast, err := manager.parseDartContent(content, "unnamed_extension.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find unnamed extension (we'll give it a generated name)
		var extensionSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeExtension {
				extensionSymbol = symbol
				break
			}
		}
		
		require.NotNil(t, extensionSymbol, "Should find unnamed extension")
		t.Logf("Found unnamed extension: %s", extensionSymbol.Name)
	})
}

// TestMixinAndExtensionIntegration tests mixins and extensions working together
func TestMixinAndExtensionIntegration(t *testing.T) {
	manager := NewManager()
	
	t.Run("real-world mixin and extension usage", func(t *testing.T) {
		content := `// Mixins for common behavior
mixin Loggable {
  void log(String message) {
    print('[${runtimeType}] $message');
  }
}

mixin Disposable {
  bool _disposed = false;
  bool get isDisposed => _disposed;
  
  void dispose() {
    if (!_disposed) {
      _disposed = true;
      onDispose();
    }
  }
  
  void onDispose();
}

// Extensions for enhanced functionality
extension StreamExtensions<T> on Stream<T> {
  Stream<T> get logItems {
    return map((item) {
      print('Stream item: $item');
      return item;
    });
  }
  
  Future<List<T>> collectToList() async {
    final items = <T>[];
    await for (final item in this) {
      items.add(item);
    }
    return items;
  }
}

extension FutureExtensions<T> on Future<T> {
  Future<T> get logCompletion {
    return then((value) {
      print('Future completed with: $value');
      return value;
    });
  }
  
  Future<R> mapAsync<R>(R Function(T value) mapper) {
    return then((value) => mapper(value));
  }
}

// Classes using mixins
abstract class BaseService with Loggable, Disposable {
  String get name;
  
  void start() {
    log('Starting $name service');
  }
  
  @override
  void onDispose() {
    log('Disposing $name service');
  }
}

class DataService extends BaseService {
  @override
  String get name => 'DataService';
  
  Stream<String> getData() {
    return Stream.fromIterable(['data1', 'data2', 'data3'])
        .logItems;
  }
  
  Future<String> fetchData() {
    return Future.delayed(
      Duration(seconds: 1), 
      () => 'fetched data'
    ).logCompletion;
  }
}`

		ast, err := manager.parseDartContent(content, "integration_example.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Count different symbol types
		symbolCounts := make(map[types.SymbolType]int)
		for _, symbol := range symbols {
			symbolCounts[symbol.Type]++
		}
		
		assert.GreaterOrEqual(t, symbolCounts[types.SymbolTypeMixin], 2, "Should find mixins")
		assert.GreaterOrEqual(t, symbolCounts[types.SymbolTypeExtension], 2, "Should find extensions")
		assert.GreaterOrEqual(t, symbolCounts[types.SymbolTypeClass], 2, "Should find classes")
		assert.GreaterOrEqual(t, symbolCounts[types.SymbolTypeMethod], 5, "Should find methods")
		
		t.Logf("Symbol counts: %+v", symbolCounts)
		
		// Log found symbols by type
		for symbolType, count := range symbolCounts {
			t.Logf("Found %d symbols of type %s", count, symbolType)
		}
	})
}

// TestAdvancedMixinPatterns tests sophisticated mixin usage patterns
func TestAdvancedMixinPatterns(t *testing.T) {
	manager := NewManager()
	
	t.Run("flutter mixin patterns", func(t *testing.T) {
		// Common Flutter mixin patterns
		content := `import 'package:flutter/material.dart';

mixin ValidationMixin {
  String? validateEmail(String? value) {
    if (value == null || value.isEmpty) return 'Email is required';
    if (!value.contains('@')) return 'Invalid email format';
    return null;
  }
  
  String? validatePassword(String? value) {
    if (value == null || value.isEmpty) return 'Password is required';
    if (value.length < 8) return 'Password must be at least 8 characters';
    return null;
  }
}

mixin FormMixin<T extends StatefulWidget> on State<T> {
  final GlobalKey<FormState> formKey = GlobalKey<FormState>();
  
  bool validateForm() {
    return formKey.currentState?.validate() ?? false;
  }
  
  void resetForm() {
    formKey.currentState?.reset();
  }
}

class LoginScreen extends StatefulWidget {
  @override
  _LoginScreenState createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> 
    with ValidationMixin, FormMixin<LoginScreen> {
  
  final TextEditingController emailController = TextEditingController();
  final TextEditingController passwordController = TextEditingController();
  
  @override
  void dispose() {
    emailController.dispose();
    passwordController.dispose();
    super.dispose();
  }
  
  void _login() {
    if (validateForm()) {
      print('Login attempt with: ${emailController.text}');
    }
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Login')),
      body: Form(
        key: formKey,
        child: Column(
          children: [
            TextFormField(
              controller: emailController,
              validator: validateEmail,
              decoration: InputDecoration(labelText: 'Email'),
            ),
            TextFormField(
              controller: passwordController,
              validator: validatePassword,
              obscureText: true,
              decoration: InputDecoration(labelText: 'Password'),
            ),
            ElevatedButton(
              onPressed: _login,
              child: Text('Login'),
            ),
          ],
        ),
      ),
    );
  }
}`

		ast, err := manager.parseDartContent(content, "flutter_mixins.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should detect Flutter with mixins
		hasFlutter, _ := ast.Root.Metadata["has_flutter"].(bool)
		assert.True(t, hasFlutter, "Should detect Flutter")
		
		// Count symbol types
		mixinCount := 0
		widgetCount := 0
		stateClassCount := 0
		
		for _, symbol := range symbols {
			switch symbol.Type {
			case types.SymbolTypeMixin:
				mixinCount++
				t.Logf("Found mixin: %s", symbol.Name)
			case types.SymbolTypeWidget:
				widgetCount++
				t.Logf("Found widget: %s", symbol.Name)
			case types.SymbolTypeStateClass:
				stateClassCount++
				t.Logf("Found state class: %s", symbol.Name)
			}
		}
		
		assert.GreaterOrEqual(t, mixinCount, 2, "Should find mixins")
		assert.GreaterOrEqual(t, widgetCount, 1, "Should find widgets")
		assert.GreaterOrEqual(t, stateClassCount, 1, "Should find state classes")
	})
}

// TestMixinClassDetection tests classes that use mixins (mixinClass pattern)
func TestMixinClassDetection(t *testing.T) {
	manager := NewManager()
	
	t.Run("class with single mixin", func(t *testing.T) {
		content := `
mixin Flyable {
  void fly() => print('Flying');
}

class Bird with Flyable {
  void chirp() => print('Chirp');
}`

		ast, err := manager.parseDartContent(content, "mixin_class.dart")
		require.NoError(t, err)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find mixin, class, and methods
		var mixinSymbol, classSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "Flyable" && symbol.Type == types.SymbolTypeMixin {
				mixinSymbol = symbol
			}
			if symbol.Name == "Bird" && symbol.Type == types.SymbolTypeClass {
				classSymbol = symbol
			}
		}
		
		require.NotNil(t, mixinSymbol, "Should find Flyable mixin")
		require.NotNil(t, classSymbol, "Should find Bird class")
		
		t.Logf("Found mixin: %s and mixin class: %s", mixinSymbol.Name, classSymbol.Name)
	})
	
	t.Run("class with multiple mixins", func(t *testing.T) {
		content := `
mixin Flyable {
  void fly() => print('Flying');
}

mixin Swimmable {
  void swim() => print('Swimming');
}

class Duck extends Animal with Flyable, Swimmable {
  void quack() => print('Quack');
}

class Animal {
  void makeSound();
}`

		ast, err := manager.parseDartContent(content, "multi_mixin.dart")
		require.NoError(t, err)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Count each type
		var mixinCount, classCount int
		var duckClass *types.Symbol
		
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeMixin {
				mixinCount++
			}
			if symbol.Type == types.SymbolTypeClass {
				classCount++
			}
			if symbol.Name == "Duck" {
				duckClass = symbol
			}
		}
		
		assert.Equal(t, 2, mixinCount, "Should find 2 mixins")
		assert.Equal(t, 2, classCount, "Should find 2 classes") 
		require.NotNil(t, duckClass, "Should find Duck class")
		
		t.Logf("Multi-mixin class: %s with %d mixins and %d total classes", 
			duckClass.Name, mixinCount, classCount)
	})
}