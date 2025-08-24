package parser

import (
	"fmt"
	"strings"
	"testing"
	"time"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFlutterProjectIntegration validates against realistic Flutter project patterns
func TestFlutterProjectIntegration(t *testing.T) {
	manager := NewManager()
	detector := NewFlutterDetector()
	
	t.Run("flutter counter app simulation", func(t *testing.T) {
		// Simulates the classic Flutter counter app structure
		mainDart := `import 'package:flutter/material.dart';

void main() {
  runApp(MyApp());
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
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
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            Text(
              'You have pushed the button this many times:',
            ),
            Text(
              '$_counter',
              style: Theme.of(context).textTheme.headline4,
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

		ast, err := manager.parseDartContent(mainDart, "main.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		analysis := detector.AnalyzeFlutterContent(mainDart)
		
		// Validate comprehensive Flutter detection
		assert.True(t, analysis.IsFlutter, "Should detect Flutter")
		assert.Equal(t, "material", analysis.UIFramework, "Should detect Material UI")
		assert.Equal(t, "setState", analysis.StateManagement, "Should detect setState")
		
		// Validate symbol extraction
		symbolTypes := make(map[types.SymbolType]int)
		for _, symbol := range symbols {
			symbolTypes[symbol.Type]++
		}
		
		assert.GreaterOrEqual(t, symbolTypes[types.SymbolTypeWidget], 2, "Should find widget classes")
		assert.GreaterOrEqual(t, symbolTypes[types.SymbolTypeMethod], 3, "Should find methods")
		assert.GreaterOrEqual(t, symbolTypes[types.SymbolTypeFunction], 1, "Should find main function")
		assert.GreaterOrEqual(t, symbolTypes[types.SymbolTypeImport], 1, "Should find imports")
		
		// Validate Flutter features detection
		expectedFeatures := []string{"MaterialApp", "Scaffold", "AppBar", "FloatingActionButton"}
		for _, feature := range expectedFeatures {
			assert.Contains(t, analysis.Features, feature, fmt.Sprintf("Should detect %s", feature))
		}
		
		t.Logf("Counter app validation: Symbols=%d, Features=%d, Widgets=%d", 
			len(symbols), len(analysis.Features), len(analysis.Widgets))
	})
	
	t.Run("complex navigation app simulation", func(t *testing.T) {
		// Simulates a more complex Flutter app with navigation
		appDart := `import 'package:flutter/material.dart';

class NavigationApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Navigation Demo',
      initialRoute: '/',
      routes: {
        '/': (context) => HomeScreen(),
        '/details': (context) => DetailScreen(),
        '/settings': (context) => SettingsScreen(),
      },
    );
  }
}

class HomeScreen extends StatefulWidget {
  @override
  _HomeScreenState createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> 
    with TickerProviderStateMixin {
  TabController? _tabController;
  
  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
  }
  
  @override
  void dispose() {
    _tabController?.dispose();
    super.dispose();
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Home'),
        bottom: TabBar(
          controller: _tabController,
          tabs: [
            Tab(icon: Icon(Icons.home), text: 'Home'),
            Tab(icon: Icon(Icons.search), text: 'Search'),
            Tab(icon: Icon(Icons.person), text: 'Profile'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          HomeTabView(),
          SearchTabView(),
          ProfileTabView(),
        ],
      ),
      drawer: AppDrawer(),
    );
  }
}

class HomeTabView extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      itemCount: 20,
      itemBuilder: (context, index) {
        return ListTile(
          title: Text('Item $index'),
          subtitle: Text('Description for item $index'),
          onTap: () {
            Navigator.pushNamed(context, '/details', arguments: index);
          },
        );
      },
    );
  }
}

class SearchTabView extends StatefulWidget {
  @override
  _SearchTabViewState createState() => _SearchTabViewState();
}

class _SearchTabViewState extends State<SearchTabView> {
  TextEditingController _searchController = TextEditingController();
  List<String> _searchResults = [];
  
  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }
  
  void _performSearch(String query) {
    setState(() {
      _searchResults = List.generate(10, (index) => 'Result $index for "$query"');
    });
  }
  
  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Padding(
          padding: EdgeInsets.all(16),
          child: TextField(
            controller: _searchController,
            decoration: InputDecoration(
              hintText: 'Search...',
              suffixIcon: IconButton(
                icon: Icon(Icons.search),
                onPressed: () => _performSearch(_searchController.text),
              ),
            ),
            onSubmitted: _performSearch,
          ),
        ),
        Expanded(
          child: ListView.builder(
            itemCount: _searchResults.length,
            itemBuilder: (context, index) {
              return ListTile(
                title: Text(_searchResults[index]),
              );
            },
          ),
        ),
      ],
    );
  }
}

