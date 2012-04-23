/*
The rx command is a dependency and version management system for Go projects.
It is built on top of the go tool and utilizes the $GOPATH convention.

Installation

As usual, the rx tool can be installed or upgraded via the "go" tool:
    go get -u kylelemons.net/go/rx

General Usage

The rx command is composed of numerous sub-commands.
Sub-commands can be abbreviated to any unique prefix on the command-line.
The general usage is:

    rx [<options>] [<subcommand> [<suboptions>] [<arguments> ...]]

Options:
  --autosave = true         Automatically save dependency graph (disable for concurrent runs)
  --rescan   = false        Force a rescan of repositories
  --rxdir    = $HOME/.rx    Directory in which to save state
  -v         = false        Turn on verbose logging

Commands:
    help       Help on the rx command and subcommands.
    list       List recognized repositories.
    tags       List known repository tags.
    prescribe  Update the repository to the given tag/rev.

Use "rx help <command>" for more help with a command.


See below for a description of the various sub-commands understood by rx.

Help Command

Help on the rx command and subcommands.

Usage:
    rx help [<command>]

Options:
  --godoc = false    Dump the godoc output for the command(s)


List Command

List recognized repositories.

Usage:
    rx list [<filter>]

Options:
  -f                List output format
  --long = false    Use long output format

The list command scans all available packages and collects information about
their repositories.  By default, each repository is listed along with its
dependencies and contained packages. If a <filter> regular expression is
provided, only repositories whose root path matches the filter will be listed.

The -f option takes a template as a format.  The data passed into the
template invocation is an (rx/graph) Graph, and the default format is:

  {{range .Repository}}{{.}} :{{range .Packages}}{{$pkg := index $.Package .}} {{$pkg.Name}}{{end}}
  {{end}}

If you specify --long, the format will be:

  {{range .Repository}}Repository ({{.VCS}}) {{.}}:
      Packages:{{range .Packages}}
          {{$pkg := index $.Package .}}{{$pkg.ImportPath}}{{end}}
  {{with $.RepoDeps .}}    Dependencies:{{range .}}
          {{.}}{{end}}
  {{end}}{{with $.RepoUsers .}}    Users:{{range .}}
          {{.}}{{end}}
  {{end}}
  {{end}}
Tags Command

List known repository tags.

Usage:
    rx tags <repo>

Options:
  --down = false    Only show downgrades
  -f                tags output format
  --long = false    Use long output format
  --up   = false    Only show updates (overrides --down)

The tags command scans the specified repository and lists
information about its tags.  The <repo> can be any piece of the repository root
path, as long as it is unique.

The -f option takes a template as a format.  The data passed into the
template invocation is an (rx/graph) TagList, and the default format is:

  {{range .}}{{.Rev}} {{.Name}}
  {{end}}
Prescribe Command

Update the repository to the given tag/rev.

Usage:
    rx prescribe <repo> <tag>

Options:
  --build    = true     build all updated packages
  --cascade  = true     recursively process depending packages too
  --install  = true     install all updated packages
  --link     = false    link and install all updated binaries
  --rollback = true     automatically roll back failed upgrade
  --test     = true     test all updated packages

The prescribe command updates the repository to the named tag or
revision.  The <repo> can be any piece of the repository root path, as long as
it is unique.  The <tag> is anything understood by the underlying version
control system as a commit, usually a tag, branch, or commit.

After updating, prescribe will test, build, and the install each package
in the updated repository.  These steps can be disabled via flags such as
"rx prescribe --test=false repo tag".

By default, this will not link and install affected binaries; to turn this
behavior on, see the --link option.

*/
package main
