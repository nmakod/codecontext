package parser

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlutterDetector(t *testing.T) {
	detector := NewFlutterDetector()
	
	t.Run("basic Flutter app", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Flutter Demo',
      home: MyHomePage(),
    );
  }
}

class MyHomePage extends StatefulWidget {
  @override
  _MyHomePageState createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  int _counter = 0;

  void _incrementCounter() {
    setState(() {
      _counter++;
    });
  }

  @override
  void initState() {
    super.initState();
  }

  @override
  void dispose() {
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Demo')),
      body: Column(
        children: [
          Text('Count: $_counter'),
          ElevatedButton(
            onPressed: _incrementCounter,
            child: Text('Increment'),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _incrementCounter,
        child: Icon(Icons.add),
      ),
    );
  }
}`
		
		analysis := detector.AnalyzeFlutterContent(content)
		
		// Basic Flutter detection
		assert.True(t, analysis.IsFlutter, "Should detect Flutter")
		assert.Equal(t, "flutter", analysis.Framework)
		assert.Equal(t, "material", analysis.UIFramework)
		
		// Widget analysis
		assert.GreaterOrEqual(t, len(analysis.Widgets), 2, "Should find at least 2 widgets")
		
		// Find specific widgets
		var foundMyApp, foundMyHomePage, foundState bool
		for _, widget := range analysis.Widgets {
			switch widget.Name {
			case "MyApp":
				foundMyApp = true
				assert.Equal(t, "stateless", widget.Type)
				assert.True(t, widget.HasBuildMethod)
			case "MyHomePage":
				foundMyHomePage = true
				assert.Equal(t, "stateful", widget.Type)
			case "_MyHomePageState":
				foundState = true
				assert.Equal(t, "state", widget.Type)
				assert.True(t, widget.HasBuildMethod)
			}
		}
		
		assert.True(t, foundMyApp, "Should find MyApp widget")
		assert.True(t, foundMyHomePage, "Should find MyHomePage widget")
		assert.True(t, foundState, "Should find state class")
		
		// State management
		assert.Equal(t, "setState", analysis.StateManagement)
		
		// Features
		assert.Contains(t, analysis.Features, "MaterialApp")
		assert.Contains(t, analysis.Features, "Scaffold")
		assert.Contains(t, analysis.Features, "AppBar")
		assert.Contains(t, analysis.Features, "FloatingActionButton")
		
		// Lifecycle methods
		assert.Contains(t, analysis.LifecycleMethods, "initState")
		assert.Contains(t, analysis.LifecycleMethods, "dispose")
	})
	
	t.Run("Riverpod state management", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

final counterProvider = StateProvider<int>((ref) => 0);

class MyApp extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final count = ref.watch(counterProvider);
    return MaterialApp(
      home: Text('Count: $count'),
    );
  }
}`
		
		analysis := detector.AnalyzeFlutterContent(content)
		
		assert.True(t, analysis.IsFlutter)
		assert.Equal(t, "material", analysis.UIFramework)
		assert.Equal(t, "riverpod", analysis.StateManagement)
		assert.Contains(t, analysis.Features, "MaterialApp")
	})
	
	t.Run("BLoC state management", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class CounterBloc extends Bloc<CounterEvent, int> {
  CounterBloc() : super(0);
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return BlocBuilder<CounterBloc, int>(
      builder: (context, count) {
        return MaterialApp(
          home: Text('Count: $count'),
        );
      },
    );
  }
}`
		
		analysis := detector.AnalyzeFlutterContent(content)
		
		assert.True(t, analysis.IsFlutter)
		assert.Equal(t, "bloc", analysis.StateManagement)
		assert.Contains(t, analysis.Features, "MaterialApp")
	})
	
	t.Run("Cupertino app", func(t *testing.T) {
		content := `import 'package:flutter/cupertino.dart';

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return CupertinoApp(
      title: 'Cupertino Demo',
      home: CupertinoPageScaffold(
        child: Center(child: Text('Hello')),
      ),
    );
  }
}`
		
		analysis := detector.AnalyzeFlutterContent(content)
		
		assert.True(t, analysis.IsFlutter)
		assert.Equal(t, "cupertino", analysis.UIFramework)
		assert.Contains(t, analysis.Features, "CupertinoApp")
	})
	
	t.Run("plain Dart code", func(t *testing.T) {
		content := `void main() {
  print('Hello World');
}

class MyClass {
  void method() {}
}`
		
		analysis := detector.AnalyzeFlutterContent(content)
		
		assert.False(t, analysis.IsFlutter)
		assert.Equal(t, "none", analysis.Framework)
		assert.Equal(t, "none", analysis.UIFramework)
		assert.Equal(t, "none", analysis.StateManagement)
		assert.Empty(t, analysis.Widgets)
		assert.Empty(t, analysis.Features)
	})
	
	t.Run("navigation detection", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';

class MyApp extends StatelessWidget {
  void navigateToHome() {
    Navigator.pushNamed(context, '/home');
  }
  
  @override
  Widget build(BuildContext context) {
    return MaterialApp();
  }
}`
		
		analysis := detector.AnalyzeFlutterContent(content)
		
		assert.True(t, analysis.IsFlutter)
		assert.True(t, analysis.HasNavigation)
	})
}

func TestFlutterIntegration(t *testing.T) {
	manager := NewManager()
	
	t.Run("enhanced Flutter analysis integration", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

final counterProvider = StateProvider<int>((ref) => 0);

class MyApp extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return MaterialApp(
      title: 'Flutter Demo',
      home: Scaffold(
        appBar: AppBar(title: Text('Demo')),
        body: Text('Hello'),
      ),
    );
  }
}`
		
		ast, err := manager.parseDartContent(content, "main.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		require.NotNil(t, ast.Root)
		require.NotNil(t, ast.Root.Metadata)
		
		// Check basic Flutter detection
		hasFlutter, _ := ast.Root.Metadata["has_flutter"].(bool)
		assert.True(t, hasFlutter)
		
		// Check enhanced analysis
		flutterFramework, _ := ast.Root.Metadata["flutter_framework"].(string)
		assert.Equal(t, "material", flutterFramework)
		
		stateManagement, _ := ast.Root.Metadata["state_management"].(string)
		assert.Equal(t, "riverpod", stateManagement)
		
		// Check full analysis object
		fullAnalysis, ok := ast.Root.Metadata["flutter_analysis"].(*FlutterAnalysis)
		require.True(t, ok, "Should have full Flutter analysis")
		
		assert.True(t, fullAnalysis.IsFlutter)
		assert.Equal(t, "material", fullAnalysis.UIFramework)
		assert.Equal(t, "riverpod", fullAnalysis.StateManagement)
		assert.Contains(t, fullAnalysis.Features, "MaterialApp")
		assert.Contains(t, fullAnalysis.Features, "Scaffold")
		assert.Contains(t, fullAnalysis.Features, "AppBar")
		
		// Check widgets
		assert.GreaterOrEqual(t, len(fullAnalysis.Widgets), 1)
		var foundConsumerWidget bool
		for _, widget := range fullAnalysis.Widgets {
			if widget.Name == "MyApp" {
				foundConsumerWidget = true
				// Note: Our current implementation may not detect ConsumerWidget 
				// as a specific type, but will detect it as a class
				break
			}
		}
		assert.True(t, foundConsumerWidget, "Should find MyApp widget")
	})
}