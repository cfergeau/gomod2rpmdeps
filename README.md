`gomod2rpmdeps` converts a list of go modules with their pseudoversion to a
list of virtual Requires suitable for use in a spec file. The purpose of this
is to make it easier to automate detection of which packages need to be updated
when a security vulnerability is reported in one of its dependencies.

To make use of it, simply run `gomod2rpmdeps` in the top-level directory of the
source tree you want to generate information for.

To build it, simply run `go build ./cmd/gomod2rpmdeps`.

For bugs or improvements, file
[issues](https://github.com/cfergeau/gomod2rpmdeps/issues/new) or
[PRs](https://github.com/cfergeau/gomod2rpmdeps/compare) :)
    

It currently converts from 'go mod vendor -v' output:
```    
# github.com/code-ready/machine v0.0.0-20210122113819-281ccfbb4566
github.com/code-ready/machine/drivers/libvirt
github.com/code-ready/machine/libmachine/drivers
github.com/code-ready/machine/libmachine/drivers/plugin
github.com/code-ready/machine/libmachine/drivers/plugin/localbinary
github.com/code-ready/machine/libmachine/drivers/rpc
github.com/code-ready/machine/libmachine/state
github.com/code-ready/machine/libmachine/version
# github.com/davecgh/go-spew v1.1.1
github.com/davecgh/go-spew/spew
# github.com/libvirt/libvirt-go v3.4.0+incompatible
github.com/libvirt/libvirt-go
# github.com/libvirt/libvirt-go-xml v6.8.0+incompatible
github.com/libvirt/libvirt-go-xml
# github.com/pmezard/go-difflib v1.0.0
github.com/pmezard/go-difflib/difflib
# github.com/sirupsen/logrus v1.7.0
github.com/sirupsen/logrus
# github.com/stretchr/testify v1.7.0
github.com/stretchr/testify/assert
# golang.org/x/sys v0.0.0-20191026070338-33540a1f6037
golang.org/x/sys/unix
golang.org/x/sys/windows
# gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
gopkg.in/yaml.v3
```
    
to a format valid for a RPM spec file:
    
```
Provides: bundled(golang(github.com/code-ready/machine)) = 0.0.0-0.20210122git281ccfbb4566
Provides: bundled(golang(github.com/davecgh/go-spew)) = 1.1.1
Provides: bundled(golang(github.com/libvirt/libvirt-go)) = 3.4.0
Provides: bundled(golang(github.com/libvirt/libvirt-go-xml)) = 6.8.0
Provides: bundled(golang(github.com/pmezard/go-difflib)) = 1.0.0
Provides: bundled(golang(github.com/sirupsen/logrus)) = 1.7.0
Provides: bundled(golang(github.com/stretchr/testify)) = 1.7.0
Provides: bundled(golang(golang.org/x/sys)) = 0.0.0-0.20191026git33540a1f6037
Provides: bundled(golang(gopkg.in/yaml.v3)) = 3.0.0-0.20210107git496545a6307b
```

## Implementation notes

Go module pseudoversions are described in
https://golang.org/doc/modules/version-numbers

Fedora versioning guidelines can be found there:
https://docs.fedoraproject.org/en-US/packaging-guidelines/Versioning/
and recommandations for handling bundled source is documented in
https://fedoraproject.org/wiki/Bundled_Libraries?rd=Packaging:Bundled_Libraries#Requirement_if_you_bundle


The running of `go mod vendor -v` is abstracted through an interface, as
there are multiple ways of getting lists of modules used (go list -m all, go
mod vendor -v, fedora's golist, cat go.mod, ...), so it's good to have some
flexibility in how this information is gathered.


`go list -m all` lists the modules which are needed by tests in the
vendored modules, `gomod2rpmdeps` is not using it since it's not clear if
`bundled()` annotation is needed for test-only dependencies.


The `go.mod` file only lists the modules which are directly used
by the codebase, but it misses some of the modules indirectly used by these
modules.

One drawback of using `go mod vendor -v` is that it can make changes to the
`vendor/` directory if it's not up to date or if it's non-existent.
Making use of `vendor/modules.txt` hasn't been investigated.
