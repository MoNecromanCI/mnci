# Install mnci from its GitHub Release on Windows.
#
#   irm https://raw.githubusercontent.com/MoNecromanCI/mnci/main/install.ps1 | iex
#   & ([scriptblock]::Create((irm .../install.ps1))) -Version 0.1.0
#
# Installs to $env:MNCI_INSTALL_DIR (default $HOME\.local\bin) and adds it to the
# user PATH.
param([string]$Version = $(if ($env:MNCI_VERSION) { $env:MNCI_VERSION } else { 'latest' }))
$ErrorActionPreference = 'Stop'
$repo = 'MoNecromanCI/mnci'
$dir = if ($env:MNCI_INSTALL_DIR) { $env:MNCI_INSTALL_DIR } else { Join-Path $HOME '.local\bin' }
$arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'Arm64') { 'arm64' } else { 'amd64' }

if ($Version -eq 'latest') {
  # GitHub limits anonymous API calls per address, which shared CI runners
  # exhaust; a token (CI provides one) lifts the limit.
  $headers = @{}
  if ($env:GITHUB_TOKEN) { $headers['Authorization'] = "Bearer $env:GITHUB_TOKEN" }
  $release = (Invoke-RestMethod -Headers $headers "https://api.github.com/repos/$repo/releases?per_page=50") |
    Where-Object { $_.tag_name -like 'mnci@*' } | Select-Object -First 1
  if (-not $release) { throw 'no mnci release found' }
  $Version = $release.tag_name.Substring(5)
}
$Version = $Version.TrimStart('v')
$asset = "mnci_${Version}_windows_${arch}.zip"
$base = "https://github.com/$repo/releases/download/mnci%40$Version"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid())
New-Item -ItemType Directory -Force $tmp | Out-Null
try {
  Write-Host "Installing mnci $Version (windows/$arch) into $dir"
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
  Copy-Item (Join-Path $tmp 'out\mnci.exe') (Join-Path $dir 'mnci.exe') -Force
} finally {
  Remove-Item -Recurse -Force $tmp
}
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $dir) {
  [Environment]::SetEnvironmentVariable('Path', "$dir;$userPath", 'User')
  Write-Host "Added $dir to your user PATH; open a new terminal to use mnci."
}
if ($env:GITHUB_PATH) { Add-Content $env:GITHUB_PATH $dir }
& (Join-Path $dir 'mnci.exe') --version
