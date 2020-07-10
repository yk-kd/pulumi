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

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/alecthomas/chroma/quick"
	"github.com/pulumi/pulumi/pkg/v2/backend/display"
	"github.com/pulumi/pulumi/sdk/v2/go/common/diag/colors"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/contract"
	"github.com/spf13/cobra"
	survey "gopkg.in/AlecAivazis/survey.v1"
	surveycore "gopkg.in/AlecAivazis/survey.v1/core"
)

type cloudMap = map[string][]modInfo

type modInfo struct {
	name  string
	types []typeInfo
}

type typeInfo struct {
	name     string
	examples []exampleInfo
}

type exampleInfo struct {
	name     string
	snippets []snippetInfo
}

type snippetInfo struct {
	lang string
	dir  string
}

func (s snippetInfo) displayName() string {
	switch s.lang {
	case "csharp":
		return "C#"
	case "go":
		return "Go"
	case "python":
		return "Python"
	case "typescript":
		return "TypeScript"
	default:
		contract.Failf("Unexpected lang: %s", s.lang)
		return s.lang
	}
}

func (s snippetInfo) codeFileName() string {
	switch s.lang {
	case "csharp":
		return "MyStack.cs"
	case "go":
		return "main.go"
	case "python":
		return "__main__.py"
	case "typescript":
		return "index.ts"
	default:
		contract.Failf("Unexpected lang: %s", s.lang)
		return s.lang
	}
}

func (s snippetInfo) readCode() string {
	path := filepath.Join(s.dir, s.codeFileName())
	bytes, err := ioutil.ReadFile(path)
	contract.AssertNoError(err)
	return string(bytes)
}

const localDir = "/Users/justin/go/src/github.com/justinvp/templates-aws"

func newExampleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "example",
		Short: "Display examples",
		// TODO provide a better detailed description.
		Long: "Display examples",
		Args: cmdutil.NoArgs,
		Run: cmdutil.RunFunc(func(cmd *cobra.Command, args []string) error {
			opts := display.Options{
				Color: cmdutil.GetGlobalColorization(),
			}

			// TODO parse the template dir for real, instead of hardcoding.
			// files, err := ioutil.ReadDir(localDir)
			// if err != nil {
			// 	return err
			// }

			// for _, f := range files {
			// 	if !f.IsDir() {
			// 		continue
			// 	}

			// 	// TODO parse the template directory.
			// }

			clouds := cloudMap{
				"AWS": []modInfo{
					{
						name: "s3",
						types: []typeInfo{
							{
								name: "Bucket",
								examples: []exampleInfo{
									{
										name: "Private Bucket w/ Tags",
										snippets: []snippetInfo{
											{lang: "typescript", dir: filepath.Join(localDir, "/aws.s3.Bucket.private-bucket-w-tags-typescript")},
											{lang: "go", dir: filepath.Join(localDir, "/aws.s3.Bucket.private-bucket-w-tags-go")},
											{lang: "python", dir: filepath.Join(localDir, "/aws.s3.Bucket.private-bucket-w-tags-python")},
											{lang: "csharp", dir: filepath.Join(localDir, "/aws.s3.Bucket.private-bucket-w-tags-csharp")},
										},
									},
									{
										name: "Using CORS",
										snippets: []snippetInfo{
											{lang: "typescript", dir: filepath.Join(localDir, "/aws.s3.Bucket.using-cors-typescript")},
											{lang: "go", dir: filepath.Join(localDir, "/aws.s3.Bucket.using-cors-go")},
											{lang: "python", dir: filepath.Join(localDir, "/aws.s3.Bucket.using-cors-python")},
											{lang: "csharp", dir: filepath.Join(localDir, "/aws.s3.Bucket.using-cors-csharp")},
										},
									},
								},
							},
							{
								name: "Inventory",
								examples: []exampleInfo{
									{
										name: "Add Inventory Configuration",
										snippets: []snippetInfo{
											{lang: "typescript", dir: filepath.Join(localDir, "/aws.s3.Inventory.add-inventory-configuration-typescript")},
											{lang: "go", dir: filepath.Join(localDir, "/aws.s3.Inventory.add-inventory-configuration-go")},
											{lang: "python", dir: filepath.Join(localDir, "/aws.s3.Inventory.add-inventory-configuration-python")},
											{lang: "csharp", dir: filepath.Join(localDir, "/aws.s3.Inventory.add-inventory-configuration-csharp")},
										},
									},
								},
							},
							{name: "Access Point"},
						},
					},
					{name: "accessanalyzer"},
					{name: "acm"},
					{name: "acmpca"},
					{name: "alb"},
					{name: "apigateway"},
					{name: "apigatewayv2"},
					{name: "appautoscaling"},
					{name: "applicationloadbalancing"},
					{name: "appmesh"},
					{name: "appsync"},
					{name: "athena"},
					{name: "autoscaling"},
					{name: "backup"},
					{name: "batch"},
					{name: "budgets"},
					{name: "cfg"},
					{name: "cloud9"},
					{name: "cloudformation"},
					{name: "cloudfront"},
					{name: "cloudhsmv2"},
					{name: "cloudtrail"},
					{name: "cloudwatch"},
					{name: "codebuild"},
					{name: "codecommit"},
					{name: "codedeploy"},
					{name: "codepipeline"},
					{name: "codestarnotifications"},
					{name: "cognito"},
					{name: "cur"},
					{name: "datapipeline"},
					{name: "datasync"},
					{name: "dax"},
					{name: "devicefarm"},
					{name: "directconnect"},
					{name: "directoryservice"},
					{name: "dlm"},
					{name: "dms"},
					{name: "docdb"},
					{name: "dynamodb"},
					{name: "ebs"},
					{name: "ec2"},
					{name: "ec2clientvpn"},
					{name: "ec2transitgateway"},
					{name: "ecr"},
					{name: "ecs"},
					{name: "efs"},
					{name: "eks"},
					{name: "elasticache"},
					{name: "elasticbeanstalk"},
					{name: "elasticloadbalancing"},
					{name: "elasticloadbalancingv2"},
					{name: "elasticsearch"},
					{name: "elastictranscoder"},
					{name: "elb"},
					{name: "emr"},
					{name: "fms"},
					{name: "fsx"},
					{name: "gamelift"},
					{name: "glacier"},
					{name: "globalaccelerator"},
					{name: "glue"},
					{name: "guardduty"},
					{name: "iam"},
					{name: "inspector"},
					{name: "iot"},
					{name: "kinesis"},
					{name: "kms"},
					{name: "lambda"},
					{name: "lb"},
					{name: "licensemanager"},
					{name: "lightsail"},
					{name: "macie"},
					{name: "mediaconvert"},
					{name: "mediapackage"},
					{name: "mediastore"},
					{name: "mq"},
					{name: "msk"},
					{name: "neptune"},
					{name: "opsworks"},
					{name: "organizations"},
					{name: "outposts"},
					{name: "pinpoint"},
					{name: "pricing"},
					{name: "qldb"},
					{name: "quicksight"},
					{name: "ram"},
					{name: "rds"},
					{name: "redshift"},
					{name: "resourcegroups"},
					{name: "route53"},
					{name: "sagemaker"},
					{name: "secretsmanager"},
					{name: "securityhub"},
					{name: "servicecatalog"},
					{name: "servicediscovery"},
					{name: "servicequotas"},
					{name: "ses"},
					{name: "sfn"},
					{name: "shield"},
					{name: "simpledb"},
					{name: "sns"},
					{name: "sqs"},
					{name: "ssm"},
					{name: "storagegateway"},
					{name: "swf"},
					{name: "transfer"},
					{name: "waf"},
					{name: "wafregional"},
					{name: "wafv2"},
					{name: "worklink"},
					{name: "workspaces"},
					{name: "xray"},
				},
				"Azure":        []modInfo{},
				"Google Cloud": []modInfo{},
				"Kubernetes":   []modInfo{},
			}

			var snippet *snippetInfo
			var create bool

			for {
				mods := chooseCloud(clouds, opts)
				for {
					mod := chooseMod(mods, opts)
					if mod == nil {
						break
					}
					for {
						typ := chooseType(mod.types, opts)
						if typ == nil {
							break
						}
						for {
							var example *exampleInfo
							if len(typ.examples) <= 1 {
								example = &typ.examples[0]
							} else {
								example = chooseExample(typ.examples, opts)
							}
							if example == nil {
								break
							}
							for {
								snippet, create = chooseSnippet(example.snippets, snippet, opts)
								if snippet == nil {
									break
								}

								if create {
									return runNew(newArgs{
										interactive:       cmdutil.Interactive(),
										prompt:            promptForValue,
										secretsProvider:   "default",
										templateNameOrURL: snippet.dir,
									})
								}

								fmt.Println()
								fmt.Println()
								const style = "paraiso-dark" // "monokai"
								if err := quick.Highlight(os.Stdout, snippet.readCode(), snippet.lang, "terminal256", style); err != nil {
									return err
								}
								fmt.Println()
								fmt.Println()
							}
						}
					}
				}
			}
		}),
	}

	return cmd
}

const back = "Go back"
const exit = "Exit"

// TODO clean up below to share more code.

func chooseCloud(clouds cloudMap, opts display.Options) []modInfo {
	// Customize the prompt a little bit (and disable color since it doesn't match our scheme).
	surveycore.DisableColor = true
	surveycore.QuestionIcon = ""
	surveycore.SelectFocusIcon = opts.Color.Colorize(colors.BrightGreen + ">" + colors.Reset)
	message := "\rChoose a provider:"
	message = opts.Color.Colorize(colors.SpecPrompt + message + colors.Reset)

	const showOthers = "Show others"

	var options []string
	for key := range clouds {
		options = append(options, key)
	}
	sort.Strings(options)
	options = append(options, showOthers)
	options = append(options, exit)

	cmdutil.EndKeypadTransmitMode()

	for {
		var option string
		if err := survey.AskOne(&survey.Select{
			Message:  message,
			Options:  options,
			PageSize: len(options),
		}, &option, nil); err != nil || option == exit {
			os.Exit(0)
		}
		if option == showOthers {
			continue
		}
		return clouds[option]
	}
}

