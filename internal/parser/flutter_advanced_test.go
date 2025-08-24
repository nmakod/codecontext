package parser

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComplexWidgetHierarchies tests detection of complex Flutter widget patterns
func TestComplexWidgetHierarchies(t *testing.T) {
	manager := NewManager()
	
	t.Run("nested widget composition", func(t *testing.T) {
		dartCode := `import 'package:flutter/material.dart';

class AppContainer extends StatelessWidget {
  final Widget child;
  
  const AppContainer({Key? key, required this.child}) : super(key: key);
  
  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.all(16.0),
      child: child,
    );
  }
}

class MainScreen extends StatefulWidget {
  @override
  _MainScreenState createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen> 
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
    return AppContainer(
      child: Scaffold(
        appBar: AppBar(
          title: Text('Complex App'),
        ),
        body: CustomWidget(),
      ),
    );
  }
}

class CustomWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        HeaderWidget(),
        ContentWidget(),
        FooterWidget(),
      ],
    );
  }
}

class HeaderWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container(
      height: 100,
      child: Text('Header'),
    );
  }
}

class ContentWidget extends StatefulWidget {
  @override
  _ContentWidgetState createState() => _ContentWidgetState();
}

class _ContentWidgetState extends State<ContentWidget> {
  String content = 'Initial content';
  
  void updateContent(String newContent) {
    setState(() {
      content = newContent;
    });
  }
  
  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Text(content),
    );
  }
}

class FooterWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container(
      height: 50,
      child: Text('Footer'),
    );
  }
}`

		ast, err := manager.parseDartContent(dartCode, "complex_app.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should detect multiple widget classes
		widgetCount := 0
		buildMethodCount := 0
		stateClassCount := 0
		lifecycleMethodCount := 0
		
		for _, symbol := range symbols {
			switch symbol.Type {
			case "widget", "class": // widget is preferred but class is acceptable
				if symbol.Name == "AppContainer" || symbol.Name == "MainScreen" || 
				   symbol.Name == "CustomWidget" || symbol.Name == "HeaderWidget" ||
				   symbol.Name == "ContentWidget" || symbol.Name == "FooterWidget" {
					widgetCount++
				}
				if symbol.Name == "_MainScreenState" || symbol.Name == "_ContentWidgetState" {
					stateClassCount++
				}
			case "build_method": // build_method is preferred
				if symbol.Name == "build" {
					buildMethodCount++
				}
			case "method": // regular methods including build and lifecycle
				if symbol.Name == "build" {
					buildMethodCount++
				}
				if symbol.Name == "initState" || symbol.Name == "dispose" {
					lifecycleMethodCount++
				}
			}
		}
		
		assert.GreaterOrEqual(t, widgetCount, 6, "Should find 6 widget classes")
		assert.GreaterOrEqual(t, buildMethodCount, 6, "Should find 6 build methods")
		assert.GreaterOrEqual(t, stateClassCount, 2, "Should find 2 state classes")
		assert.GreaterOrEqual(t, lifecycleMethodCount, 2, "Should find lifecycle methods")
		
		// Check Flutter analysis metadata
		hasFlutter, _ := ast.Root.Metadata["has_flutter"].(bool)
		assert.True(t, hasFlutter, "Should detect Flutter")
		
		flutterFramework, _ := ast.Root.Metadata["flutter_framework"].(string)
		assert.Equal(t, "material", flutterFramework, "Should detect Material framework")
	})
	
	t.Run("mixin usage detection", func(t *testing.T) {
		dartCode := `import 'package:flutter/material.dart';

class AnimatedWidget extends StatefulWidget {
  @override
  _AnimatedWidgetState createState() => _AnimatedWidgetState();
}

class _AnimatedWidgetState extends State<AnimatedWidget> 
    with SingleTickerProviderStateMixin {
  AnimationController? _controller;
  Animation<double>? _animation;
  
  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: Duration(seconds: 1),
      vsync: this,
    );
    _animation = Tween<double>(begin: 0, end: 1).animate(_controller!);
  }
  
  @override
  void dispose() {
    _controller?.dispose();
    super.dispose();
  }
  
  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _animation!,
      builder: (context, child) {
        return Opacity(
          opacity: _animation!.value,
          child: Container(
            width: 100,
            height: 100,
            color: Colors.blue,
          ),
        );
      },
    );
  }
}`

		ast, err := manager.parseDartContent(dartCode, "animated_widget.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should detect the widget and state class with mixin
		var foundAnimatedWidget, foundStateClass bool
		for _, symbol := range symbols {
			if symbol.Name == "AnimatedWidget" && (symbol.Type == "widget" || symbol.Type == "class") {
				foundAnimatedWidget = true
			}
			if symbol.Name == "_AnimatedWidgetState" && symbol.Type == "class" {
				foundStateClass = true
			}
		}
		
		assert.True(t, foundAnimatedWidget, "Should find AnimatedWidget")
		assert.True(t, foundStateClass, "Should find state class with mixin")
	})
	
	t.Run("custom widget inheritance", func(t *testing.T) {
		dartCode := `import 'package:flutter/material.dart';

abstract class BaseWidget extends StatelessWidget {
  const BaseWidget({Key? key}) : super(key: key);
  
  // Template method pattern
  Widget buildContent(BuildContext context);
  
  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.all(8.0),
      child: buildContent(context),
    );
  }
}

class HeaderWidget extends BaseWidget {
  final String title;
  
  const HeaderWidget({Key? key, required this.title}) : super(key: key);
  
  @override
  Widget buildContent(BuildContext context) {
    return Text(
      title,
      style: Theme.of(context).textTheme.headline6,
    );
  }
}

class ButtonWidget extends BaseWidget {
  final String text;
  final VoidCallback? onPressed;
  
  const ButtonWidget({
    Key? key,
    required this.text,
    this.onPressed,
  }) : super(key: key);
  
  @override
  Widget buildContent(BuildContext context) {
    return ElevatedButton(
      onPressed: onPressed,
      child: Text(text),
    );
  }
}`

		ast, err := manager.parseDartContent(dartCode, "custom_inheritance.dart")
		require.NoError(t, err)
		require.NotNil(t, ast)
		
		symbols, err := manager.ExtractSymbols(ast)
		require.NoError(t, err)
		
		// Should detect inheritance hierarchy
		var foundBaseWidget, foundHeaderWidget, foundButtonWidget bool
		buildMethods := 0
		
		for _, symbol := range symbols {
			switch symbol.Name {
			case "BaseWidget":
				if symbol.Type == "class" || symbol.Type == "widget" {
					foundBaseWidget = true
				}
			case "HeaderWidget":
				if symbol.Type == "class" || symbol.Type == "widget" {
					foundHeaderWidget = true
				}
			case "ButtonWidget":
				if symbol.Type == "class" || symbol.Type == "widget" {
					foundButtonWidget = true
				}
			case "build", "buildContent":
				if symbol.Type == "method" || symbol.Type == "build_method" {
					buildMethods++
				}
			}
		}
		
		assert.True(t, foundBaseWidget, "Should find BaseWidget")
		assert.True(t, foundHeaderWidget, "Should find HeaderWidget")
		assert.True(t, foundButtonWidget, "Should find ButtonWidget")
		assert.GreaterOrEqual(t, buildMethods, 3, "Should find multiple build methods")
	})
}

// TestAdvancedFlutterPatterns tests more advanced Flutter patterns
func TestAdvancedFlutterPatterns(t *testing.T) {
	detector := NewFlutterDetector()
	
	t.Run("widget composition with providers", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

class UserModel extends ChangeNotifier {
  String _name = '';
  String get name => _name;
  
  void updateName(String newName) {
    _name = newName;
    notifyListeners();
  }
}

class UserApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (context) => UserModel(),
      child: MaterialApp(
        home: UserScreen(),
      ),
    );
  }
}

class UserScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('User Profile')),
      body: Column(
        children: [
          Consumer<UserModel>(
            builder: (context, user, child) {
              return Text('Name: ${user.name}');
            },
          ),
          UserInput(),
        ],
      ),
    );
  }
}

