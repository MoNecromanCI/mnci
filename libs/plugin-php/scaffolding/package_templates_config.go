package scaffolding

// Versions the root composer.json and new packages pin; bump them
// deliberately.
const (
	PHPRequirement = ">=8.2"
	PHPUnitVersion = "^11.5"
	PHPStanVersion = "^2.1"
)

const rootManifest = "composer.json"

// rootTemplates are written once, by the first package: the root requires
// every package from its directory and holds the tools, so the workspace
// has one vendor/ and one composer.lock.
var rootTemplates = map[string]string{
	rootManifest: `{
    "name": "[[.Vendor]]/workspace",
    "description": "The workspace's PHP packages, resolved together (smith lists them).",
    "type": "project",
    "license": "MIT",
    "require": {
        "php": "` + PHPRequirement + `"
    },
    "require-dev": {
        "phpunit/phpunit": "` + PHPUnitVersion + `",
        "phpstan/phpstan": "` + PHPStanVersion + `"
    },
    "minimum-stability": "dev",
    "prefer-stable": true
}
`,
}

var manifestTemplate = map[string]string{
	"[[.Dir]]/composer.json": `{
    "name": "[[.Package]]",
    "description": "[[.Name]]",
    "type": "[[if eq .Kind "app"]]project[[else]]library[[end]]",
    "license": "MIT",
    "version": "0.0.0",
    "require": {
        "php": "` + PHPRequirement + `"
    },
    "autoload": {
        "psr-4": {
            "[[.NamespaceJSON]]\\": "src/"
        }
    },
[[- if eq .Kind "app"]]
    "bin": ["bin/[[.Name]]"],
[[- end]]
    "extra": {
        "smith": {
            "kind": "[[.Kind]]"[[if .Stack]],
            "stack": "[[.Stack]]"[[end]]
        }
    }
}
`,
}

// The greeting slice, a worked example of a vertical slice in PHP: a
// namespace (and folder) per outcome, a class per role, the tests mirrored
// under tests/. PSR-4 names each file after its class.
var sliceTemplates = map[string]string{
	"[[.Dir]]/src/Greeting/GreetUseCase.php": `<?php

declare(strict_types=1);

namespace [[.Namespace]]\Greeting;

/** Greets people: a worked example of a slice. Replace it with your own outcome. */
final class GreetUseCase
{
    /** Greets someone by name. */
    public static function greet(string $name): string
    {
        return "Hello, {$name}!";
    }
}
`,
	"[[.Dir]]/tests/Greeting/GreetUseCaseTest.php": `<?php

declare(strict_types=1);

namespace [[.Namespace]]\Tests\Greeting;

use [[.Namespace]]\Greeting\GreetUseCase;
use PHPUnit\Framework\TestCase;

final class GreetUseCaseTest extends TestCase
{
    public function testGreetsByName(): void
    {
        self::assertSame('Hello, world!', GreetUseCase::greet('world'));
    }
}
`,
}

var appTemplates = map[string]string{
	"[[.Dir]]/bin/[[.Name]]": `<?php

declare(strict_types=1);

use [[.Namespace]]\Greeting\GreetUseCase;

require $_composer_autoload_path ?? __DIR__ . '/[[.Up]]vendor/autoload.php';

echo GreetUseCase::greet('world'), PHP_EOL;
`,
}
