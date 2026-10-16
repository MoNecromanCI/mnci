package scaffolding

// Versions the workspace's Gemfile and new gems pin; bump them deliberately.
const (
	RubyRequirement = ">= 3.2"
	RSpecVersion    = "~> 3.13"
	RuboCopVersion  = "~> 1.70"
	RakeVersion     = "~> 13.0"
)

const (
	gemfile = "Gemfile"
	rubocop = ".rubocop.yml"
)

// rootTemplates are written once, by the first gem.
var rootTemplates = map[string]string{
	gemfile: `# frozen_string_literal: true

source 'https://rubygems.org'

# The workspace's gems, each from its own directory (mnci lists them): one
# resolution, one Gemfile.lock, for all of them.

group :development, :test do
  gem 'rake', '` + RakeVersion + `'
  gem 'rspec', '` + RSpecVersion + `'
  gem 'rubocop', '` + RuboCopVersion + `', require: false
end
`,
	rubocop: `AllCops:
  TargetRubyVersion: 3.2
  NewCops: enable
  SuggestExtensions: false
  Exclude:
    - "dist/**/*"
    - "node_modules/**/*"
    - "vendor/**/*"

# The same files on every OS.
Layout/EndOfLine:
  EnforcedStyle: lf
`,
}

var gemTemplate = map[string]string{
	"[[.Dir]]/[[.Gem]].gemspec": `# frozen_string_literal: true

Gem::Specification.new do |spec|
  spec.name = '[[.Gem]]'
  spec.version = '0.0.0'
  spec.summary = '[[.Name]]'
  spec.authors = ['[[.Workspace]]']
  spec.license = 'MIT'
  spec.required_ruby_version = '` + RubyRequirement + `'
  spec.files = Dir['lib/**/*.rb'[[if eq .Kind "app"]], 'exe/*'[[end]]]
[[- if eq .Kind "app"]]
  spec.bindir = 'exe'
  spec.executables = ['[[.Gem]]']
[[- end]]
  spec.require_paths = ['lib']
  spec.metadata['rubygems_mfa_required'] = 'true'
[[- if ne .Kind "package"]]
  # Released by its tag, never pushed to a gem server.
  spec.metadata['allowed_push_host'] = 'none'
[[- end]]
  spec.metadata['mnci_kind'] = '[[.Kind]]'
[[- if .Stack]]
  spec.metadata['mnci_stack'] = '[[.Stack]]'
[[- end]]
end
`,
}

// The greeting slice, a worked example of a vertical slice in Ruby: a
// folder per outcome with the file that requires it, a file per role, the
// specs mirrored under spec/.
var sliceTemplates = map[string]string{
	"[[.Dir]]/lib/[[.Gem]].rb": `# frozen_string_literal: true

require_relative '[[.Gem]]/greeting/greeting'

# [[.Name]]: its code lives in the slices under [[.Gem]]/, one per outcome.
module [[.Module]]
end
`,
	"[[.Dir]]/lib/[[.Gem]]/greeting/greeting.rb": `# frozen_string_literal: true

require_relative 'greet_use_case'
`,
	"[[.Dir]]/lib/[[.Gem]]/greeting/greet_use_case.rb": `# frozen_string_literal: true

module [[.Module]]
  # Greets people: a worked example of a slice. Replace it with your own outcome.
  module Greeting
    # Greets someone by name.
    def self.greet(name)
      "Hello, #{name}!"
    end
  end
end
`,
	"[[.Dir]]/spec/[[.Gem]]/greeting/greet_use_case_spec.rb": `# frozen_string_literal: true

require '[[.Gem]]'

RSpec.describe [[.Module]]::Greeting do
  it 'greets by name' do
    expect(described_class.greet('world')).to eq('Hello, world!')
  end
end
`,
}

var appTemplates = map[string]string{
	"[[.Dir]]/exe/[[.Gem]]": `#!/usr/bin/env ruby
# frozen_string_literal: true

require '[[.Gem]]'

puts [[.Module]]::Greeting.greet('world')
`,
}
