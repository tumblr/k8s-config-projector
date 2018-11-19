#!/usr/bin/env ruby
# script will, given a path for projection manifests
# figure out what namespaces need to be created ahead of time
# output: list of <cluster>/<namespace> ...
# example: bf2-DEVEL/foo bf2-PRODUCTION/bar
require 'yaml'
require 'find'
require 'pathname'

path = ARGV[0] || '.'
abort "Require manifests path in as first parameter" if path.nil?
abort "#{path} is not a directory" unless File.directory? path

# now, read all yamls and figure out their namespaces
nstuples = Find.find(path).map do |p|
  if File.file?(p) && File.readable?(p) && p.end_with?('.yaml')
    # we need to determine what cluster this namespace is 
    # based on directory structure.
    fullpath = File.expand_path(p)
    path_components = fullpath.split(File::SEPARATOR)
    az = path_components[-4]
    cluster = path_components[-3]

    manifest = YAML.load_file(p)
    namespace = manifest['namespace']
    if namespace.nil? || az.nil? || cluster.nil?
      nil
    else
      ["#{az}-#{cluster}", namespace]
    end
  else
    nil
  end
end.compact

nstuples.each do |(cluster, ns)|
  puts "#{cluster}/#{ns}"
end
