// Package scaffolding writes a new Maven module in Java, adds it to the root
// aggregator pom, and writes that pom the first time: plugin versions,
// Checkstyle, and every module's build under dist/maven. The kotlin stack
// writes a Gradle module in Kotlin instead, included in the root settings,
// linted with ktlint and built under dist/gradle.
package scaffolding
