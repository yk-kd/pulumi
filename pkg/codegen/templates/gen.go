// Copyright 2016-2020, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Pulling out some of the repeated strings tokens into constants would harm readability, so we just ignore the
// goconst linter's warning.
//
// nolint: lll, goconst
package templates

import (
	"fmt"
	"html/template"
	"path"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"

	"github.com/pulumi/pulumi/pkg/v2/codegen/schema"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/contract"
	"github.com/pulumi/pulumi/sdk/v2/go/common/workspace"
)

var (
	supportedLanguages = []string{"csharp", "go", "nodejs", "python"}
	snippetLanguages   = []string{"csharp", "go", "python", "typescript"}
	templates          *template.Template

	// langModuleNameLookup is a map of module name to its language-specific
	// name.
	langModuleNameLookup map[string]string
	// titleLookup is a map to map module package name to the desired display name
	// for display in the TOC menu under API Reference.
	titleLookup = map[string]string{
		"aiven":        "Aiven",
		"alicloud":     "AliCloud",
		"auth0":        "Auth0",
		"aws":          "AWS",
		"azure":        "Azure",
		"azuread":      "Azure AD",
		"azuredevops":  "Azure DevOps",
		"cloudamqp":    "CloudAMQP",
		"cloudflare":   "Cloudflare",
		"consul":       "Consul",
		"datadog":      "Datadog",
		"digitalocean": "DigitalOcean",
		"dnsimple":     "DNSimple",
		"docker":       "Docker",
		"f5bigip":      "f5 BIG-IP",
		"fastly":       "Fastly",
		"gcp":          "GCP",
		"github":       "GitHub",
		"gitlab":       "GitLab",
		"hcloud":       "Hetzner Cloud",
		"kafka":        "Kafka",
		"keycloak":     "Keycloak",
		"kong":         "Kong",
		"kubernetes":   "Kubernetes",
		"linode":       "Linode",
		"mailgun":      "Mailgun",
		"mongodbatlas": "MongoDB Atlas",
		"mysql":        "MySQL",
		"newrelic":     "New Relic",
		"ns1":          "NS1",
		"okta":         "Okta",
		"openstack":    "Open Stack",
		"packet":       "Packet",
		"pagerduty":    "PagerDuty",
		"postgresql":   "PostgreSQL",
		"rabbitmq":     "RabbitMQ",
		"rancher2":     "Rancher 2",
		"random":       "Random",
		"signalfx":     "SignalFx",
		"spotinst":     "Spotinst",
		"tls":          "TLS",
		"vault":        "Vault",
		"vsphere":      "vSphere",
	}
)

// tokenToName returns the resource name from a Pulumi token.
func tokenToName(tok string) string {
	components := strings.Split(tok, ":")
	contract.Assertf(len(components) == 3, "malformed token %v", tok)
	return components[2]
}

type fs map[string][]byte

func (fs fs) add(path string, contents []byte) {
	_, has := fs[path]
	contract.Assertf(!has, "duplicate file: %s", path)
	fs[path] = contents
}

func language(lang string) string {
	switch lang {
	case "csharp":
		return "C#"
	case "go":
		return "Go"
	case "python":
		return "Python"
	case "typescript":
		return "TypeScript"
	default:
		contract.Failf("Unexpected lang: %s", lang)
		return lang
	}
}

var spaceRE = regexp.MustCompile(`\s+`)
var illegalRE = regexp.MustCompile(`[^a-zA-Z0-9 -]+`)

func gen(tool string, pkg *schema.Package, fs fs, token, comment string) error {
	dir := pkg.Name
	modName := pkg.TokenToModule(token)
	if modName != "" {
		dir = fmt.Sprintf("%s.%s", dir, modName)
	}
	dir = fmt.Sprintf("%s.%s", dir, tokenToName(token))

	docInfo := decomposeDocstring(comment)

	for _, example := range docInfo.examples {
		for lang, code := range example.Snippets {
			snippetDir := dir
			title := strings.ToLower(example.Title)
			title = strings.TrimPrefix(title, "### ")
			title = strings.TrimSpace(title)
			title = illegalRE.ReplaceAllString(title, "")
			title = spaceRE.ReplaceAllString(title, "-")
			if title != "" {
				snippetDir = fmt.Sprintf("%s.%s", snippetDir, title)
			}
			snippetDir = fmt.Sprintf("%s-%s", snippetDir, lang)

			language := language(lang)

			// TODO make the descriptions more consistent.
			description := strings.TrimPrefix(example.Title, "### ")
			description = strings.TrimSpace(description)
			if description == "" {
				description = fmt.Sprintf("An AWS %s %s Pulumi program", tokenToName(token), language)
			}

			code = strings.TrimSpace(code)
			code = strings.TrimPrefix(code, fmt.Sprintf("```%s\n", lang))
			code = strings.TrimSuffix(code, "```")
			code = fmt.Sprintf("%s\n", code)

			switch lang {
			case "csharp":
				genCSharpTemplate(fs, snippetDir, description, code)
			case "go":
				genGoTemplate(fs, snippetDir, description, code)
			case "python":
				genPythonTemplate(fs, snippetDir, description, code)
			case "typescript":
				genTypeScriptTemplate(fs, snippetDir, description, code)
			default:
				contract.Failf("Unexpected lang: %s", lang)
			}
		}
	}

	return nil
}

