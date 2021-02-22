package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var modVersions = map[string]goModuleInfo{
	"# github.com/spf13/cobra v1.1.1":                                                                                       goModuleInfo{name: "github.com/spf13/cobra", pseudoVersion: "v1.1.1"},
	"# github.com/stretchr/testify v1.3.0":                                                                                  goModuleInfo{name: "github.com/stretchr/testify", pseudoVersion: "v1.3.0"},
	"# gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b":                                                                 goModuleInfo{name: "gopkg.in/yaml.v3", pseudoVersion: "v3.0.0-20210107192922-496545a6307b"},
	"# github.com/libvirt/libvirt-go-xml v6.8.0+incompatible":                                                               goModuleInfo{name: "github.com/libvirt/libvirt-go-xml", pseudoVersion: "v6.8.0+incompatible"},
	"# github.com/containers/image => github.com/openshift/containers-image v0.0.0-20190130162819-76de87591e9d":             goModuleInfo{name: "github.com/openshift/containers-image", pseudoVersion: "v0.0.0-20190130162819-76de87591e9d"},
	"# k8s.io/client-go v0.19.0 => github.com/openshift/kubernetes-client-go v1.20.0-alpha.0.0.20200922142336-4700daee7399": goModuleInfo{name: "github.com/openshift/kubernetes-client-go", pseudoVersion: "v1.20.0-alpha.0.0.20200922142336-4700daee7399"},
	"# github.com/docker/docker v1.13.1 => github.com/docker/docker v1.4.2-0.20191121165722-d1d5f6476656":                   goModuleInfo{name: "github.com/docker/docker", pseudoVersion: "v1.4.2-0.20191121165722-d1d5f6476656"},

	//"## explicit":                          goModuleInfo{name: "", pseudoVersion: ""},
}

func TestParseModuleLine(t *testing.T) {
	for line, expectedModInfo := range modVersions {
		modInfo, err := parseModuleLine(line)
		require.NoError(t, err)
		require.Equal(t, expectedModInfo, modInfo)
	}
}

var pseudoVersions = map[string]string{
	"v1.1.1":                             "1.1.1",
	"v3.0.0-20210107192922-496545a6307b": "3.0.0-0.20210107git496545a6307b",
	"v6.8.0+incompatible":                "6.8.0",
	"v1.20.0-alpha.0.0.20200922142336-4700daee7399":              "1.20.0-0.alpha.20200922git4700daee7399",
	"v1.4.2-0.20191121165722-d1d5f6476656":                       "1.4.2-0.20191121gitd1d5f6476656",
	"v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible": "17.12.0-0.ce.20200309gitaa6a9891b09c",
	"v1.0.0-rc1": "1.0.0-0.rc1",
}

func TestPseudoVersionToRpmVersion(t *testing.T) {
	for pseudoVersion, expectedRpmVersion := range pseudoVersions {
		rpmVersion, err := pseudoVersionToRpmVersion(pseudoVersion)
		require.NoError(t, err)
		require.Equal(t, expectedRpmVersion, rpmVersion)
	}
}
