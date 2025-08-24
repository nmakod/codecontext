package parser

import (
	"regexp"
	"strings"
	
	"github.com/nuthan-ms/codecontext/pkg/types"
)

// FlutterDetector provides enhanced Flutter-specific detection capabilities
type FlutterDetector struct {
	patterns           map[string]*regexp.Regexp
	stateManagementPatterns map[string]*regexp.Regexp
}

// NewFlutterDetector creates a new Flutter detector with comprehensive patterns
func NewFlutterDetector() *FlutterDetector {
	return &FlutterDetector{
		patterns: map[string]*regexp.Regexp{
			// Core Flutter patterns
			"flutter_import":     regexp.MustCompile(`import\s+['"]package:flutter/`),
			"material_import":    regexp.MustCompile(`import\s+['"]package:flutter/material\.dart['"]`),
			"cupertino_import":   regexp.MustCompile(`import\s+['"]package:flutter/cupertino\.dart['"]`),
			"widgets_import":     regexp.MustCompile(`import\s+['"]package:flutter/widgets\.dart['"]`),
			
			// Widget patterns
			"stateless_widget":   regexp.MustCompile(`class\s+\w+\s+extends\s+StatelessWidget`),
			"stateful_widget":    regexp.MustCompile(`class\s+\w+\s+extends\s+StatefulWidget`),
			"inherited_widget":   regexp.MustCompile(`class\s+\w+\s+extends\s+InheritedWidget`),
			"consumer_widget":    regexp.MustCompile(`class\s+\w+\s+extends\s+ConsumerWidget`),
			"hook_widget":        regexp.MustCompile(`class\s+\w+\s+extends\s+HookWidget`),
			"state_class":        regexp.MustCompile(`class\s+\w+\s+extends\s+State<`),
			
			// Build method patterns
			"build_method":       regexp.MustCompile(`@override\s+Widget\s+build\s*\(\s*BuildContext\s+\w+\s*\)`),
			"build_method_simple": regexp.MustCompile(`Widget\s+build\s*\(\s*BuildContext\s+\w+\s*\)`),
			"build_helper":       regexp.MustCompile(`Widget\s+(_\w+)\s*\([^)]*\)\s*\{`),
			
			// Override annotation
			"override_annotation": regexp.MustCompile(`@override`),
			
			// Common Flutter widgets
			"scaffold":           regexp.MustCompile(`Scaffold\s*\(`),
			"material_app":       regexp.MustCompile(`MaterialApp\s*\(`),
			"cupertino_app":      regexp.MustCompile(`CupertinoApp\s*\(`),
			"app_bar":            regexp.MustCompile(`AppBar\s*\(`),
			"floating_action":    regexp.MustCompile(`FloatingActionButton\s*\(`),
			"container":          regexp.MustCompile(`Container\s*\(`),
			"column":             regexp.MustCompile(`Column\s*\(`),
			"row":                regexp.MustCompile(`Row\s*\(`),
			"text":               regexp.MustCompile(`Text\s*\(`),
			"elevated_button":    regexp.MustCompile(`ElevatedButton\s*\(`),
			"text_button":        regexp.MustCompile(`TextButton\s*\(`),
			
			// Lifecycle methods
			"init_state":         regexp.MustCompile(`@override\s+void\s+initState\s*\(`),
			"dispose":            regexp.MustCompile(`@override\s+void\s+dispose\s*\(`),
			"did_update_widget":  regexp.MustCompile(`@override\s+void\s+didUpdateWidget\s*\(`),
			
			// Navigation
			"navigator":          regexp.MustCompile(`Navigator\.\w+`),
			"named_route":        regexp.MustCompile(`pushNamed\s*\(`),
		},
		
		stateManagementPatterns: map[string]*regexp.Regexp{
			// Provider pattern
			"provider":           regexp.MustCompile(`import\s+['"]package:provider/`),
			"change_notifier":    regexp.MustCompile(`extends\s+ChangeNotifier`),
			"consumer":           regexp.MustCompile(`Consumer<`),
			"provider_widget":    regexp.MustCompile(`Provider<`),
			
			// Riverpod
			"riverpod":           regexp.MustCompile(`import\s+['"]package:flutter_riverpod/`),
			"state_provider":     regexp.MustCompile(`StateProvider<`),
			"future_provider":    regexp.MustCompile(`FutureProvider<`),
			"stream_provider":    regexp.MustCompile(`StreamProvider<`),
			"consumer_widget":    regexp.MustCompile(`ConsumerWidget`),
			
			// BLoC
			"bloc":               regexp.MustCompile(`import\s+['"]package:flutter_bloc/`),
			"bloc_class":         regexp.MustCompile(`extends\s+Bloc<`),
			"cubit_class":        regexp.MustCompile(`extends\s+Cubit<`),
			"bloc_builder":       regexp.MustCompile(`BlocBuilder<`),
			"bloc_consumer":      regexp.MustCompile(`BlocConsumer<`),
			
			// GetX
			"getx":               regexp.MustCompile(`import\s+['"]package:get/`),
			"getx_controller":    regexp.MustCompile(`extends\s+GetxController`),
			"obx":                regexp.MustCompile(`Obx\s*\(`),
		},
	}
}

