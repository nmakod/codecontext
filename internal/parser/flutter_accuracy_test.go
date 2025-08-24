package parser

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFlutterWidgetDetectionAccuracy tests accuracy against a variety of Flutter patterns
func TestFlutterWidgetDetectionAccuracy(t *testing.T) {
	detector := NewFlutterDetector()
	manager := NewManager()
	
	testCases := []struct {
		name           string
		content        string
		expectedWidgets int
		expectedType   string
		shouldDetect   bool
	}{
		{
			name: "basic StatelessWidget",
			content: `import 'package:flutter/material.dart';

class MyWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container();
  }
}`,
			expectedWidgets: 1,
			expectedType: "stateless",
			shouldDetect: true,
		},
		{
			name: "basic StatefulWidget with State",
			content: `import 'package:flutter/material.dart';

class MyWidget extends StatefulWidget {
  @override
  _MyWidgetState createState() => _MyWidgetState();
}

class _MyWidgetState extends State<MyWidget> {
  @override
  Widget build(BuildContext context) {
    return Container();
  }
}`,
			expectedWidgets: 2, // StatefulWidget + State class
			expectedType: "stateful",
			shouldDetect: true,
		},
		{
			name: "ConsumerWidget (Riverpod)",
			content: `import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class MyWidget extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Container();
  }
}`,
			expectedWidgets: 1,
			expectedType: "consumer",
			shouldDetect: true,
		},
		{
			name: "multiple widgets in one file",
			content: `import 'package:flutter/material.dart';

class HeaderWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return AppBar(title: Text('Header'));
  }
}

class BodyWidget extends StatefulWidget {
  @override
  _BodyWidgetState createState() => _BodyWidgetState();
}

class _BodyWidgetState extends State<BodyWidget> {
  @override
  Widget build(BuildContext context) {
    return Container();
  }
}

class FooterWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Text('Footer');
  }
}`,
			expectedWidgets: 4, // 2 StatelessWidget + 1 StatefulWidget + 1 State class
			expectedType: "mixed",
			shouldDetect: true,
		},
		{
			name: "widget with mixin",
			content: `import 'package:flutter/material.dart';

class AnimatedWidget extends StatefulWidget {
  @override
  _AnimatedWidgetState createState() => _AnimatedWidgetState();
}

class _AnimatedWidgetState extends State<AnimatedWidget> 
    with TickerProviderStateMixin {
  @override
  Widget build(BuildContext context) {
    return Container();
  }
}`,
			expectedWidgets: 2,
			expectedType: "stateful",
			shouldDetect: true,
		},
		{
			name: "plain Dart class (no Flutter)",
			content: `class MyClass {
  String name = 'test';
  
  void doSomething() {
    print('Hello');
  }
}`,
			expectedWidgets: 0,
			expectedType: "none",
			shouldDetect: false,
		},
		{
			name: "Flutter import but no widgets",
			content: `import 'package:flutter/material.dart';

void main() {
  runApp(Container());
}

class DataModel {
  String name = '';
}`,
			expectedWidgets: 0,
			expectedType: "none",
			shouldDetect: true, // Has Flutter import
		},
		{
			name: "complex real-world example",
			content: `import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

class UserModel extends ChangeNotifier {
  String _name = '';
  String get name => _name;
  
  void updateName(String newName) {
    _name = newName;
    notifyListeners();
  }
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (context) => UserModel(),
      child: MaterialApp(
        title: 'Flutter Demo',
        theme: ThemeData(primarySwatch: Colors.blue),
        home: HomeScreen(),
      ),
    );
  }
}

class HomeScreen extends StatefulWidget {
  @override
  _HomeScreenState createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> 
    with AutomaticKeepAliveClientMixin {
  @override
  bool get wantKeepAlive => true;
  
  @override
  Widget build(BuildContext context) {
    super.build(context);
    return Scaffold(
      appBar: AppBar(
        title: Text('Home'),
        actions: [
          Consumer<UserModel>(
            builder: (context, user, child) {
              return IconButton(
                icon: Icon(Icons.person),
                onPressed: () {},
              );
            },
          ),
        ],
      ),
      body: UserProfile(),
      floatingActionButton: FloatingActionButton(
        onPressed: () {},
        child: Icon(Icons.add),
      ),
    );
  }
}

class UserProfile extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Consumer<UserModel>(
      builder: (context, user, child) {
        return Column(
          children: [
            Text('Name: ${user.name}'),
            UserForm(),
          ],
        );
      },
    );
  }
}

class UserForm extends StatefulWidget {
  @override
  _UserFormState createState() => _UserFormState();
}

class _UserFormState extends State<UserForm> {
  final _controller = TextEditingController();
  
  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }
  
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.all(16),
      child: Row(
        children: [
          Expanded(
            child: TextField(
              controller: _controller,
              decoration: InputDecoration(hintText: 'Enter name'),
            ),
          ),
          ElevatedButton(
            onPressed: () {
              context.read<UserModel>().updateName(_controller.text);
              _controller.clear();
            },
            child: Text('Update'),
          ),
        ],
      ),
    );
  }
}`,
			expectedWidgets: 6, // MyApp, HomeScreen, _HomeScreenState, UserProfile, UserForm, _UserFormState
			expectedType: "mixed",
			shouldDetect: true,
		},
	}
	
	correctDetections := 0
	totalTests := len(testCases)
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test Flutter detection
			analysis := detector.AnalyzeFlutterContent(tc.content)
			
			detectionCorrect := analysis.IsFlutter == tc.shouldDetect
			if detectionCorrect {
				correctDetections++
			}
			
			assert.Equal(t, tc.shouldDetect, analysis.IsFlutter, 
				"Flutter detection should be %v", tc.shouldDetect)
			
			if tc.shouldDetect {
				// Test widget count accuracy
				if tc.expectedWidgets > 0 {
					assert.GreaterOrEqual(t, len(analysis.Widgets), tc.expectedWidgets, 
						"Should find at least %d widgets", tc.expectedWidgets)
				}
				
				// Test specific widget types
				if tc.expectedType != "mixed" && tc.expectedType != "none" && len(analysis.Widgets) > 0 {
					found := false
					for _, widget := range analysis.Widgets {
						if widget.Type == tc.expectedType {
							found = true
							break
						}
					}
					assert.True(t, found, "Should find widget of type %s", tc.expectedType)
				}
			}
			
			// Test with full parsing pipeline
			ast, err := manager.parseDartContent(tc.content, tc.name+".dart")
			require.NoError(t, err)
			require.NotNil(t, ast)
			
			symbols, err := manager.ExtractSymbols(ast)
			require.NoError(t, err)
			
			// Verify Flutter metadata is correctly set
			if tc.shouldDetect {
				hasFlutter, _ := ast.Root.Metadata["has_flutter"].(bool)
				assert.True(t, hasFlutter, "AST should have Flutter metadata")
			}
			
			t.Logf("Test '%s': Flutter=%v, Widgets=%d, Symbols=%d", 
				tc.name, analysis.IsFlutter, len(analysis.Widgets), len(symbols))
		})
	}
	
	// Calculate accuracy percentage
	accuracy := float64(correctDetections) / float64(totalTests) * 100
	t.Logf("Flutter detection accuracy: %.1f%% (%d/%d correct)", accuracy, correctDetections, totalTests)
	
	// Week 2 requirement: >90% accuracy
	assert.GreaterOrEqual(t, accuracy, 90.0, "Flutter widget detection accuracy should be ≥90%")
}

