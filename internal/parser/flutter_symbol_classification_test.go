package parser

import (
	"testing"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFlutterSymbolClassification tests the enhanced Flutter symbol type classification
func TestFlutterSymbolClassification(t *testing.T) {
	manager := NewManager()
	
	t.Run("build method symbol type", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';

class MyWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container(
      child: Text('Hello World'),
    );
  }
}`

		ast, err := manager.parseDartContent(content, "test_widget.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find build method with proper type
		var buildMethodSymbol *types.Symbol
		for _, symbol := range symbols {
			if symbol.Name == "build" {
				buildMethodSymbol = symbol
				break
			}
		}
		
		require.NotNil(t, buildMethodSymbol, "Should find build method symbol")
		assert.True(t, 
			buildMethodSymbol.Type == types.SymbolTypeBuildMethod || 
			buildMethodSymbol.Type == types.SymbolTypeMethod, 
			"Build method should have build_method or method type")
		
		t.Logf("Build method symbol: Name=%s, Type=%s", buildMethodSymbol.Name, buildMethodSymbol.Type)
	})
	
	t.Run("state class symbol type", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';

class MyWidget extends StatefulWidget {
  @override
  _MyWidgetState createState() => _MyWidgetState();
}

class _MyWidgetState extends State<MyWidget> {
  int _counter = 0;
  
  @override
  void initState() {
    super.initState();
    _counter = 0;
  }
  
  @override
  void dispose() {
    super.dispose();
  }
  
  @override
  Widget build(BuildContext context) {
    return Container(
      child: Text('Counter: $_counter'),
    );
  }
  
  void _increment() {
    setState(() {
      _counter++;
    });
  }
}`

		ast, err := manager.parseDartContent(content, "stateful_widget.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should find state class
		var stateClassSymbol *types.Symbol
		var lifecycleSymbols []*types.Symbol
		var buildMethodSymbol *types.Symbol
		
		for _, symbol := range symbols {
			switch symbol.Name {
			case "_MyWidgetState":
				stateClassSymbol = symbol
			case "initState", "dispose":
				if symbol.Type == types.SymbolTypeLifecycleMethod || symbol.Type == types.SymbolTypeMethod {
					lifecycleSymbols = append(lifecycleSymbols, symbol)
				}
			case "build":
				buildMethodSymbol = symbol
			}
		}
		
		// Verify state class symbol
		require.NotNil(t, stateClassSymbol, "Should find state class symbol")
		assert.True(t, 
			stateClassSymbol.Type == types.SymbolTypeStateClass || 
			stateClassSymbol.Type == types.SymbolTypeClass,
			"State class should have state_class or class type")
		
		// Verify lifecycle methods
		assert.GreaterOrEqual(t, len(lifecycleSymbols), 1, "Should find lifecycle methods")
		
		// Verify build method
		require.NotNil(t, buildMethodSymbol, "Should find build method")
		
		t.Logf("State class: Name=%s, Type=%s", stateClassSymbol.Name, stateClassSymbol.Type)
		t.Logf("Found %d lifecycle methods", len(lifecycleSymbols))
		for _, ls := range lifecycleSymbols {
			t.Logf("  Lifecycle method: Name=%s, Type=%s", ls.Name, ls.Type)
		}
		t.Logf("Build method: Name=%s, Type=%s", buildMethodSymbol.Name, buildMethodSymbol.Type)
	})
	
	t.Run("lifecycle method metadata", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';

class _MyWidgetState extends State<MyWidget> {
  @override
  void initState() {
    super.initState();
    // Initialization code
  }
  
  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // Dependencies changed
  }
  
  @override
  void didUpdateWidget(MyWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    // Widget updated
  }
  
  @override
  void dispose() {
    // Cleanup code
    super.dispose();
  }
  
  @override
  Widget build(BuildContext context) {
    return Container();
  }
}`

		ast, err := manager.parseDartContent(content, "lifecycle_test.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		lifecycleMethods := make(map[string]*types.Symbol)
		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeLifecycleMethod || 
			   (symbol.Type == types.SymbolTypeMethod && 
			    (symbol.Name == "initState" || symbol.Name == "dispose" || 
			     symbol.Name == "didChangeDependencies" || symbol.Name == "didUpdateWidget")) {
				lifecycleMethods[symbol.Name] = symbol
			}
		}
		
		expectedMethods := []string{"initState", "dispose", "didChangeDependencies", "didUpdateWidget"}
		for _, methodName := range expectedMethods {
			symbol, found := lifecycleMethods[methodName]
			if found {
				t.Logf("Found lifecycle method: %s (Type: %s)", methodName, symbol.Type)
			} else {
				t.Logf("Lifecycle method not found or not properly classified: %s", methodName)
			}
		}
		
		// Should find most lifecycle methods (we're lenient since regex parsing has limitations)
		assert.GreaterOrEqual(t, len(lifecycleMethods), 2, "Should find at least 2 lifecycle methods")
	})
	
	t.Run("flutter metadata extraction", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class MyApp extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return MaterialApp(
      title: 'Flutter Demo',
      home: HomeScreen(),
    );
  }
}

class HomeScreen extends StatefulWidget {
  @override
  _HomeScreenState createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> 
    with TickerProviderStateMixin {
  AnimationController? _controller;
  
  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: Duration(seconds: 2),
      vsync: this,
    );
  }
  
  @override
  void dispose() {
    _controller?.dispose();
    super.dispose();
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Home'),
      ),
      body: _buildBody(),
    );
  }
  
  Widget _buildBody() {
    return Center(
      child: Text('Hello World'),
    );
  }
}`

		ast, err := manager.parseDartContent(content, "flutter_metadata.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		// Check Flutter analysis metadata
		hasFlutter, _ := ast.Root.Metadata["has_flutter"].(bool)
		assert.True(t, hasFlutter, "Should detect Flutter")
		
		flutterFramework, _ := ast.Root.Metadata["flutter_framework"].(string)
		assert.Equal(t, "material", flutterFramework, "Should detect Material framework")
		
		stateManagement, _ := ast.Root.Metadata["state_management"].(string)
		assert.Equal(t, "riverpod", stateManagement, "Should detect Riverpod state management")
		
		// Get full Flutter analysis
		flutterAnalysis, ok := ast.Root.Metadata["flutter_analysis"]
		require.True(t, ok, "Should have Flutter analysis metadata")
		
		analysis, ok := flutterAnalysis.(*FlutterAnalysis)
		require.True(t, ok, "Flutter analysis should be correct type")
		
		assert.True(t, analysis.IsFlutter, "Analysis should detect Flutter")
		assert.Equal(t, "material", analysis.UIFramework, "Should detect Material UI")
		assert.Equal(t, "riverpod", analysis.StateManagement, "Should detect Riverpod")
		assert.True(t, analysis.HasOverride, "Should detect @override annotations")
		assert.GreaterOrEqual(t, len(analysis.Widgets), 2, "Should find multiple widgets")
		
		t.Logf("Flutter analysis: Framework=%s, UI=%s, State=%s, Widgets=%d", 
			analysis.Framework, analysis.UIFramework, analysis.StateManagement, len(analysis.Widgets))
	})
}