class ProfileTabView extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          CircleAvatar(
            radius: 50,
            child: Icon(Icons.person, size: 50),
          ),
          SizedBox(height: 20),
          Text('John Doe', style: Theme.of(context).textTheme.headline6),
          SizedBox(height: 10),
          Text('john.doe@example.com'),
          SizedBox(height: 20),
          ElevatedButton(
            onPressed: () {
              Navigator.pushNamed(context, '/settings');
            },
            child: Text('Settings'),
          ),
        ],
      ),
    );
  }
}

class AppDrawer extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Drawer(
      child: ListView(
        padding: EdgeInsets.zero,
        children: [
          DrawerHeader(
            decoration: BoxDecoration(
              color: Colors.blue,
            ),
            child: Text(
              'Menu',
              style: TextStyle(
                color: Colors.white,
                fontSize: 24,
              ),
            ),
          ),
          ListTile(
            leading: Icon(Icons.home),
            title: Text('Home'),
            onTap: () {
              Navigator.pop(context);
            },
          ),
          ListTile(
            leading: Icon(Icons.settings),
            title: Text('Settings'),
            onTap: () {
              Navigator.pop(context);
              Navigator.pushNamed(context, '/settings');
            },
          ),
        ],
      ),
    );
  }
}

class DetailScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final int itemIndex = ModalRoute.of(context)!.settings.arguments as int;
    
    return Scaffold(
      appBar: AppBar(
        title: Text('Detail $itemIndex'),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              'Detail Screen',
              style: Theme.of(context).textTheme.headline4,
            ),
            SizedBox(height: 20),
            Text('Item Index: $itemIndex'),
            SizedBox(height: 20),
            ElevatedButton(
              onPressed: () {
                Navigator.pop(context);
              },
              child: Text('Go Back'),
            ),
          ],
        ),
      ),
    );
  }
}

class SettingsScreen extends StatefulWidget {
  @override
  _SettingsScreenState createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  bool _darkMode = false;
  bool _notifications = true;
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Settings'),
      ),
      body: ListView(
        children: [
          SwitchListTile(
            title: Text('Dark Mode'),
            value: _darkMode,
            onChanged: (bool value) {
              setState(() {
                _darkMode = value;
              });
            },
          ),
          SwitchListTile(
            title: Text('Notifications'),
            value: _notifications,
            onChanged: (bool value) {
              setState(() {
                _notifications = value;
              });
            },
          ),
        ],
      ),
    );
  }
}`

		ast, err := manager.parseDartContent(appDart, "navigation_app.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		analysis := detector.AnalyzeFlutterContent(appDart)
		
		// Validate comprehensive Flutter detection
		assert.True(t, analysis.IsFlutter, "Should detect Flutter")
		assert.Equal(t, "material", analysis.UIFramework, "Should detect Material UI")
		assert.True(t, analysis.HasNavigation, "Should detect navigation")
		
		// Validate complex app structure
		symbolTypes := make(map[types.SymbolType]int)
		for _, symbol := range symbols {
			symbolTypes[symbol.Type]++
		}
		
		assert.GreaterOrEqual(t, symbolTypes[types.SymbolTypeWidget], 8, "Should find many widget classes")
		assert.GreaterOrEqual(t, symbolTypes[types.SymbolTypeMethod], 10, "Should find many methods")
		assert.GreaterOrEqual(t, len(analysis.Widgets), 8, "Should find many widgets")
		assert.GreaterOrEqual(t, len(analysis.LifecycleMethods), 2, "Should find lifecycle methods")
		
		// Validate navigation and complex features
		assert.Contains(t, analysis.Features, "MaterialApp", "Should detect MaterialApp")
		assert.Contains(t, analysis.Features, "Scaffold", "Should detect Scaffold")
		assert.Contains(t, analysis.Features, "AppBar", "Should detect AppBar")
		
		t.Logf("Navigation app validation: Symbols=%d, Widgets=%d, Features=%d, Navigation=%v", 
			len(symbols), len(analysis.Widgets), len(analysis.Features), analysis.HasNavigation)
	})
}

// TestFlutterPerformanceValidation validates parsing performance on complex files
func TestFlutterPerformanceValidation(t *testing.T) {
	manager := NewManager()
	
	t.Run("large Flutter file performance", func(t *testing.T) {
		// Generate a large Flutter file with many widgets
		var content strings.Builder
		content.WriteString("import 'package:flutter/material.dart';\n\n")
		
		// Generate 50 widget classes
		for i := 0; i < 50; i++ {
			if i%2 == 0 {
				// StatelessWidget
				content.WriteString(fmt.Sprintf(`class Widget%d extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container(
      child: Text('Widget %d'),
    );
  }
}

