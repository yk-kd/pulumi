//nolint:goconst
package lifecycletest

import (
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"

	. "github.com/pulumi/pulumi/pkg/v3/engine"
	"github.com/pulumi/pulumi/pkg/v3/resource/deploy/deploytest"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
)

func TestComponentLifecycle(t *testing.T) {
	t.Parallel()

	loaders := []*deploytest.ProviderLoader{
		deploytest.NewProviderLoader("foo", semver.MustParse("1.0.0"), func() (plugin.Provider, error) {
			return &deploytest.Provider{}, nil
		}),
	}

	// Equivalent to:
	//
	// class First(ComponentResource):
	//     def __init__(self, name: str, opts: Optional[ResourceOptions]=None):
	//         super().__init__("pkg:index:first", name, {}, opts)
	//
	// class Second(ComponentResource):
	//     def __init__(self, name: str, opts: Optional[ResourceOptions]=None):
	//         super().__init__("pkg:index:second", name, {}, opts)
	//
	//         RandomString("str",
	//             length=10,
	//             opts=ResourceOptions(parent=self))
	//
	// first = First("first")
	// second = Second("second",
	//     opts=ResourceOptions(parent=first, depends_on=[first]))
	//
	program := deploytest.NewLanguageRuntime(func(_ plugin.RunInfo, monitor *deploytest.ResourceMonitor) error {
		firstURN, _, _, err := monitor.RegisterResource("pkg:index:first", "myfirst", false)
		assert.NoError(t, err)
		secondURN, _, _, err := monitor.RegisterResource("pkg:index:second", "mysecond", false, deploytest.ResourceOptions{
			Parent:       firstURN,
			Dependencies: []resource.URN{firstURN},
		})
		assert.NoError(t, err)
		_, _, _, err = monitor.RegisterResource("foo:index:child", "mychild", true, deploytest.ResourceOptions{
			Parent: secondURN,
		})
		assert.NoError(t, err)
		return nil
	})
	host := deploytest.NewPluginHost(nil, nil, program, loaders...)

	p := &TestPlan{
		Options: UpdateOptions{Host: host},
		Steps:   MakeBasicLifecycleSteps(t, 4),
	}
	p.Run(t, nil)
}