// AnalyzeFlutterContent performs comprehensive Flutter analysis on Dart content
func (fd *FlutterDetector) AnalyzeFlutterContent(content string) *FlutterAnalysis {
	analysis := &FlutterAnalysis{
		IsFlutter:        false,
		Framework:        "none",
		Widgets:          make([]FlutterWidget, 0),
		StateManagement:  "none",
		Features:         make([]string, 0),
		UIFramework:      "none",
		HasNavigation:    false,
		LifecycleMethods: make([]string, 0),
		CompositionDepth: 0,
		BuildHelpers:     make([]string, 0),
		HasOverride:      false,
	}
	
	// Check for Flutter imports
	if fd.patterns["flutter_import"].MatchString(content) {
		analysis.IsFlutter = true
		analysis.Framework = "flutter"
		
		// Determine UI framework
		if fd.patterns["material_import"].MatchString(content) {
			analysis.UIFramework = "material"
		} else if fd.patterns["cupertino_import"].MatchString(content) {
			analysis.UIFramework = "cupertino"
		} else if fd.patterns["widgets_import"].MatchString(content) {
			analysis.UIFramework = "widgets"
		}
		
		// Analyze widgets
		analysis.Widgets = fd.findWidgets(content)
		
		// Analyze state management
		analysis.StateManagement = fd.detectStateManagement(content)
		
		// Check for navigation
		if fd.patterns["navigator"].MatchString(content) || fd.patterns["named_route"].MatchString(content) {
			analysis.HasNavigation = true
		}
		
		// Find lifecycle methods
		analysis.LifecycleMethods = fd.findLifecycleMethods(content)
		
		// Detect features
		analysis.Features = fd.detectFeatures(content)
		
		// Detect build helpers (private methods that return Widget)
		analysis.BuildHelpers = fd.findBuildHelpers(content)
		
		// Check for @override annotation usage
		analysis.HasOverride = fd.patterns["override_annotation"].MatchString(content)
		
		// Calculate composition depth (rough estimate based on widget count)
		analysis.CompositionDepth = len(analysis.Widgets)
		if analysis.CompositionDepth > 0 {
			// If we have widgets, assume at least depth 1
			analysis.CompositionDepth = max(1, analysis.CompositionDepth/3) // Rough heuristic
		}
	}
	
	return analysis
}

// findWidgets identifies Flutter widgets in the content
func (fd *FlutterDetector) findWidgets(content string) []FlutterWidget {
	widgets := make([]FlutterWidget, 0)
	
	// Find StatelessWidget classes
	if matches := fd.patterns["stateless_widget"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			className := fd.extractClassName(match[0])
			if className != "" {
				widgets = append(widgets, FlutterWidget{
					Name: className,
					Type: "stateless",
					HasBuildMethod: fd.hasBuildMethod(content, className),
				})
			}
		}
	}
	
	// Find ConsumerWidget classes (Riverpod)
	if matches := fd.patterns["consumer_widget"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			className := fd.extractClassName(match[0])
			if className != "" {
				widgets = append(widgets, FlutterWidget{
					Name: className,
					Type: "consumer",
					HasBuildMethod: fd.hasBuildMethod(content, className),
				})
			}
		}
	}
	
	// Find HookWidget classes (flutter_hooks)
	if matches := fd.patterns["hook_widget"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			className := fd.extractClassName(match[0])
			if className != "" {
				widgets = append(widgets, FlutterWidget{
					Name: className,
					Type: "hook",
					HasBuildMethod: fd.hasBuildMethod(content, className),
				})
			}
		}
	}
	
	// Find StatefulWidget classes
	if matches := fd.patterns["stateful_widget"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			className := fd.extractClassName(match[0])
			if className != "" {
				widgets = append(widgets, FlutterWidget{
					Name: className,
					Type: "stateful",
					HasBuildMethod: false, // StatefulWidget itself doesn't have build method
				})
			}
		}
	}
	
	// Find State classes
	if matches := fd.patterns["state_class"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			className := fd.extractClassName(match[0])
			if className != "" {
				widgets = append(widgets, FlutterWidget{
					Name: className,
					Type: "state",
					HasBuildMethod: fd.hasBuildMethod(content, className),
				})
			}
		}
	}
	
	return widgets
}