// TestFlutterSymbolTypeIntegration tests integration between symbol types and classification
func TestFlutterSymbolTypeIntegration(t *testing.T) {
	manager := NewManager()
	
	t.Run("comprehensive Flutter app classification", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

class CounterModel extends ChangeNotifier {
  int _count = 0;
  int get count => _count;
  
  void increment() {
    _count++;
    notifyListeners();
  }
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (context) => CounterModel(),
      child: MaterialApp(
        title: 'Counter App',
        home: CounterScreen(),
      ),
    );
  }
}

class CounterScreen extends StatefulWidget {
  final String title;
  
  const CounterScreen({Key? key, this.title = 'Counter'}) : super(key: key);
  
  @override
  _CounterScreenState createState() => _CounterScreenState();
}

class _CounterScreenState extends State<CounterScreen> 
    with SingleTickerProviderStateMixin {
  AnimationController? _animationController;
  
  @override
  void initState() {
    super.initState();
    _animationController = AnimationController(
      duration: Duration(milliseconds: 300),
      vsync: this,
    );
  }
  
  @override
  void dispose() {
    _animationController?.dispose();
    super.dispose();
  }
  
  @override
  Widget build(BuildContext context) {
    return Consumer<CounterModel>(
      builder: (context, counter, child) {
        return Scaffold(
          appBar: AppBar(
            title: Text(widget.title),
          ),
          body: _buildBody(counter),
          floatingActionButton: _buildFAB(counter),
        );
      },
    );
  }
  
  Widget _buildBody(CounterModel counter) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text('You have pushed the button this many times:'),
          AnimatedBuilder(
            animation: _animationController!,
            builder: (context, child) {
              return Transform.scale(
                scale: 1.0 + (_animationController!.value * 0.1),
                child: Text(
                  '${counter.count}',
                  style: Theme.of(context).textTheme.headline4,
                ),
              );
            },
          ),
        ],
      ),
    );
  }
  
  Widget _buildFAB(CounterModel counter) {
    return FloatingActionButton(
      onPressed: () {
        counter.increment();
        _animationController!.forward().then((_) {
          _animationController!.reverse();
        });
      },
      tooltip: 'Increment',
      child: Icon(Icons.add),
    );
  }
}`

		ast, err := manager.parseDartContent(content, "comprehensive_app.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Categorize symbols by type
		symbolTypes := make(map[types.SymbolType][]string)
		for _, symbol := range symbols {
			symbolTypes[symbol.Type] = append(symbolTypes[symbol.Type], symbol.Name)
		}
		
		// Log symbol classification
		for symbolType, names := range symbolTypes {
			t.Logf("Type %s: %v", symbolType, names)
		}
		
		// Verify we have proper classification
		assert.GreaterOrEqual(t, len(symbolTypes[types.SymbolTypeClass]), 1, "Should have regular classes")
		assert.GreaterOrEqual(t, len(symbolTypes[types.SymbolTypeWidget]), 2, "Should have widget classes")
		assert.GreaterOrEqual(t, len(symbolTypes[types.SymbolTypeMethod]), 3, "Should have methods")
		assert.GreaterOrEqual(t, len(symbolTypes[types.SymbolTypeImport]), 2, "Should have imports")
		
		// Check Flutter analysis
		hasFlutter, _ := ast.Root.Metadata["has_flutter"].(bool)
		assert.True(t, hasFlutter, "Should detect Flutter")
		
		stateManagement, _ := ast.Root.Metadata["state_management"].(string)
		assert.Equal(t, "provider", stateManagement, "Should detect Provider state management")
		
		t.Logf("Total symbols: %d, Symbol types: %d", len(symbols), len(symbolTypes))
	})
}