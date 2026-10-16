package pipeline

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

// AndroidToolsRevision is the Android command-line tools build installed
// where there is no SDK; bump it deliberately.
const AndroidToolsRevision = "16111833"

// androidSDK installs the Android SDK where none is (GitHub's and Azure's
// hosted images carry one at ANDROID_HOME): the command-line tools, with
// the licences accepted so AGP downloads the platform and build tools a
// module asks for. ANDROID_HOME reaches the later steps on GitHub
// ($GITHUB_ENV) and Azure (a logging command).
func androidSDK() protocol.Toolchain {
	return protocol.Toolchain{
		Name: "android-sdk", Version: AndroidToolsRevision,
		Install: "if [ -z \"$ANDROID_HOME\" ] || [ ! -d \"$ANDROID_HOME\" ]; then\n" +
			"  export ANDROID_HOME=\"$HOME/android-sdk\"\n" +
			"  mkdir -p \"$ANDROID_HOME/cmdline-tools\"\n" +
			"  curl -fsSLo /tmp/android-tools.zip https://dl.google.com/android/repository/commandlinetools-linux-" + AndroidToolsRevision + "_latest.zip\n" +
			"  unzip -q -o /tmp/android-tools.zip -d \"$ANDROID_HOME/cmdline-tools\" && rm /tmp/android-tools.zip\n" +
			"  mv \"$ANDROID_HOME/cmdline-tools/cmdline-tools\" \"$ANDROID_HOME/cmdline-tools/latest\"\n" +
			"  yes | \"$ANDROID_HOME/cmdline-tools/latest/bin/sdkmanager\" --licenses > /dev/null\n" +
			"  if [ -n \"$GITHUB_ENV\" ]; then echo \"ANDROID_HOME=$ANDROID_HOME\" >> \"$GITHUB_ENV\"; fi\n" +
			"  echo \"##vso[task.setvariable variable=ANDROID_HOME]$ANDROID_HOME\"\n" +
			"fi",
	}
}

var sdkDir = regexp.MustCompile(`(?m)^sdk\.dir\s*=\s*(.+)$`)

// androidSDKFound reports whether Gradle will find an SDK: ANDROID_HOME
// (or the older ANDROID_SDK_ROOT), else the workspace's local.properties.
func androidSDKFound(workspaceRoot string) bool {
	for _, env := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		if dir := os.Getenv(env); dir != "" {
			if _, err := os.Stat(dir); err == nil {
				return true
			}
		}
	}
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, "local.properties"))
	if err != nil {
		return false
	}
	m := sdkDir.FindSubmatch(raw)
	if m == nil {
		return false
	}
	_, err = os.Stat(string(m[1]))
	return err == nil
}
