package scaffolding

// KotlinStack is the stack that writes a Kotlin module built with Gradle.
const KotlinStack = "kotlin"

// Versions the Gradle templates pin; bump them deliberately.
const (
	KotlinVersion = "2.4.21"
	KtlintVersion = "14.2.0"
	JUnitBOM      = "6.1.3"
	settings      = "settings.gradle.kts"
	rootScript    = "build.gradle.kts"
)

// gradleRootTemplates are written once, by the first Gradle module: the
// settings list every module and the repositories (Google's for Android),
// the root script pins the plugins' versions.
var gradleRootTemplates = map[string]string{
	settings: `rootProject.name = "[[.Workspace]]"

pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}

dependencyResolutionManagement {
    repositories {
        google()
        mavenCentral()
    }
}
`,
	rootScript: `// The plugins' versions, for every module; each module applies the ones it
// uses.
plugins {
    kotlin("jvm") version "` + KotlinVersion + `" apply false
    id("org.jlleitschuh.gradle.ktlint") version "` + KtlintVersion + `" apply false
}
`,
}

// gradleModuleTemplates are a Kotlin module's script and properties. Its
// build goes to the workspace's dist/ (mnci), never build/ in a source tree.
var gradleModuleTemplates = map[string]string{
	"[[.Dir]]/gradle.properties": `mnci.kind=[[.Kind]]
[[- if .Stack]]
mnci.stack=[[.Stack]]
[[- end]]
`,
	"[[.Dir]]/build.gradle.kts": `import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
    kotlin("jvm")
    id("org.jlleitschuh.gradle.ktlint")
[[- if eq .Kind "app"]]
    application
[[- end]]
[[- if eq .Kind "package"]]
    ` + "`maven-publish`" + `
[[- end]]
}

group = "[[.Group]]"
version = "0.0.0"

layout.buildDirectory = rootProject.layout.projectDirectory.dir("dist/gradle/[[.ArtifactID]]")

java {
    sourceCompatibility = JavaVersion.VERSION_` + JavaRelease + `
    targetCompatibility = JavaVersion.VERSION_` + JavaRelease + `
}

kotlin {
    compilerOptions {
        jvmTarget = JvmTarget.JVM_` + JavaRelease + `
    }
}
[[- if eq .Kind "app"]]

application {
    mainClass = "[[.Package]].MainKt"
}
[[- end]]

dependencies {
    testImplementation(platform("org.junit:junit-bom:` + JUnitBOM + `"))
    testImplementation("org.junit.jupiter:junit-jupiter")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}

tasks.test {
    useJUnitPlatform()
}
[[- if eq .Kind "package"]]

// mnci release publishes to MAVEN_REPOSITORY_URL with MAVEN_USERNAME and
// MAVEN_PASSWORD.
publishing {
    publications {
        create<MavenPublication>("maven") {
            artifactId = "[[.ArtifactID]]"
            from(components["java"])
        }
    }
    repositories {
        System.getenv("MAVEN_REPOSITORY_URL")?.let { repository ->
            maven {
                name = "mnci"
                url = uri(repository)
                credentials {
                    username = System.getenv("MAVEN_USERNAME")
                    password = System.getenv("MAVEN_PASSWORD")
                }
            }
        }
    }
}
[[- end]]
`,
}

// The greeting slice in Kotlin: a package per outcome, a file per role, the
// tests mirrored under src/test.
var kotlinSliceTemplates = map[string]string{
	"[[.Dir]]/src/main/kotlin/[[.PackagePath]]/greeting/GreetUseCase.kt": `package [[.Package]].greeting

/** Greets someone by name. */
object GreetUseCase {
    fun greet(name: String): String = "Hello, $name!"
}
`,
	"[[.Dir]]/src/test/kotlin/[[.PackagePath]]/greeting/GreetUseCaseTest.kt": `package [[.Package]].greeting

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class GreetUseCaseTest {
    @Test
    fun greetsByName() {
        assertEquals("Hello, world!", GreetUseCase.greet("world"))
    }
}
`,
}

var kotlinAppTemplates = map[string]string{
	"[[.Dir]]/src/main/kotlin/[[.PackagePath]]/Main.kt": `package [[.Package]]

import [[.Package]].greeting.GreetUseCase

/** The app's entry point. */
fun main() {
    println(GreetUseCase.greet("world"))
}
`,
}
