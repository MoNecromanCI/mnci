#!/usr/bin/env node
// Written by mnci: 'mnci add go-app --release' creates it and 'mnci upgrade'
// rewrites it, so local edits do not survive an upgrade.
//
// A releasable Go app (tag release:go) is versioned from its git tag, with no
// manifest, and its per-platform zips are attached to its GitHub Release:
//   release.version.versionActions   this file, loaded by 'nx release'
//   node tools/go-app-release.cjs assets            the apps built for all six platforms
//   node tools/go-app-release.cjs assets --native   the apps that need a C toolchain, this OS only
//
// What an attached file is called is the app's own choice, in its project.json (#317):
//   "release": { "asset": {
//     "name": "{product}_{version}_{os}_{arch}.{ext}",   placeholders: product, version, os, arch, ext
//     "product": "mvd",                                  default: the project name
//     "osAlias": { "darwin": "macos" },                  Go's name -> the one you write
//     "archAlias": {},
//     "extra": [{ "target": "package-dmg", "files": "dist/dmg/mvd_{version}_macos_universal.dmg" }]
//   } }
// Without "name" the zips keep the name package-all gave them. Each "extra" runs the app's target with VERSION
// set and attaches what "files" matches (a glob, {version} filled in), under the name it already has.
'use strict'
const { spawnSync } = require('node:child_process')
const { copyFileSync, existsSync, globSync, mkdirSync, readdirSync, readFileSync, rmSync } = require('node:fs')
const { dirname, join } = require('node:path')
const { VersionActions } = require('nx/release')

const FIRST_RELEASE_BASE = '0.0.0'

class GoAppVersionActions extends VersionActions {
  validManifestFilenames = null

  async validate () {
    // Nothing on disk to check: there is no manifest.
  }

  async readCurrentVersionFromSourceManifest () {
    // Reached only as the fallback of the git-tag resolver, i.e. before the first tag.
    return { currentVersion: FIRST_RELEASE_BASE, manifestPath: 'none (a Go app is versioned by its git tag)' }
  }

  async readCurrentVersionFromRegistry () {
    return null
  }

  async readCurrentVersionOfDependency () {
    return { currentVersion: null, dependencyCollection: null }
  }

  async updateProjectVersion () {
    return []
  }

  async updateProjectDependencies () {
    return []
  }
}

module.exports = GoAppVersionActions
module.exports.assetName = (...arguments_) => assetName(...arguments_)

function fail (message) {
  console.error(message)
  process.exit(1)
}

function run (command, args, env) {
  const result = spawnSync(command, args, { encoding: 'utf8', stdio: ['ignore', 'pipe', 'inherit'], env: { ...process.env, ...env }, shell: false })
  if (result.status !== 0) fail(command + ' ' + args.join(' ') + ' failed (exit ' + result.status + ')')

  return result.stdout
}

function nx (args, env) {
  const manifest = require.resolve('nx/package.json')
  const bin = require(manifest).bin
  const entry = typeof bin === 'string' ? bin : bin.nx

  return run(process.execPath, [join(dirname(manifest), entry), ...args], env)
}

const ASSET_PLACEHOLDERS = ['product', 'version', 'os', 'arch', 'ext']

function assetConfig (app) {
  try {
    return JSON.parse(readFileSync(join('apps', app, 'project.json'), 'utf8')).release?.asset || {}
  } catch {
    return {}
  }
}

// 'go-app-<app>-<goos>-<goarch>.zip' -> { os, arch, ext }, or undefined for any other file.
function parseBuiltName (app, file) {
  const prefix = 'go-app-' + app + '-'
  if (!file.startsWith(prefix) || !file.endsWith('.zip')) return null
  const platform = file.slice(prefix.length, -'.zip'.length)
  const dash = platform.lastIndexOf('-')
  if (dash < 1) return null

  return { os: platform.slice(0, dash), arch: platform.slice(dash + 1), ext: 'zip' }
}

// What a built zip is attached as. Without a configured name the file keeps the name it was built with.
function assetName (app, version, file, config) {
  const parts = parseBuiltName(app, file)
  if (!parts || !config.name) return file

  const os = (config.osAlias || {})[parts.os] || parts.os
  const arch = (config.archAlias || {})[parts.arch] || parts.arch
  const values = { product: config.product || app, version, os, arch, ext: parts.ext }

  return config.name.replaceAll(/\{(\w+)\}/g, (_match, key) => {
    if (!ASSET_PLACEHOLDERS.includes(key)) fail(app + ': release.asset.name has an unknown placeholder {' + key + '} - use ' + ASSET_PLACEHOLDERS.map(each => '{' + each + '}').join(', ') + '.')

    return values[key]
  })
}

function releasableApps (native) {
  const names = existsSync('apps') ? readdirSync('apps') : []

  return names.filter(name => {
    try {
      const tags = JSON.parse(readFileSync(join('apps', name, 'project.json'), 'utf8')).tags || []

      return tags.includes('release:go') && tags.includes('build:cgo') === native
    } catch {
      return false
    }
  })
}

function attachAssets (native) {
  const apps = releasableApps(native)
  if (apps.length === 0) {
    console.log('No releasable ' + (native ? 'native ' : '') + 'Go app - nothing to attach.')

    return
  }
  const target = native ? 'package-native' : 'package-all'
  const tags = run('git', ['tag', '--points-at', 'HEAD']).split(/\r?\n/).filter(Boolean)
  let attached = 0
  for (const app of apps) {
    const tag = tags.find(each => each.startsWith(app + '@'))
    if (!tag) {
      console.log(app + ': not released by this run - skipping.')
      continue
    }
    const version = tag.slice(app.length + 1)
    console.log(app + ': ' + (native ? 'building for this OS' : 'building the six platforms') + ' as ' + version)
    nx(['run', app + ':' + target], { VERSION: version })
    const config = assetConfig(app)
    const built = existsSync('dist/drop') ? readdirSync('dist/drop').filter(each => parseBuiltName(app, each)) : []
    if (built.length === 0) fail(app + ': ' + target + ' produced no go-app-' + app + '-*.zip in dist/drop.')
    const staging = join('dist', 'release-assets', app)
    rmSync(staging, { recursive: true, force: true })
    mkdirSync(staging, { recursive: true })
    const files = built.map(each => {
      const name = assetName(app, version, each, config)
      copyFileSync(join('dist/drop', each), join(staging, name))

      return join(staging, name)
    })
    const extras = config.extra || []
    for (const extra of extras) {
      if (!extra.target || !extra.files) fail(app + ': each release.asset.extra needs a target and files.')
      nx(['run', app + ':' + extra.target], { VERSION: version })
      const found = globSync(extra.files.replaceAll('{version}', () => version))
      if (found.length === 0) fail(app + ': ' + extra.target + ' produced nothing matching ' + extra.files + '.')
      files.push(...found)
    }
    const upload = spawnSync('gh', ['release', 'upload', tag, ...files, '--clobber'], { stdio: 'inherit', shell: process.platform === 'win32' })
    if (upload.status !== 0) fail(app + ': could not attach the files to the ' + tag + ' release (exit ' + upload.status + ').')
    console.log(app + ': attached ' + files.length + ' file(s) to ' + tag)
    attached += 1
  }
  console.log(attached + ' release(s) carry their platform zips.')
}

if (require.main === module) {
  if (process.argv[2] !== 'assets' || (process.argv[3] !== undefined && process.argv[3] !== '--native')) fail('Usage: node tools/go-app-release.cjs assets [--native]')
  attachAssets(process.argv[3] === '--native')
}