func chooseMod(mods []modInfo, opts display.Options) *modInfo {
	// Customize the prompt a little bit (and disable color since it doesn't match our scheme).
	surveycore.DisableColor = true
	surveycore.QuestionIcon = ""
	surveycore.SelectFocusIcon = opts.Color.Colorize(colors.BrightGreen + ">" + colors.Reset)
	message := "\rChoose a module:"
	message = opts.Color.Colorize(colors.SpecPrompt + message + colors.Reset)

	var options []string
	for _, mod := range mods {
		options = append(options, mod.name)
	}
	sort.Strings(options)
	options = append(options, back)
	options = append(options, exit)

	cmdutil.EndKeypadTransmitMode()

	var option string
	if err := survey.AskOne(&survey.Select{
		Message:  message,
		Options:  options,
		PageSize: len(options),
	}, &option, nil); err != nil || option == exit {
		os.Exit(0)
	}
	if option == back {
		return nil
	}
	for _, mod := range mods {
		if mod.name == option {
			return &mod
		}
	}
	return nil
}

func chooseType(types []typeInfo, opts display.Options) *typeInfo {
	// Customize the prompt a little bit (and disable color since it doesn't match our scheme).
	surveycore.DisableColor = true
	surveycore.QuestionIcon = ""
	surveycore.SelectFocusIcon = opts.Color.Colorize(colors.BrightGreen + ">" + colors.Reset)
	message := "\rChoose a type:"
	message = opts.Color.Colorize(colors.SpecPrompt + message + colors.Reset)

	var options []string
	for _, typ := range types {
		options = append(options, typ.name)
	}
	sort.Strings(options)
	options = append(options, back)
	options = append(options, exit)

	cmdutil.EndKeypadTransmitMode()

	var option string
	if err := survey.AskOne(&survey.Select{
		Message:  message,
		Options:  options,
		PageSize: len(options),
	}, &option, nil); err != nil || option == exit {
		os.Exit(0)
	}
	if option == back {
		return nil
	}
	for _, typ := range types {
		if typ.name == option {
			return &typ
		}
	}
	return nil
}

func chooseExample(examples []exampleInfo, opts display.Options) *exampleInfo {
	// Customize the prompt a little bit (and disable color since it doesn't match our scheme).
	surveycore.DisableColor = true
	surveycore.QuestionIcon = ""
	surveycore.SelectFocusIcon = opts.Color.Colorize(colors.BrightGreen + ">" + colors.Reset)
	message := "\rChoose example:"
	message = opts.Color.Colorize(colors.SpecPrompt + message + colors.Reset)

	var options []string
	for _, example := range examples {
		options = append(options, example.name)
	}
	sort.Strings(options)
	options = append(options, back)
	options = append(options, exit)

	cmdutil.EndKeypadTransmitMode()

	var option string
	if err := survey.AskOne(&survey.Select{
		Message:  message,
		Options:  options,
		PageSize: len(options),
	}, &option, nil); err != nil || option == exit {
		os.Exit(0)
	}
	if option == back {
		return nil
	}
	for _, example := range examples {
		if example.name == option {
			return &example
		}
	}
	return nil
}

const create = "Create project"

func chooseSnippet(snippets []snippetInfo, chosen *snippetInfo, opts display.Options) (*snippetInfo, bool) {
	// Customize the prompt a little bit (and disable color since it doesn't match our scheme).
	surveycore.DisableColor = true
	surveycore.QuestionIcon = ""
	surveycore.SelectFocusIcon = opts.Color.Colorize(colors.BrightGreen + ">" + colors.Reset)
	message := "\rChoose a language:"
	message = opts.Color.Colorize(colors.SpecPrompt + message + colors.Reset)

	var options []string
	for _, snippet := range snippets {
		options = append(options, snippet.displayName())
	}
	sort.Strings(options)
	options = append(options, back)
	options = append(options, exit)

	cmdutil.EndKeypadTransmitMode()

	if chosen != nil {
		options = append([]string{create}, options...)
	}

	var option string
	if err := survey.AskOne(&survey.Select{
		Message:  message,
		Options:  options,
		PageSize: len(options),
	}, &option, nil); err != nil || option == exit {
		os.Exit(0)
	}
	if option == back {
		return nil, false
	}
	if chosen != nil && option == create {
		return chosen, true
	}
	for _, snippet := range snippets {
		if snippet.displayName() == option {
			return &snippet, false
		}
	}
	return nil, false
}
