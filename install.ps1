# Install smith from its GitHub Release on Windows.
#
#   irm https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.ps1 | iex
#   & ([scriptblock]::Create((irm .../install.ps1))) -Version 0.1.0
#
# Installs to $env:SMITH_INSTALL_DIR (default $HOME\.local\bin) and adds it to the
# user PATH.
param([string]$Version = $(if ($env:SMITH_VERSION) { $env:SMITH_VERSION } else { 'latest' }))
$ErrorActionPreference = 'Stop'
$repo = 'monosmith-app/monosmith'
$dir = if ($env:SMITH_INSTALL_DIR) { $env:SMITH_INSTALL_DIR } else { Join-Path $HOME '.local\bin' }
$arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'Arm64') { 'arm64' } else { 'amd64' }

# Releases are tagged smith@<version>; each holds monosmith_<version>_<os>_<arch>.zip
# with smith.exe in it.
$tag, $product, $packed = 'smith', 'monosmith', 'smith.exe'

if ($Version -eq 'latest') {
  # GitHub limits anonymous API calls per address, which shared CI runners
  # exhaust; a token (CI provides one) lifts the limit.
  $headers = @{}
  if ($env:GITHUB_TOKEN) { $headers['Authorization'] = "Bearer $env:GITHUB_TOKEN" }
  $releases = Invoke-RestMethod -Headers $headers "https://api.github.com/repos/$repo/releases?per_page=100"
  # GitHub lists releases by tag name as text (smith@0.9.0 before smith@0.20.0),
  # so take the highest version, comparing its numbers.
  function Newest($name) {
    $releases | Where-Object { $_.tag_name -match "^$name@\d+\.\d+\.\d+$" -and -not $_.prerelease } |
      Sort-Object { [version]$_.tag_name.Substring($name.Length + 1) } | Select-Object -Last 1
  }
  $release = Newest 'smith'
  if (-not $release) {
    # Until smith@1.0.0 is released, the newest release is the tool's former
    # name, mnci; it is installed as smith.
    $release = Newest 'mnci'
    $tag, $product, $packed = 'mnci', 'mnci', 'mnci.exe'
  }
  if (-not $release) { throw 'no smith release found' }
  $Version = $release.tag_name.Substring($tag.Length + 1)
}
$Version = $Version.TrimStart('v')
$asset = "${product}_${Version}_windows_${arch}.zip"
$base = "https://github.com/$repo/releases/download/$tag%40$Version"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid())
New-Item -ItemType Directory -Force $tmp | Out-Null
try {
  Write-Host "Installing smith $Version (windows/$arch) into $dir"
  Invoke-WebRequest "$base/$asset" -OutFile (Join-Path $tmp $asset) -UseBasicParsing
  try {
    Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt') -UseBasicParsing
    $line = Get-Content (Join-Path $tmp 'checksums.txt') | Where-Object { $_ -like "* $asset" }
    if ($line) {
      $expected = ($line -split ' ')[0]
      $actual = (Get-FileHash (Join-Path $tmp $asset) -Algorithm SHA256).Hash.ToLower()
      if ($expected -ne $actual) { throw "checksum mismatch for $asset" }
    }
  } catch [System.Net.WebException] { }
  Expand-Archive (Join-Path $tmp $asset) -DestinationPath (Join-Path $tmp 'out') -Force
  New-Item -ItemType Directory -Force $dir | Out-Null
  Copy-Item (Join-Path $tmp "out\$packed") (Join-Path $dir 'smith.exe') -Force
} finally {
  Remove-Item -Recurse -Force $tmp
}
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $dir) {
  [Environment]::SetEnvironmentVariable('Path', "$dir;$userPath", 'User')
  Write-Host "Added $dir to your user PATH; open a new terminal to use smith."
}
if ($env:GITHUB_PATH) { Add-Content $env:GITHUB_PATH $dir }
& (Join-Path $dir 'smith.exe') --version
