package parser

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkDartParsing(b *testing.B) {
	manager := NewManager()
	
	// Sample Flutter app content for benchmarking
	flutterContent := `import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

final counterProvider = StateProvider<int>((ref) => 0);

class MyApp extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final count = ref.watch(counterProvider);
    return MaterialApp(
      title: 'Flutter Demo',
      theme: ThemeData(
        primarySwatch: Colors.blue,
      ),
      home: MyHomePage(title: 'Flutter Demo Home Page'),
    );
  }
}

class MyHomePage extends StatefulWidget {
  MyHomePage({Key? key, required this.title}) : super(key: key);

  final String title;

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
      appBar: AppBar(
        title: Text(widget.title),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            Text('You have pushed the button this many times:'),
            Text(
              '$_counter',
              style: Theme.of(context).textTheme.headline4,
            ),
            ElevatedButton(
              onPressed: _incrementCounter,
              child: Text('Increment'),
            ),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _incrementCounter,
        tooltip: 'Increment',
        child: Icon(Icons.add),
      ),
    );
  }
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ast, err := manager.parseDartContent(flutterContent, "test.dart")
		if err != nil {
			b.Fatal(err)
		}
		
		// Extract symbols to test complete pipeline
		_, err = manager.ExtractSymbols(ast)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFlutterDetection(b *testing.B) {
	detector := NewFlutterDetector()
	
	content := `import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class MyApp extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return MaterialApp(
      home: Scaffold(
        appBar: AppBar(title: Text('Demo')),
        body: Column(
          children: [
            Text('Hello'),
            ElevatedButton(onPressed: () {}, child: Text('Click')),
          ],
        ),
      ),
    );
  }
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analysis := detector.AnalyzeFlutterContent(content)
		if !analysis.IsFlutter {
			b.Fatal("Should detect Flutter")
		}
	}
}

func BenchmarkLargeFileParsing(b *testing.B) {
	manager := NewManager()
	
	// Create a large Dart file by repeating class definitions
	baseClass := `
class TestClass%d {
  int value = %d;
  String name = "test%d";
  
  void method%d() {
    print("Method %d");
  }
  
  int calculate%d(int x, int y) {
    return x + y + %d;
  }
}
`
	
	var content strings.Builder
	content.WriteString("import 'dart:io';\nimport 'dart:math';\n\n")
	
	// Generate 100 classes
	for i := 0; i < 100; i++ {
		content.WriteString(fmt.Sprintf(baseClass, i, i, i, i, i, i, i))
	}
	
	largeContent := content.String()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ast, err := manager.parseDartContent(largeContent, "large_test.dart")
		if err != nil {
			b.Fatal(err)
		}
		
		// Extract symbols to test complete pipeline
		symbols, err := manager.ExtractSymbols(ast)
		if err != nil {
			b.Fatal(err)
		}
		
		// Should find many symbols
		if len(symbols) < 200 { // At least 2 symbols per class (class + method)
			b.Fatalf("Expected many symbols, got %d", len(symbols))
		}
	}
}

// Performance validation test
func TestDartPerformanceValidation(t *testing.T) {
	t.Run("basic parsing performance", func(t *testing.T) {
		manager := NewManager()
		content := `import 'package:flutter/material.dart';

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(home: Text('Hello'));
  }
}`
		
		// Should parse quickly
		ast, err := manager.parseDartContent(content, "test.dart")
		if err != nil {
			t.Fatal(err)
		}
		
		if ast == nil {
			t.Fatal("AST should not be nil")
		}
		
		t.Logf("Successfully parsed Dart content with %d root children", len(ast.Root.Children))
	})
	
	t.Run("flutter detection performance", func(t *testing.T) {
		detector := NewFlutterDetector()
		content := `import 'package:flutter/material.dart';

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        appBar: AppBar(title: Text('Demo')),
        body: Center(child: Text('Hello World')),
      ),
    );
  }
}`
		
		analysis := detector.AnalyzeFlutterContent(content)
		
		if !analysis.IsFlutter {
			t.Fatal("Should detect Flutter")
		}
		
		if analysis.UIFramework != "material" {
			t.Fatalf("Expected material framework, got %s", analysis.UIFramework)
		}
		
		t.Logf("Flutter analysis completed: Framework=%s, UI=%s, Features=%v", 
			analysis.Framework, analysis.UIFramework, analysis.Features)
	})
}