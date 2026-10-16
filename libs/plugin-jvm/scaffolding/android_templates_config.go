package scaffolding

// AndroidStack is the stack that writes an Android app in Kotlin.
const AndroidStack = "android"

// Versions the Android templates pin; bump them deliberately. AGP builds
// Kotlin itself (no Kotlin Android plugin).
const (
	AGPVersion = "9.4.1"
	CompileSDK = "36"
	MinSDK     = "26"
)

// androidPlugin is the root script's line that pins AGP for every Android
// module.
const androidPlugin = `id("com.android.application") version "` + AGPVersion + `" apply false`

// androidTemplates are an Android app: its script and properties, the
// manifest, a screen showing the greeting slice, and the slice's test. The
// build goes to the workspace's dist/, never build/ in a source tree.
var androidTemplates = map[string]string{
	"[[.Dir]]/gradle.properties": `smith.kind=[[.Kind]]
[[- if .Stack]]
smith.stack=[[.Stack]]
[[- end]]
`,
	"[[.Dir]]/build.gradle.kts": `plugins {
    id("com.android.application")
    id("org.jlleitschuh.gradle.ktlint")
}

group = "[[.Group]]"
version = "0.0.0"

layout.buildDirectory = rootProject.layout.projectDirectory.dir("dist/gradle/[[.ArtifactID]]")

android {
    namespace = "[[.Package]]"
    compileSdk = ` + CompileSDK + `
    defaultConfig {
        applicationId = "[[.Package]]"
        minSdk = ` + MinSDK + `
        targetSdk = ` + CompileSDK + `
        versionCode = 1
        versionName = project.version.toString()
    }
}

dependencies {
    testImplementation("junit:junit:4.13.2")
}
`,
	"[[.Dir]]/src/main/AndroidManifest.xml": `<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android">
    <application android:label="[[.Name]]">
        <activity android:name=".MainActivity" android:exported="true">
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>
    </application>
</manifest>
`,
	"[[.Dir]]/src/main/kotlin/[[.PackagePath]]/MainActivity.kt": `package [[.Package]]

import android.app.Activity
import android.os.Bundle
import android.widget.TextView
import [[.Package]].greeting.GreetUseCase

/** The app's screen: the greeting. */
class MainActivity : Activity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(TextView(this).apply { text = GreetUseCase.greet("world") })
    }
}
`,
	"[[.Dir]]/src/main/kotlin/[[.PackagePath]]/greeting/GreetUseCase.kt": `package [[.Package]].greeting

/** Greets someone by name. */
object GreetUseCase {
    fun greet(name: String): String = "Hello, $name!"
}
`,
	"[[.Dir]]/src/test/kotlin/[[.PackagePath]]/greeting/GreetUseCaseTest.kt": `package [[.Package]].greeting

import org.junit.Assert.assertEquals
import org.junit.Test

class GreetUseCaseTest {
    @Test
    fun greetsByName() {
        assertEquals("Hello, world!", GreetUseCase.greet("world"))
    }
}
`,
}