func generateFromSchemaPackage(tool string, pkg *schema.Package, fs fs) error {
	// TODO does the provider have examples?
	for _, r := range pkg.Resources {
		if err := gen(tool, pkg, fs, r.Token, r.Comment); err != nil {
			return err
		}
	}

	for _, f := range pkg.Functions {
		if err := gen(tool, pkg, fs, f.Token, f.Comment); err != nil {
			return err
		}
	}

	return nil
}

// GeneratePackage generates the templates for each resource/function example.
func GeneratePackage(tool string, pkg *schema.Package) (map[string][]byte, error) {
	defer glog.Flush()

	files := fs{}
	if err := generateFromSchemaPackage(tool, pkg, files); err != nil {
		return nil, err
	}

	return files, nil
}

func genCSharpTemplate(fs fs, dir, description, code string) {
	// TODO don't hardcode AWS dep
	csproj := `<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>netcoreapp3.1</TargetFramework>
    <Nullable>enable</Nullable>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Pulumi.Aws" Version="2.*" />
  </ItemGroup>

</Project>
`
	fs.add(path.Join(dir, "${PROJECT}.csproj"), []byte(csproj))

	gitignore := `## Ignore Visual Studio temporary files, build results, and
## files generated by popular Visual Studio add-ons.
##
## Get latest from https://github.com/github/gitignore/blob/master/VisualStudio.gitignore

# User-specific files
*.rsuser
*.suo
*.user
*.userosscache
*.sln.docstates

# User-specific files (MonoDevelop/Xamarin Studio)
*.userprefs

# Mono auto generated files
mono_crash.*

# Build results
[Dd]ebug/
[Dd]ebugPublic/
[Rr]elease/
[Rr]eleases/
x64/
x86/
[Aa][Rr][Mm]/
[Aa][Rr][Mm]64/
bld/
[Bb]in/
[Oo]bj/
[Ll]og/
[Ll]ogs/

# Visual Studio 2015/2017 cache/options directory
.vs/
# Uncomment if you have tasks that create the project's static files in wwwroot
#wwwroot/

# Visual Studio 2017 auto generated files
Generated\ Files/

# MSTest test Results
[Tt]est[Rr]esult*/
[Bb]uild[Ll]og.*

# NUnit
*.VisualState.xml
TestResult.xml
nunit-*.xml

# Build Results of an ATL Project
[Dd]ebugPS/
[Rr]eleasePS/
dlldata.c

# Benchmark Results
BenchmarkDotNet.Artifacts/

# .NET Core
project.lock.json
project.fragment.lock.json
artifacts/

# StyleCop
StyleCopReport.xml

# Files built by Visual Studio
*_i.c
*_p.c
*_h.h
*.ilk
*.meta
*.obj
*.iobj
*.pch
*.pdb
*.ipdb
*.pgc
*.pgd
*.rsp
*.sbr
*.tlb
*.tli
*.tlh
*.tmp
*.tmp_proj
*_wpftmp.csproj
*.log
*.vspscc
*.vssscc
.builds
*.pidb
*.svclog
*.scc

# Chutzpah Test files
_Chutzpah*

# Visual C++ cache files
ipch/
*.aps
*.ncb
*.opendb
*.opensdf
*.sdf
*.cachefile
*.VC.db
*.VC.VC.opendb

# Visual Studio profiler
*.psess
*.vsp
*.vspx
*.sap

# Visual Studio Trace Files
*.e2e

# TFS 2012 Local Workspace
$tf/

# Guidance Automation Toolkit
*.gpState

# ReSharper is a .NET coding add-in
_ReSharper*/
*.[Rr]e[Ss]harper
*.DotSettings.user

# JustCode is a .NET coding add-in
.JustCode

# TeamCity is a build add-in
_TeamCity*

# DotCover is a Code Coverage Tool
*.dotCover

# AxoCover is a Code Coverage Tool
.axoCover/*
!.axoCover/settings.json

# Visual Studio code coverage results
*.coverage
*.coveragexml

# NCrunch
_NCrunch_*
.*crunch*.local.xml
nCrunchTemp_*

# MightyMoose
*.mm.*
AutoTest.Net/

# Web workbench (sass)
.sass-cache/

# Installshield output folder
[Ee]xpress/

# DocProject is a documentation generator add-in
DocProject/buildhelp/
DocProject/Help/*.HxT
DocProject/Help/*.HxC
DocProject/Help/*.hhc
DocProject/Help/*.hhk
DocProject/Help/*.hhp
DocProject/Help/Html2
DocProject/Help/html

# Click-Once directory
publish/

# Publish Web Output
*.[Pp]ublish.xml
*.azurePubxml
# Note: Comment the next line if you want to checkin your web deploy settings,
# but database connection strings (with potential passwords) will be unencrypted
*.pubxml
*.publishproj

# Microsoft Azure Web App publish settings. Comment the next line if you want to
# checkin your Azure Web App publish settings, but sensitive information contained
# in these scripts will be unencrypted
PublishScripts/

# NuGet Packages
*.nupkg
# NuGet Symbol Packages
*.snupkg
# The packages folder can be ignored because of Package Restore
**/[Pp]ackages/*
# except build/, which is used as an MSBuild target.
!**/[Pp]ackages/build/
# Uncomment if necessary however generally it will be regenerated when needed
#!**/[Pp]ackages/repositories.config
# NuGet v3's project.json files produces more ignorable files
*.nuget.props
*.nuget.targets

# Microsoft Azure Build Output
csx/
*.build.csdef

# Microsoft Azure Emulator
ecf/
rcf/

# Windows Store app package directories and files
AppPackages/
BundleArtifacts/
Package.StoreAssociation.xml
_pkginfo.txt
*.appx
*.appxbundle
*.appxupload

# Visual Studio cache files
# files ending in .cache can be ignored
*.[Cc]ache
# but keep track of directories ending in .cache
!?*.[Cc]ache/

# Others
ClientBin/
~$*
*~
*.dbmdl
*.dbproj.schemaview
*.jfm
*.pfx
*.publishsettings
orleans.codegen.cs

# Including strong name files can present a security risk
# (https://github.com/github/gitignore/pull/2483#issue-259490424)
#*.snk

# Since there are multiple workflows, uncomment next line to ignore bower_components
# (https://github.com/github/gitignore/pull/1529#issuecomment-104372622)
#bower_components/

# RIA/Silverlight projects
Generated_Code/

# Backup & report files from converting an old project file
# to a newer Visual Studio version. Backup files are not needed,
# because we have git ;-)
_UpgradeReport_Files/
Backup*/
UpgradeLog*.XML
UpgradeLog*.htm
ServiceFabricBackup/
*.rptproj.bak

# SQL Server files
*.mdf
*.ldf
*.ndf

# Business Intelligence projects
*.rdl.data
*.bim.layout
*.bim_*.settings
*.rptproj.rsuser
*- [Bb]ackup.rdl
*- [Bb]ackup ([0-9]).rdl
*- [Bb]ackup ([0-9][0-9]).rdl

# Microsoft Fakes
FakesAssemblies/

# GhostDoc plugin setting file
*.GhostDoc.xml

# Node.js Tools for Visual Studio
.ntvs_analysis.dat
node_modules/

# Visual Studio 6 build log
*.plg

# Visual Studio 6 workspace options file
*.opt

# Visual Studio 6 auto-generated workspace file (contains which files were open etc.)
*.vbw

# Visual Studio LightSwitch build output
**/*.HTMLClient/GeneratedArtifacts
**/*.DesktopClient/GeneratedArtifacts
**/*.DesktopClient/ModelManifest.xml
**/*.Server/GeneratedArtifacts
**/*.Server/ModelManifest.xml
_Pvt_Extensions

# Paket dependency manager
.paket/paket.exe
paket-files/

# FAKE - F# Make
.fake/

# CodeRush personal settings
.cr/personal

# Python Tools for Visual Studio (PTVS)
__pycache__/
*.pyc

# Cake - Uncomment if you are using it
# tools/**
# !tools/packages.config

# Tabs Studio
*.tss

# Telerik's JustMock configuration file
*.jmconfig

# BizTalk build output
*.btp.cs
*.btm.cs
*.odx.cs
*.xsd.cs

# OpenCover UI analysis results
OpenCover/

# Azure Stream Analytics local run output
ASALocalRun/

# MSBuild Binary and Structured Log
*.binlog

# NVidia Nsight GPU debugger configuration file
*.nvuser

# MFractors (Xamarin productivity tool) working folder
.mfractor/

# Local History for Visual Studio
.localhistory/

# BeatPulse healthcheck temp database
healthchecksdb

# Backup folder for Package Reference Convert tool in Visual Studio 2017
MigrationBackup/

# Ionide (cross platform F# VS Code tools) working folder
.ionide/
`
	fs.add(path.Join(dir, ".gitignore"), []byte(gitignore))

	program := `using System.Threading.Tasks;
using Pulumi;

class Program
{
    static Task<int> Main() => Deployment.RunAsync<MyStack>();
}
`
	fs.add(path.Join(dir, "Program.cs"), []byte(program))

	fs.add(path.Join(dir, "MyStack.cs"), []byte(code))

	fs.add(path.Join(dir, "Pulumi.yaml"), genProject("dotnet", description))
}