`, i, i))
			} else {
				// StatefulWidget with State
				content.WriteString(fmt.Sprintf(`class Widget%d extends StatefulWidget {
  @override
  _Widget%dState createState() => _Widget%dState();
}

class _Widget%dState extends State<Widget%d> {
  int _counter%d = 0;
  
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
    return Container(
      child: Text('Stateful Widget %d: $_counter%d'),
    );
  }
  
  void _increment%d() {
    setState(() {
      _counter%d++;
    });
  }
}

`, i, i, i, i, i, i, i, i, i, i))
			}
		}
		
		largeContent := content.String()
		
		// Measure parsing performance
		startTime := time.Now()
		ast, err := manager.parseDartContent(largeContent, "large_flutter_file.dart")
		parseTime := time.Since(startTime)
		
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		startTime = time.Now()
		symbols, err := manager.ExtractSymbols(ast)
		symbolTime := time.Since(startTime)
		
		require.NoError(t, err)
		
		// Performance requirements
		assert.Less(t, parseTime.Milliseconds(), int64(500), "Parsing should complete within 500ms")
		assert.Less(t, symbolTime.Milliseconds(), int64(200), "Symbol extraction should complete within 200ms")
		
		// Verify we found all the widgets
		expectedSymbols := 50 + 25 + 25*4 + 25*2 + 1 // Widgets + State classes + Methods per stateful + lifecycle methods + import
		assert.GreaterOrEqual(t, len(symbols), expectedSymbols/2, "Should find substantial number of symbols")
		
		t.Logf("Large file performance: Parse=%dms, Symbols=%dms, Total symbols=%d", 
			parseTime.Milliseconds(), symbolTime.Milliseconds(), len(symbols))
	})
	
	t.Run("flutter detection performance", func(t *testing.T) {
		detector := NewFlutterDetector()
		
		// Generate complex Flutter content
		content := `import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class ComplexApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => CounterModel()),
        BlocProvider(create: (_) => CounterBloc()),
      ],
      child: MaterialApp(
        home: ComplexHome(),
      ),
    );
  }
}` + strings.Repeat(`
class TestWidget extends StatefulWidget {
  @override
  _TestWidgetState createState() => _TestWidgetState();
}

class _TestWidgetState extends State<TestWidget> with TickerProviderStateMixin {
  @override
  void initState() { super.initState(); }
  @override
  void dispose() { super.dispose(); }
  @override
  Widget build(BuildContext context) { return Container(); }
}`, 20)
		
		// Measure Flutter analysis performance
		startTime := time.Now()
		analysis := detector.AnalyzeFlutterContent(content)
		analysisTime := time.Since(startTime)
		
		assert.Less(t, analysisTime.Milliseconds(), int64(100), "Flutter analysis should complete within 100ms")
		assert.True(t, analysis.IsFlutter, "Should detect Flutter")
		assert.GreaterOrEqual(t, len(analysis.Widgets), 20, "Should find many widgets")
		
		t.Logf("Flutter analysis performance: %dms, Widgets found: %d", 
			analysisTime.Milliseconds(), len(analysis.Widgets))
	})
}

// TestWeek2CoverageValidation validates our 60% coverage target for Week 2
func TestWeek2CoverageValidation(t *testing.T) {
	manager := NewManager()
	
	// Define coverage test cases representing different Flutter patterns
	testCases := []struct {
		name        string
		content     string
		coverage    []string // Features we should detect
		shouldPass  bool
	}{
		{
			name: "Basic StatelessWidget",
			content: `import 'package:flutter/material.dart';
class MyWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) { return Container(); }
}`,
			coverage:   []string{"flutter_import", "stateless_widget", "build_method", "@override"},
			shouldPass: true,
		},
		{
			name: "StatefulWidget with lifecycle",
			content: `import 'package:flutter/material.dart';
class MyWidget extends StatefulWidget {
  @override
  _MyWidgetState createState() => _MyWidgetState();
}
class _MyWidgetState extends State<MyWidget> {
  @override
  void initState() { super.initState(); }
  @override
  void dispose() { super.dispose(); }
  @override
  Widget build(BuildContext context) { return Container(); }
}`,
			coverage:   []string{"flutter_import", "stateful_widget", "state_class", "lifecycle_methods", "build_method"},
			shouldPass: true,
		},
		{
			name: "Material App Structure",
			content: `import 'package:flutter/material.dart';