// extractClassName extracts class name from a class declaration match
func (fd *FlutterDetector) extractClassName(classDecl string) string {
	// Simple regex to extract class name
	classNamePattern := regexp.MustCompile(`class\s+(\w+)`)
	if matches := classNamePattern.FindStringSubmatch(classDecl); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// hasBuildMethod checks if a class has a build method
func (fd *FlutterDetector) hasBuildMethod(content, className string) bool {
	// Look for build method in the context of the class
	// This is a simplified check - a full implementation would need proper scope analysis
	return fd.patterns["build_method"].MatchString(content) || fd.patterns["build_method_simple"].MatchString(content)
}

// detectStateManagement identifies the state management approach used
func (fd *FlutterDetector) detectStateManagement(content string) string {
	// Check for Riverpod
	if fd.stateManagementPatterns["riverpod"].MatchString(content) {
		return "riverpod"
	}
	
	// Check for BLoC
	if fd.stateManagementPatterns["bloc"].MatchString(content) {
		return "bloc"
	}
	
	// Check for Provider
	if fd.stateManagementPatterns["provider"].MatchString(content) {
		return "provider"
	}
	
	// Check for GetX
	if fd.stateManagementPatterns["getx"].MatchString(content) {
		return "getx"
	}
	
	// Check for built-in setState
	if strings.Contains(content, "setState") {
		return "setState"
	}
	
	return "none"
}

// findLifecycleMethods identifies lifecycle methods in the content
func (fd *FlutterDetector) findLifecycleMethods(content string) []string {
	methods := make([]string, 0)
	
	if fd.patterns["init_state"].MatchString(content) {
		methods = append(methods, "initState")
	}
	if fd.patterns["dispose"].MatchString(content) {
		methods = append(methods, "dispose")
	}
	if fd.patterns["did_update_widget"].MatchString(content) {
		methods = append(methods, "didUpdateWidget")
	}
	
	return methods
}

// detectFeatures identifies Flutter features and widgets used
func (fd *FlutterDetector) detectFeatures(content string) []string {
	features := make([]string, 0)
	
	if fd.patterns["material_app"].MatchString(content) {
		features = append(features, "MaterialApp")
	}
	if fd.patterns["cupertino_app"].MatchString(content) {
		features = append(features, "CupertinoApp")
	}
	if fd.patterns["scaffold"].MatchString(content) {
		features = append(features, "Scaffold")
	}
	if fd.patterns["app_bar"].MatchString(content) {
		features = append(features, "AppBar")
	}
	if fd.patterns["floating_action"].MatchString(content) {
		features = append(features, "FloatingActionButton")
	}
	
	return features
}

// findBuildHelpers identifies helper build methods (private methods returning Widget)
func (fd *FlutterDetector) findBuildHelpers(content string) []string {
	helpers := make([]string, 0)
	
	if matches := fd.patterns["build_helper"].FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				helpers = append(helpers, match[1])
			}
		}
	}
	
	return helpers
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// FlutterAnalysis contains the results of Flutter content analysis
type FlutterAnalysis struct {
	IsFlutter         bool            `json:"is_flutter"`
	Framework         string          `json:"framework"`        // "flutter" | "none"
	UIFramework       string          `json:"ui_framework"`     // "material" | "cupertino" | "widgets" | "none"
	Widgets           []FlutterWidget `json:"widgets"`
	StateManagement   string          `json:"state_management"` // "riverpod" | "bloc" | "provider" | "getx" | "setState" | "none"
	Features          []string        `json:"features"`         // List of Flutter widgets/features used
	HasNavigation     bool            `json:"has_navigation"`
	LifecycleMethods  []string        `json:"lifecycle_methods"`
	CompositionDepth  int             `json:"composition_depth"`  // How many levels of widget composition
	BuildHelpers      []string        `json:"build_helpers"`      // Helper build methods like _buildHeader()
	HasOverride       bool            `json:"has_override"`       // Uses @override annotation
}

// FlutterWidget represents a Flutter widget found in the code
type FlutterWidget struct {
	Name           string `json:"name"`
	Type           string `json:"type"`            // "stateless" | "stateful" | "state"
	HasBuildMethod bool   `json:"has_build_method"`
}

// IntegrateFlutterAnalysis integrates Flutter analysis with existing symbol extraction
func (m *Manager) IntegrateFlutterAnalysis(ast *types.AST, analysis *FlutterAnalysis) {
	if analysis.IsFlutter && ast.Root != nil {
		if ast.Root.Metadata == nil {
			ast.Root.Metadata = make(map[string]interface{})
		}
		
		// Store comprehensive Flutter analysis in metadata
		ast.Root.Metadata["flutter_analysis"] = analysis
		ast.Root.Metadata["has_flutter"] = true
		ast.Root.Metadata["flutter_framework"] = analysis.UIFramework
		ast.Root.Metadata["state_management"] = analysis.StateManagement
		ast.Root.Metadata["has_navigation"] = analysis.HasNavigation
	}
}