func genGoTemplate(fs fs, dir, description, code string) {
	// TODO don't hardcode AWS dep
	fs.add(path.Join(dir, "go.mod"), []byte(`module ${PROJECT}

go 1.14

require (
	github.com/pulumi/pulumi-aws/sdk/v2 v2.0.0
	github.com/pulumi/pulumi/sdk/v2 v2.0.0
)
`))
	fs.add(path.Join(dir, "main.go"), []byte(code))
	fs.add(path.Join(dir, "Pulumi.yaml"), genProject("go", description))
}

func genPythonTemplate(fs fs, dir, description, code string) {
	fs.add(path.Join(dir, ".gitignore"), []byte(`*.pyc
venv/
`))
	// TODO don't hardcode AWS dep
	fs.add(path.Join(dir, "requirements.txt"), []byte(`pulumi>=2.0.0,<3.0.0
pulumi-aws>=2.0.0,<3.0.0
`))
	fs.add(path.Join(dir, "__main__.py"), []byte(code))
	fs.add(path.Join(dir, "Pulumi.yaml"), genProject("python", description))
}

func genTypeScriptTemplate(fs fs, dir, description, code string) {
	fs.add(path.Join(dir, ".gitignore"), []byte(`/bin/
/node_modules/
`))
	// TODO don't hardcode AWS dep
	fs.add(path.Join(dir, "package.json"), []byte(`{
    "name": "${PROJECT}",
    "devDependencies": {
        "@types/node": "^10.0.0"
    },
    "dependencies": {
        "@pulumi/pulumi": "^2.0.0",
        "@pulumi/aws": "^2.0.0",
        "@pulumi/awsx": "^0.20.0"
    }
}
`))
	fs.add(path.Join(dir, "tsconfig.json"), []byte(`{
    "compilerOptions": {
        "strict": true,
        "outDir": "bin",
        "target": "es2016",
        "module": "commonjs",
        "moduleResolution": "node",
        "sourceMap": true,
        "experimentalDecorators": true,
        "pretty": true,
        "noFallthroughCasesInSwitch": true,
        "noImplicitReturns": true,
        "forceConsistentCasingInFileNames": true
    },
    "files": [
        "index.ts"
    ]
}
`))
	fs.add(path.Join(dir, "index.ts"), []byte(code))
	fs.add(path.Join(dir, "Pulumi.yaml"), genProject("nodejs", description))
}

func genProject(runtime, description string) []byte {
	desc := "${DESCRIPTION}"
	proj := &workspace.Project{
		Name:        "${PROJECT}",
		Description: &desc,
		Runtime:     workspace.NewProjectRuntimeInfo(runtime, nil),
		Template: &workspace.ProjectTemplate{
			Description: description,
			Config: map[string]workspace.ProjectTemplateConfigValue{
				"aws:region": {
					Description: "The AWS region to deploy into",
					Default:     "us-east-1",
				},
			},
		},
	}
	bytes, err := yaml.Marshal(proj)
	contract.AssertNoError(err)
	return bytes
}
