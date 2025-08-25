package parser

import (
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDartEnumDetection tests enum parsing and detection
func TestDartEnumDetection(t *testing.T) {
	manager := NewManager()
	
	t.Run("basic enum declaration", func(t *testing.T) {
		content := `enum Color {
  red,
  green,
  blue,
  yellow,
}`

		ast, err := manager.parseDartContent(content, "color_enum.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find enum symbol
		var enumSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "Color" && symbol.Type == types.SymbolTypeEnum {
				enumSymbol = symbol
				break
			}
		}
		
		require.NotNil(t, enumSymbol, "Should find enum symbol")
		assert.Equal(t, types.SymbolTypeEnum, enumSymbol.Type, "Should be enum type")
		assert.Equal(t, "Color", enumSymbol.Name, "Should have correct name")
		
		t.Logf("Found enum: %s", enumSymbol.Name)
	})
	
	t.Run("enhanced enum with constructor", func(t *testing.T) {
		content := `enum Planet {
  mercury(3.303e+23, 2.4397e6),
  venus(4.869e+24, 6.0518e6),
  earth(5.976e+24, 6.37814e6),
  mars(6.421e+23, 3.3972e6);

  const Planet(this.mass, this.radius);

  final double mass;      // in kilograms
  final double radius;    // in meters

  double get surfaceGravity => 6.67300E-11 * mass / (radius * radius);
}`

		ast, err := manager.parseDartContent(content, "planet_enum.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find enhanced enum
		var planetEnum *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "Planet" && symbol.Type == types.SymbolTypeEnum {
				planetEnum = symbol
				break
			}
		}
		
		require.NotNil(t, planetEnum, "Should find Planet enum")
		t.Logf("Found enhanced enum: %s", planetEnum.Name)
	})
	
	t.Run("generic enum (Dart 3.0+)", func(t *testing.T) {
		content := `enum Result<T> {
  success<T>(T value),
  error<String>(String message);

  const Result(this.data);
  
  final T data;
  
  bool get isSuccess => this == success;
  bool get isError => this == error;
}`

		ast, err := manager.parseDartContent(content, "result_enum.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find generic enum
		var resultEnum *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "Result" && symbol.Type == types.SymbolTypeEnum {
				resultEnum = symbol
				break
			}
		}
		
		require.NotNil(t, resultEnum, "Should find Result enum")
		t.Logf("Found generic enum: %s", resultEnum.Name)
	})
	
	t.Run("enum implementing interface", func(t *testing.T) {
		content := `abstract class Comparable<T> {
  int compareTo(T other);
}

enum Size implements Comparable<Size> {
  small,
  medium,
  large;

  @override
  int compareTo(Size other) {
    return index.compareTo(other.index);
  }
  
  bool get isLarge => this == Size.large;
}`

		ast, err := manager.parseDartContent(content, "size_enum.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Debug: log all found symbols
		t.Logf("Found %d symbols:", len(symbols))
		for _, symbol := range symbols {
			t.Logf("  Symbol: %s (type: %s)", symbol.Name, symbol.Type)
		}
		
		// Should find both abstract class and enum
		var abstractClass, sizeEnum *types.Symbol
		for _, symbol := range symbols {
			switch {
			case symbol.Name == "Comparable" && symbol.Type == types.SymbolTypeClass:
				abstractClass = symbol
			case symbol.Name == "Size" && symbol.Type == types.SymbolTypeEnum:
				sizeEnum = symbol
			}
		}
		
		assert.NotNil(t, abstractClass, "Should find Comparable abstract class")
		assert.NotNil(t, sizeEnum, "Should find Size enum")
		
		if abstractClass != nil {
			t.Logf("Found abstract class: %s", abstractClass.Name)
		}
		if sizeEnum != nil {
			t.Logf("Found implementing enum: %s", sizeEnum.Name)
		}
	})
}

// TestDartTypedefDetection tests typedef parsing and detection
func TestDartTypedefDetection(t *testing.T) {
	manager := NewManager()
	
	t.Run("basic function typedef", func(t *testing.T) {
		content := `typedef IntCallback = void Function(int value);

void processNumbers(List<int> numbers, IntCallback callback) {
  for (final number in numbers) {
    callback(number);
  }
}`

		ast, err := manager.parseDartContent(content, "callback_typedef.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find typedef symbol
		var typedefSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "IntCallback" && symbol.Type == types.SymbolTypeTypedef {
				typedefSymbol = symbol
				break
			}
		}
		
		require.NotNil(t, typedefSymbol, "Should find typedef symbol")
		assert.Equal(t, types.SymbolTypeTypedef, typedefSymbol.Type, "Should be typedef type")
		assert.Equal(t, "IntCallback", typedefSymbol.Name, "Should have correct name")
		
		t.Logf("Found typedef: %s with signature: %s", typedefSymbol.Name, typedefSymbol.Signature)
	})
	
	t.Run("generic function typedef", func(t *testing.T) {
		content := `typedef Converter<S, T> = T Function(S source);
typedef Predicate<T> = bool Function(T item);
typedef Factory<T> = T Function();

class Processor<T> {
  final List<T> items = [];
  
  void addWhere(T item, Predicate<T> predicate) {
    if (predicate(item)) {
      items.add(item);
    }
  }
  
  List<R> convertAll<R>(Converter<T, R> converter) {
    return items.map(converter).toList();
  }
}`

		ast, err := manager.parseDartContent(content, "generic_typedefs.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Count typedefs
		typedefCount := 0
		typedefNames := []string{}
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeTypedef {
				typedefCount++
				typedefNames = append(typedefNames, symbol.Name)
			}
		}
		
		assert.GreaterOrEqual(t, typedefCount, 3, "Should find at least 3 typedefs")
		t.Logf("Found %d typedefs: %v", typedefCount, typedefNames)
	})
	
	t.Run("class type alias", func(t *testing.T) {
		content := `typedef StringList = List<String>;
typedef IntMap = Map<String, int>;
typedef JsonObject = Map<String, dynamic>;

class DataProcessor {
  void processStrings(StringList strings) {
    // Process string list
  }
  
  void processData(IntMap data) {
    // Process integer map
  }
  
  JsonObject toJson() {
    return <String, dynamic>{
      'processed': true,
    };
  }
}`

		ast, err := manager.parseDartContent(content, "type_aliases.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Verify type aliases
		expectedTypedefs := []string{"StringList", "IntMap", "JsonObject"}
		foundTypedefs := make(map[string]bool)
		
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeTypedef {
				foundTypedefs[symbol.Name] = true
			}
		}
		
		for _, expected := range expectedTypedefs {
			assert.True(t, foundTypedefs[expected], "Should find typedef: %s", expected)
		}
		
		t.Logf("Found type aliases: %v", foundTypedefs)
	})
}

// TestAdvancedTypePatterns tests complex type patterns and generics
func TestAdvancedTypePatterns(t *testing.T) {
	manager := NewManager()
	
	t.Run("complex generic types", func(t *testing.T) {
		content := `typedef AsyncResult<T> = Future<Result<T, String>>;
typedef EventHandler<T extends Event> = void Function(T event);
typedef Repository<T extends Entity> = Future<List<T>> Function(Query query);

enum Result<T, E> {
  success(T value),
  failure(E error);

  const Result(this.data);
  
  final dynamic data;
  
  bool get isSuccess => this is Result<T, Never>;
}

abstract class Entity {
  String get id;
}

abstract class Event {
  DateTime get timestamp;
}

class UserEvent extends Event {
  final String userId;
  final String action;
  
  const UserEvent(this.userId, this.action);
  
  @override
  DateTime get timestamp => DateTime.now();
}

class User extends Entity {
  final String name;
  final String email;
  
  const User(this.name, this.email);
  
  @override
  String get id => email;
}`

		ast, err := manager.parseDartContent(content, "advanced_types.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Categorize symbols
		symbolTypes := make(map[types.SymbolType]int)
		for _, symbol := range symbols {
			symbolTypes[symbol.Type]++
		}
		
		assert.GreaterOrEqual(t, symbolTypes[types.SymbolTypeTypedef], 3, "Should find typedefs")
		assert.GreaterOrEqual(t, symbolTypes[types.SymbolTypeEnum], 1, "Should find enum")
		assert.GreaterOrEqual(t, symbolTypes[types.SymbolTypeClass], 3, "Should find classes")
		
		t.Logf("Advanced types - Symbol counts: %+v", symbolTypes)
	})
	
	t.Run("real-world enum and typedef usage", func(t *testing.T) {
		content := `// HTTP Status codes
enum HttpStatus {
  ok(200, 'OK'),
  created(201, 'Created'),
  badRequest(400, 'Bad Request'),
  unauthorized(401, 'Unauthorized'),
  forbidden(403, 'Forbidden'),
  notFound(404, 'Not Found'),
  internalServerError(500, 'Internal Server Error');

  const HttpStatus(this.code, this.message);
  
  final int code;
  final String message;
  
  bool get isSuccess => code >= 200 && code < 300;
  bool get isError => code >= 400;
}

// API types
typedef ApiResponse<T> = Future<Result<T, ApiError>>;
typedef JsonMap = Map<String, dynamic>;
typedef RequestHandler<T> = ApiResponse<T> Function(JsonMap request);

class ApiError {
  final String message;
  final HttpStatus status;
  
  const ApiError(this.message, this.status);
  
  @override
  String toString() => 'ApiError: ${status.message} (${status.code}) - $message';
}

class Result<T, E> {
  final T? data;
  final E? error;
  
  const Result.success(this.data) : error = null;
  const Result.failure(this.error) : data = null;
  
  bool get isSuccess => error == null;
  bool get isFailure => error != null;
}

// Usage example
class UserController {
  RequestHandler<User> get getUser => (request) async {
    try {
      final userId = request['id'] as String?;
      if (userId == null) {
        return Result.failure(ApiError('Missing user ID', HttpStatus.badRequest));
      }
      
      final user = await _fetchUser(userId);
      return Result.success(user);
    } catch (e) {
      return Result.failure(ApiError(e.toString(), HttpStatus.internalServerError));
    }
  };
  
  Future<User> _fetchUser(String id) async {
    // Simulate database fetch
    await Future.delayed(Duration(milliseconds: 100));
    return User('John Doe', 'john@example.com');
  }
}`

		ast, err := manager.parseDartContent(content, "api_types.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Debug: log all found symbols
		t.Logf("Found %d symbols:", len(symbols))
		for _, symbol := range symbols {
			t.Logf("  Symbol: %s (type: %s)", symbol.Name, symbol.Type)
		}
		
		// Check for specific types by looking through all symbols
		expectedEnums := []string{"HttpStatus"}
		expectedTypedefs := []string{"ApiResponse", "JsonMap", "RequestHandler"}
		expectedClasses := []string{"ApiError", "Result", "UserController"}
		
		// Check enums
		for _, enumName := range expectedEnums {
			found := false
			for _, symbol := range symbols {
				if symbol.Name == enumName && symbol.Type == types.SymbolTypeEnum {
					found = true
					break
				}
			}
			assert.True(t, found, "Should find enum: %s", enumName)
		}
		
		// Check typedefs
		for _, typedefName := range expectedTypedefs {
			found := false
			for _, symbol := range symbols {
				if symbol.Name == typedefName && symbol.Type == types.SymbolTypeTypedef {
					found = true
					break
				}
			}
			assert.True(t, found, "Should find typedef: %s", typedefName)
		}
		
		// Check classes
		for _, className := range expectedClasses {
			found := false
			for _, symbol := range symbols {
				if symbol.Name == className && symbol.Type == types.SymbolTypeClass {
					found = true
					break
				}
			}
			assert.True(t, found, "Should find class: %s", className)
		}
		
		t.Logf("Real-world API types - Found %d symbols", len(symbols))
		
		// Check for enhanced enum features
		for _, symbol := range symbols {
			if symbol.Name == "HttpStatus" && symbol.Type == types.SymbolTypeEnum {
				t.Logf("HttpStatus enum found with enhanced features")
				break
			}
		}
	})
}

// TestEnumAndTypedefIntegration tests enums and typedefs working together
func TestEnumAndTypedefIntegration(t *testing.T) {
	manager := NewManager()
	
	t.Run("state machine with enums and typedefs", func(t *testing.T) {
		content := `enum State {
  idle,
  loading,
  success,
  error;
  
  bool get canTransitionTo => this != State.error;
}

enum Event {
  start,
  complete,
  fail,
  reset;
}

typedef StateTransition = State Function(State current, Event event);
typedef StateListener = void Function(State oldState, State newState);

class StateMachine {
  State _currentState = State.idle;
  final List<StateListener> _listeners = [];
  
  State get currentState => _currentState;
  
  void addListener(StateListener listener) {
    _listeners.add(listener);
  }
  
  void transition(Event event) {
    final oldState = _currentState;
    final newState = _handleTransition(_currentState, event);
    
    if (newState != oldState) {
      _currentState = newState;
      for (final listener in _listeners) {
        listener(oldState, newState);
      }
    }
  }
  
  State _handleTransition(State current, Event event) {
    switch (current) {
      case State.idle:
        return event == Event.start ? State.loading : current;
      case State.loading:
        switch (event) {
          case Event.complete:
            return State.success;
          case Event.fail:
            return State.error;
          default:
            return current;
        }
      case State.success:
        return event == Event.reset ? State.idle : current;
      case State.error:
        return event == Event.reset ? State.idle : current;
    }
  }
}`

		ast, err := manager.parseDartContent(content, "state_machine.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Count different symbol types
		symbolCounts := make(map[types.SymbolType]int)
		for _, symbol := range symbols {
			symbolCounts[symbol.Type]++
		}
		
		assert.GreaterOrEqual(t, symbolCounts[types.SymbolTypeEnum], 2, "Should find enums")
		assert.GreaterOrEqual(t, symbolCounts[types.SymbolTypeTypedef], 2, "Should find typedefs")
		assert.GreaterOrEqual(t, symbolCounts[types.SymbolTypeClass], 1, "Should find classes")
		assert.GreaterOrEqual(t, symbolCounts[types.SymbolTypeMethod], 3, "Should find methods")
		
		t.Logf("State machine integration - Symbol counts: %+v", symbolCounts)
	})
}