class UserInput extends StatefulWidget {
  @override
  _UserInputState createState() => _UserInputState();
}

class _UserInputState extends State<UserInput> {
  final TextEditingController _controller = TextEditingController();
  
  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }
  
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.all(16.0),
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
}`

		analysis := detector.AnalyzeFlutterContent(content)
		
		assert.True(t, analysis.IsFlutter, "Should detect Flutter")
		assert.Equal(t, "material", analysis.UIFramework, "Should detect Material UI")
		assert.Equal(t, "provider", analysis.StateManagement, "Should detect Provider")
		
		// Should find multiple widgets
		assert.GreaterOrEqual(t, len(analysis.Widgets), 4, "Should find multiple widgets")
		
		// Should detect lifecycle methods
		assert.Contains(t, analysis.LifecycleMethods, "dispose", "Should find dispose method")
		
		// Should detect common widgets
		assert.Contains(t, analysis.Features, "MaterialApp", "Should find MaterialApp")
		assert.Contains(t, analysis.Features, "Scaffold", "Should find Scaffold")
		assert.Contains(t, analysis.Features, "AppBar", "Should find AppBar")
	})
	
	t.Run("complex build method variations", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';

class ComplexWidget extends StatefulWidget {
  @override
  _ComplexWidgetState createState() => _ComplexWidgetState();
}

class _ComplexWidgetState extends State<ComplexWidget> {
  bool _isExpanded = false;
  
  // Standard build method
  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        _buildHeader(),
        if (_isExpanded) _buildContent(),
        _buildFooter(),
      ],
    );
  }
  
  // Private build helper methods
  Widget _buildHeader() {
    return GestureDetector(
      onTap: () => setState(() => _isExpanded = !_isExpanded),
      child: Container(
        padding: EdgeInsets.all(16),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text('Header'),
            Icon(_isExpanded ? Icons.expand_less : Icons.expand_more),
          ],
        ),
      ),
    );
  }
  
  Widget _buildContent() {
    return AnimatedContainer(
      duration: Duration(milliseconds: 300),
      height: _isExpanded ? 200 : 0,
      child: ListView.builder(
        itemCount: 10,
        itemBuilder: (context, index) => ListTile(
          title: Text('Item $index'),
        ),
      ),
    );
  }
  
  Widget _buildFooter() {
    return Container(
      padding: EdgeInsets.all(8),
      child: Text('Footer'),
    );
  }
}`

		analysis := detector.AnalyzeFlutterContent(content)
		
		assert.True(t, analysis.IsFlutter, "Should detect Flutter")
		assert.Equal(t, "material", analysis.UIFramework, "Should detect Material UI")
		
		// Should find the main widget
		assert.GreaterOrEqual(t, len(analysis.Widgets), 1, "Should find at least one widget")
		
		var foundComplexWidget bool
		for _, widget := range analysis.Widgets {
			if widget.Name == "ComplexWidget" {
				foundComplexWidget = true
				assert.Equal(t, "stateful", widget.Type, "Should be stateful widget")
			}
		}
		assert.True(t, foundComplexWidget, "Should find ComplexWidget")
	})
}