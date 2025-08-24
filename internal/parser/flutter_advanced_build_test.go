package parser

import (
	"fmt"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

// TestAdvancedBuildMethodDetection tests sophisticated build method patterns
func TestAdvancedBuildMethodDetection(t *testing.T) {
	detector := NewFlutterDetector()
	
	t.Run("@override annotation detection", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';

class MyWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container();
  }
}

class AnotherWidget extends StatefulWidget {
  @override
  _AnotherWidgetState createState() => _AnotherWidgetState();
}

class _AnotherWidgetState extends State<AnotherWidget> {
  @override
  void initState() {
    super.initState();
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold();
  }
  
  @override
  void dispose() {
    super.dispose();
  }
}`

		analysis := detector.AnalyzeFlutterContent(content)
		
		assert.True(t, analysis.IsFlutter, "Should detect Flutter")
		assert.True(t, analysis.HasOverride, "Should detect @override annotation")
		assert.Contains(t, analysis.LifecycleMethods, "initState", "Should find initState")
		assert.Contains(t, analysis.LifecycleMethods, "dispose", "Should find dispose")
	})
	
	t.Run("build helper methods detection", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';

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
      bottomNavigationBar: _buildBottomNav(),
    );
  }
  
  Widget _buildAppBar() {
    return AppBar(
      title: Text('Complex App'),
      actions: [
        _buildMenuButton(),
      ],
    );
  }
  
  Widget _buildBody() {
    return Column(
      children: [
        _buildHeader(),
        Expanded(child: _buildContent()),
        _buildFooter(),
      ],
    );
  }
  
  Widget _buildBottomNav() {
    return BottomNavigationBar(
      items: [
        BottomNavigationBarItem(icon: Icon(Icons.home), label: 'Home'),
        BottomNavigationBarItem(icon: Icon(Icons.search), label: 'Search'),
      ],
    );
  }
  
  Widget _buildMenuButton() {
    return IconButton(
      icon: Icon(Icons.menu),
      onPressed: () {},
    );
  }
  
  Widget _buildHeader() {
    return Container(height: 100, child: Text('Header'));
  }
  
  Widget _buildContent() {
    return ListView.builder(
      itemCount: 20,
      itemBuilder: (context, index) => ListTile(title: Text('Item $index')),
    );
  }
  
  Widget _buildFooter() {
    return Container(height: 50, child: Text('Footer'));
  }
}`

		analysis := detector.AnalyzeFlutterContent(content)
		
		assert.True(t, analysis.IsFlutter, "Should detect Flutter")
		assert.True(t, analysis.HasOverride, "Should detect @override annotation")
		
		// Should find multiple build helper methods
		assert.GreaterOrEqual(t, len(analysis.BuildHelpers), 5, "Should find multiple build helpers")
		
		expectedHelpers := []string{"_buildAppBar", "_buildBody", "_buildBottomNav", "_buildMenuButton", "_buildHeader", "_buildContent", "_buildFooter"}
		for _, helper := range expectedHelpers {
			assert.Contains(t, analysis.BuildHelpers, helper, fmt.Sprintf("Should find %s helper", helper))
		}
		
		// Should detect good composition depth
		assert.GreaterOrEqual(t, analysis.CompositionDepth, 1, "Should detect composition depth")
	})
	
	t.Run("widget composition patterns", func(t *testing.T) {
		content := `import 'package:flutter/material.dart';

class AppShell extends StatelessWidget {
  final Widget child;
  const AppShell({Key? key, required this.child}) : super(key: key);
  
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        appBar: AppBar(title: Text('App Shell')),
        body: child,
      ),
    );
  }
}

class PageWrapper extends StatelessWidget {
  final Widget content;
  final String title;
  
  const PageWrapper({Key? key, required this.content, required this.title}) : super(key: key);
  
  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(title, style: Theme.of(context).textTheme.headline6),
          SizedBox(height: 16),
          Expanded(child: content),
        ],
      ),
    );
  }
}

class ContentCard extends StatelessWidget {
  final String title;
  final String subtitle;
  final Widget? action;
  
  const ContentCard({
    Key? key,
    required this.title,
    required this.subtitle,
    this.action,
  }) : super(key: key);
  
  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(title, style: Theme.of(context).textTheme.subtitle1),
            SizedBox(height: 8),
            Text(subtitle),
            if (action != null) ...[
              SizedBox(height: 16),
              action!,
            ],
          ],
        ),
      ),
    );
  }
}

class MainPage extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return AppShell(
      child: PageWrapper(
        title: 'Main Page',
        content: ListView(
          children: [
            ContentCard(
              title: 'Card 1',
              subtitle: 'This is the first card',
              action: ElevatedButton(
                onPressed: () {},
                child: Text('Action'),
              ),
            ),
            ContentCard(
              title: 'Card 2',
              subtitle: 'This is the second card',
            ),
            ContentCard(
              title: 'Card 3',
              subtitle: 'This is the third card',
              action: TextButton(
                onPressed: () {},
                child: Text('Secondary Action'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}`

		analysis := detector.AnalyzeFlutterContent(content)
		
		assert.True(t, analysis.IsFlutter, "Should detect Flutter")
		assert.True(t, analysis.HasOverride, "Should detect @override annotation")
		assert.Equal(t, "material", analysis.UIFramework, "Should detect Material framework")
		
		// Should find multiple widgets showing composition
		assert.GreaterOrEqual(t, len(analysis.Widgets), 4, "Should find multiple widgets")
		
		// Should have good composition depth due to multiple widget classes
		assert.GreaterOrEqual(t, analysis.CompositionDepth, 1, "Should detect composition depth")
		
		// Should detect common Flutter widgets
		expectedFeatures := []string{"MaterialApp", "Scaffold", "AppBar"}
		for _, feature := range expectedFeatures {
			assert.Contains(t, analysis.Features, feature, fmt.Sprintf("Should find %s", feature))
		}
	})
}