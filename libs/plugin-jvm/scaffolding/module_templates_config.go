package scaffolding

// Versions the root pom pins; bump them deliberately.
const (
	JavaRelease  = "21"
	JUnitVersion = "5.11.4"
	rootPom      = "pom.xml"
)

// rootTemplates are written once, by the first module. The root pom is the
// aggregator (its <modules> list every module) and every module's parent.
var rootTemplates = map[string]string{
	// mnci runs Maven tasks in parallel; file locks keep processes that
	// download into the same local repository from corrupting it.
	".mvn/maven.config": `-Daether.syncContext.named.factory=file-lock
-Daether.syncContext.named.nameMapper=file-gav
`,
	rootPom: `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>[[.Group]]</groupId>
  <artifactId>[[.Workspace]]-parent</artifactId>
  <version>0.0.0</version>
  <packaging>pom</packaging>

  <modules>
  </modules>

  <properties>
    <maven.compiler.release>` + JavaRelease + `</maven.compiler.release>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
    <!-- exec:java runs apps only: an app sets exec.skip to false and its
         exec.mainClass. -->
    <exec.skip>true</exec.skip>
    <exec.mainClass>none</exec.mainClass>
  </properties>

  <dependencyManagement>
    <dependencies>
      <dependency>
        <groupId>org.junit.jupiter</groupId>
        <artifactId>junit-jupiter</artifactId>
        <version>` + JUnitVersion + `</version>
        <scope>test</scope>
      </dependency>
    </dependencies>
  </dependencyManagement>

  <build>
    <!-- Every build goes to the workspace's dist/ (mnci), never target/ in
         a source tree. -->
    <directory>${maven.multiModuleProjectDirectory}/dist/maven/${project.artifactId}</directory>
    <plugins>
      <!-- A module is installed and deployed with a self-contained pom (no
           parent), so modules that depend on it, built on their own, resolve
           it without this parent in any repository. -->
      <plugin>
        <groupId>org.codehaus.mojo</groupId>
        <artifactId>flatten-maven-plugin</artifactId>
        <version>1.7.0</version>
        <configuration>
          <outputDirectory>${project.build.directory}</outputDirectory>
          <flattenMode>ossrh</flattenMode>
        </configuration>
        <executions>
          <execution>
            <id>flatten</id>
            <phase>process-resources</phase>
            <goals><goal>flatten</goal></goals>
          </execution>
        </executions>
      </plugin>
    </plugins>
    <pluginManagement>
      <plugins>
        <plugin><artifactId>maven-compiler-plugin</artifactId><version>3.13.0</version></plugin>
        <plugin><artifactId>maven-surefire-plugin</artifactId><version>3.5.2</version></plugin>
        <plugin><artifactId>maven-jar-plugin</artifactId><version>3.4.2</version></plugin>
        <plugin><artifactId>maven-resources-plugin</artifactId><version>3.3.1</version></plugin>
        <plugin><artifactId>maven-install-plugin</artifactId><version>3.1.3</version></plugin>
        <plugin><artifactId>maven-deploy-plugin</artifactId><version>3.1.3</version></plugin>
        <plugin><groupId>org.codehaus.mojo</groupId><artifactId>exec-maven-plugin</artifactId><version>3.5.0</version></plugin>
        <plugin>
          <artifactId>maven-checkstyle-plugin</artifactId>
          <version>3.6.0</version>
          <configuration>
            <configLocation>google_checks.xml</configLocation>
            <consoleOutput>true</consoleOutput>
            <violationSeverity>warning</violationSeverity>
            <includeTestSourceDirectory>true</includeTestSourceDirectory>
          </configuration>
        </plugin>
      </plugins>
    </pluginManagement>
  </build>
</project>
`,
}

var moduleTemplate = map[string]string{
	"[[.Dir]]/pom.xml": `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <parent>
    <groupId>[[.Group]]</groupId>
    <artifactId>[[.Workspace]]-parent</artifactId>
    <version>0.0.0</version>
    <relativePath>[[.Up]]pom.xml</relativePath>
  </parent>
  <artifactId>[[.ArtifactID]]</artifactId>
  <version>0.0.0</version>
  <name>[[.Name]]</name>

  <properties>
    <mnci.kind>[[.Kind]]</mnci.kind>
[[- if .Stack]]
    <mnci.stack>[[.Stack]]</mnci.stack>
[[- end]]
[[- if eq .Kind "app"]]
    <exec.skip>false</exec.skip>
    <exec.mainClass>[[.Package]].Main</exec.mainClass>
[[- end]]
  </properties>

  <dependencies>
    <dependency>
      <groupId>org.junit.jupiter</groupId>
      <artifactId>junit-jupiter</artifactId>
    </dependency>
  </dependencies>
</project>
`,
}

// The greeting slice, a worked example of a vertical slice in Java: a
// package per outcome, a class per role, the tests mirrored under src/test.
var sliceTemplates = map[string]string{
	"[[.Dir]]/src/main/java/[[.PackagePath]]/greeting/GreetUseCase.java": `package [[.Package]].greeting;

/** Greets someone by name. */
public final class GreetUseCase {
  private GreetUseCase() {}

  public static String greet(String name) {
    return "Hello, " + name + "!";
  }
}
`,
	"[[.Dir]]/src/test/java/[[.PackagePath]]/greeting/GreetUseCaseTest.java": `package [[.Package]].greeting;

import static org.junit.jupiter.api.Assertions.assertEquals;

import org.junit.jupiter.api.Test;

class GreetUseCaseTest {
  @Test
  void greetsByName() {
    assertEquals("Hello, world!", GreetUseCase.greet("world"));
  }
}
`,
}

var appTemplates = map[string]string{
	"[[.Dir]]/src/main/java/[[.PackagePath]]/Main.java": `package [[.Package]];

import [[.Package]].greeting.GreetUseCase;

/** The app's entry point. */
public final class Main {
  private Main() {}

  public static void main(String[] args) {
    System.out.println(GreetUseCase.greet("world"));
  }
}
`,
}
