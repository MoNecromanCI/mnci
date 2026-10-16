package scaffolding

// Versions new packages pin; bump them deliberately.
const (
	SDKConstraint = "^3.7.0"
	LintsVersion  = "^6.0.0"
	TestVersion   = "^1.25.0"
)

var pubspecTemplate = map[string]string{
	"[[.Dir]]/pubspec.yaml": `name: [[.Package]]
description: [[.Name]]
version: 0.0.0
[[- if ne .Kind "package"]]
publish_to: none
[[- end]]
resolution: workspace

environment:
  sdk: ` + SDKConstraint + `
[[- if .Flutter]]

dependencies:
  flutter:
    sdk: flutter

dev_dependencies:
  flutter_lints: ` + LintsVersion + `
  flutter_test:
    sdk: flutter

flutter:
  uses-material-design: true
[[- else]]

dev_dependencies:
  lints: ` + LintsVersion + `
  test: ` + TestVersion + `
[[- end]]

smith:
  kind: [[.Kind]]
[[- if .Stack]]
  stack: [[.Stack]]
[[- end]]
`,
	"[[.Dir]]/analysis_options.yaml": `[[if .Flutter]]include: package:flutter_lints/flutter.yaml[[else]]include: package:lints/recommended.yaml[[end]]
`,
}

// The greeting slice, a worked example of a vertical slice in Dart: a
// folder per outcome with its barrel, a file per role, the tests mirrored
// under test/.
var sliceTemplates = map[string]string{
	"[[.Dir]]/lib/src/greeting/greeting.dart": `export 'greet_use_case.dart' show greet;
`,
	"[[.Dir]]/lib/src/greeting/greet_use_case.dart": `/// Greets someone by name.
String greet(String name) => 'Hello, $name!';
`,
	"[[.Dir]]/test/src/greeting/greet_use_case_test.dart": `import 'package:[[.Package]]/src/greeting/greet_use_case.dart';
import 'package:[[if .Flutter]]flutter_test/flutter_test.dart[[else]]test/test.dart[[end]]';

void main() {
  test('greets by name', () {
    expect(greet('world'), 'Hello, world!');
  });
}
`,
}

// libraryTemplates are a package's public API: everything else is lib/src.
var libraryTemplates = map[string]string{
	"[[.Dir]]/lib/[[.Package]].dart": `/// [[.Name]]: its code lives in the slices under src/, one per outcome.
library;

export 'src/greeting/greeting.dart';
`,
}

var dartAppTemplates = map[string]string{
	"[[.Dir]]/bin/main.dart": `import 'package:[[.Package]]/[[.Package]].dart';

void main() {
  print(greet('world'));
}
`,
}

// flutterAppTemplates replace what flutter create writes; it keeps web/.
var flutterAppTemplates = map[string]string{
	"[[.Dir]]/lib/main.dart": `import 'package:flutter/material.dart';

import 'src/greeting/greeting.dart';

void main() {
  runApp(const MainApp());
}

/// The app's root widget.
class MainApp extends StatelessWidget {
  const MainApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(body: Center(child: Text(greet('world')))),
    );
  }
}
`,
}

// flutterCreated are the files flutter create writes that smith writes its
// own way, or that a workspace member does not have.
var flutterCreated = []string{"pubspec.yaml", "pubspec.lock", "analysis_options.yaml", "README.md", "lib/main.dart", ".idea"}
