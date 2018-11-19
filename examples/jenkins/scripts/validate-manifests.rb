#!/usr/bin/env ruby
# script will validate that each manifest found is valid
require 'yaml'
require 'find'

path = ARGV[0] || '.'

# now, read all yamls and figure out their namespaces
Find.find(path) do |p|
  if File.file?(p) && File.readable?(p) && p.end_with?('.yaml')
    puts "Validating #{p}..."
    begin
      manifest = YAML.load_file(p)
    rescue => e
      abort "#{p} vailed to parse as valid YAML! #{e.message}"
    end
    name = manifest['name']
    namespace = manifest['namespace']
    nspath = File.basename(File.dirname(p))
    if nspath != namespace
      abort "Manifest #{p} is in a directory called '#{nspath}', but should be in '#{namespace}'!"
    end
    if name.match(%r|[^a-zA-Z0-9\-]+|)
      abort "ERROR: Manifest #{p} is named '#{name}'. \n       Please check your characters and make sure they are valid for K8s (/[a-zA-Z0-9\-]+/)."
    end
    filename = File.basename(p)
    if filename != "#{name}.yaml"
      abort "ERROR: Manifest #{p} is named '#{filename}', but should be in '#{name}.yaml' (filename and name field should align)!"
    end

  end
end
