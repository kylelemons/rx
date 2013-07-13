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
  --autosave = true           Automatically save dependency graph (disable for concurrent runs)
  --max-age  = 1h0m0s         Nominal amount of time before a rescan is done
  --rescan   = false          Force a rescan of repositories
  --rxdir    = "$HOME/.rx"    Directory in which to save state
  -v         = false          Turn on verbose logging

Commands:
    help       Help on the rx command and subcommands.
    list       List recognized repositories.
    tags       List known repository tags.
    prescribe  Update the repository to the given tag/rev.
    cabinet    Save, list, or restore dependency snapshots.
    checkpoint Save, list, or restore global repository version snapshots.

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
  -f     = ""       List output format
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
  -f     = ""       tags output format
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

Cabinet Command

Save, list, or restore dependency snapshots.

Usage:
    rx cabinet <repo> [<id>]

Options:
  --build = false    create a new cabinet
  --dump  = false    list the contents of the specified cabinet
  --list  = true     list matching cabinet files (the default)
  --open  = false    open the specified cabinet
  --test  = true     test package before saving and after loading cabinet

The cabinet command saves dependency information for the given
repository in a file within it.  The <repo> can be any piece of the
repository root path, as long as it is unique.  The <tag> is anything
understood by the underlying version control system as a commit, usually a
tag, branch, or commit.  When saving, the <id> will override the default
(based on the date) and when opening the <id> is the ID (or a unique
substring) of the cabinet to open.

Cabinets are currently stored in the repository itself named as follows:
    .rx/cabinet-YYYYMMDD-HHMMSS

If the <id> is specified for --build, it will override the date-based portion
of the cabinet name.  If a cabinet already exists with the name (even if it is
a date-based name) it will not be overwritten.

Unless --test=false is specified, the packages in the repository will be tested
before a cabinet is created and after a cabinet is restored.  If the test
before creation fails, the cabinet will not be created.  If the test after
restoration fails, the repositories will be reverted to their original
revisions.

The cabinet format is still in flux.

Checkpoint Command

Save, list, or restore global repository version snapshots.

Usage:
    rx checkpoint 

Options:
  --apply = 0        apply the specified checkpoint
  --list  = false    list checkpoints
  -n      = 15       number of checkpoints to list (0 for all)
  --save  = ""       save a new checkpoint with the given comment

The checkpoint command is similar to the cabinet command, except
that it has global scope and does not run tests when saving or applying.

Checkpoints are intended as lightweight ways to save state and to share state
among multiple developers sharing an $RX_DIR.  Checkpoints are created with a
comment and a global, sequential ID which can be used to retrieve or delete it.

*/
package main