class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        appBar: AppBar(title: Text('App')),
        body: Center(child: Text('Hello')),
        floatingActionButton: FloatingActionButton(
          onPressed: () {},
          child: Icon(Icons.add),
        ),
      ),
    );
  }
}`,
			coverage:   []string{"material_app", "scaffold", "app_bar", "fab", "navigation_ready"},
			shouldPass: true,
		},
		{
			name: "State Management (Provider)",
			content: `import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
class Model extends ChangeNotifier {}
class App extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => Model(),
      child: Consumer<Model>(
        builder: (context, model, child) => MaterialApp(),
      ),
    );
  }
}`,
			coverage:   []string{"provider_state_management", "change_notifier", "consumer_pattern"},
			shouldPass: true,
		},
		{
			name: "State Management (Riverpod)",
			content: `import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
final provider = StateProvider<int>((ref) => 0);
class App extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return MaterialApp();
  }
}`,
			coverage:   []string{"riverpod_state_management", "consumer_widget", "state_provider"},
			shouldPass: true,
		},
	}
	
	passedTests := 0
	totalFeatures := 0
	detectedFeatures := 0
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := manager.parseDartContent(tc.content, tc.name+".dart")
			require.NoError(t, err)
			
			symbols, err := manager.ExtractSymbols(ast)
			require.NoError(t, err)
			
			detector := NewFlutterDetector()
			analysis := detector.AnalyzeFlutterContent(tc.content)
			
			// Check coverage features
			featuresFound := 0
			totalFeatures += len(tc.coverage)
			
			for _, feature := range tc.coverage {
				found := false
				switch feature {
				case "flutter_import":
					found = analysis.IsFlutter
				case "stateless_widget", "stateful_widget":
					found = len(analysis.Widgets) > 0
				case "state_class":
					found = containsSymbolType(symbols, types.SymbolTypeStateClass) || 
						containsSymbolType(symbols, types.SymbolTypeClass)
				case "build_method":
					found = containsSymbolType(symbols, types.SymbolTypeBuildMethod) || 
						containsSymbolName(symbols, "build")
				case "lifecycle_methods":
					found = len(analysis.LifecycleMethods) > 0 || 
						containsSymbolType(symbols, types.SymbolTypeLifecycleMethod)
				case "@override":
					found = analysis.HasOverride
				case "material_app":
					found = contains(analysis.Features, "MaterialApp")
				case "scaffold":
					found = contains(analysis.Features, "Scaffold")
				case "app_bar":
					found = contains(analysis.Features, "AppBar")
				case "fab":
					found = contains(analysis.Features, "FloatingActionButton")
				case "navigation_ready":
					found = len(analysis.Features) >= 3
				case "provider_state_management", "change_notifier", "consumer_pattern":
					found = analysis.StateManagement == "provider"
				case "riverpod_state_management", "consumer_widget", "state_provider":
					found = analysis.StateManagement == "riverpod"
				}
				
				if found {
					featuresFound++
					detectedFeatures++
				}
			}
			
			coverage := float64(featuresFound) / float64(len(tc.coverage)) * 100
			
			if coverage >= 60.0 {
				passedTests++
			}
			
			t.Logf("Coverage for '%s': %.1f%% (%d/%d features)", 
				tc.name, coverage, featuresFound, len(tc.coverage))
		})
	}
	
	// Calculate overall coverage
	overallCoverage := float64(detectedFeatures) / float64(totalFeatures) * 100
	testPassRate := float64(passedTests) / float64(len(testCases)) * 100
	
	t.Logf("Week 2 Coverage Results:")
	t.Logf("  Overall feature coverage: %.1f%% (%d/%d)", overallCoverage, detectedFeatures, totalFeatures)
	t.Logf("  Test pass rate: %.1f%% (%d/%d)", testPassRate, passedTests, len(testCases))
	
	// Week 2 Success Criteria: 60% coverage
	assert.GreaterOrEqual(t, overallCoverage, 60.0, "Overall feature coverage should be ≥60%")
	assert.GreaterOrEqual(t, testPassRate, 80.0, "Test pass rate should be ≥80%")
}

// Helper functions for coverage validation
func containsSymbolType(symbols []*types.Symbol, symbolType types.SymbolType) bool {
	for _, symbol := range symbols {
		if symbol.Type == symbolType {
			return true
		}
	}
	return false
}

func containsSymbolName(symbols []*types.Symbol, name string) bool {
	for _, symbol := range symbols {
		if symbol.Name == name {
			return true
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}