// TestBuildMethodDetectionAccuracy tests build method identification accuracy
func TestBuildMethodDetectionAccuracy(t *testing.T) {
	manager := NewManager()
	
	testCases := []struct {
		name          string
		content       string
		expectedBuild int
	}{
		{
			name: "single build method",
			content: `import 'package:flutter/material.dart';

class MyWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container();
  }
}`,
			expectedBuild: 1,
		},
		{
			name: "multiple build methods",
			content: `import 'package:flutter/material.dart';

class Widget1 extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container();
  }
}

class Widget2 extends StatefulWidget {
  @override
  _Widget2State createState() => _Widget2State();
}

class _Widget2State extends State<Widget2> {
  @override
  Widget build(BuildContext context) {
    return Text('Hello');
  }
}`,
			expectedBuild: 2,
		},
		{
			name: "build method with helpers",
			content: `import 'package:flutter/material.dart';

class ComplexWidget extends StatefulWidget {
  @override
  _ComplexWidgetState createState() => _ComplexWidgetState();
}

class _ComplexWidgetState extends State<ComplexWidget> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: _buildAppBar(),
      body: _buildBody(),
    );
  }
  
  Widget _buildAppBar() {
    return AppBar(title: Text('Title'));
  }
  
  Widget _buildBody() {
    return Container();
  }
}`,
			expectedBuild: 1, // Only the main build method, helpers are separate
		},
	}
	
	correctBuildDetections := 0
	totalBuildTests := len(testCases)
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := manager.parseDartContent(tc.content, tc.name+".dart")
			require.NoError(t, err)
			
			symbols, err := manager.ExtractSymbols(ast)
			require.NoError(t, err)
			
			buildMethodCount := 0
			for _, symbol := range symbols {
				if symbol.Type == "build_method" || 
					(symbol.Type == "method" && symbol.Name == "build") {
					buildMethodCount++
				}
			}
			
			if buildMethodCount >= tc.expectedBuild {
				correctBuildDetections++
			}
			
			assert.GreaterOrEqual(t, buildMethodCount, tc.expectedBuild, 
				"Should find at least %d build methods", tc.expectedBuild)
			
			t.Logf("Test '%s': Expected=%d, Found=%d build methods", 
				tc.name, tc.expectedBuild, buildMethodCount)
		})
	}
	
	// Calculate build method detection accuracy
	buildAccuracy := float64(correctBuildDetections) / float64(totalBuildTests) * 100
	t.Logf("Build method detection accuracy: %.1f%% (%d/%d correct)", 
		buildAccuracy, correctBuildDetections, totalBuildTests)
	
	// Week 2 requirement: >95% accuracy for build method identification
	assert.GreaterOrEqual(t, buildAccuracy, 95.0, "Build method identification accuracy should be ≥95%")
}