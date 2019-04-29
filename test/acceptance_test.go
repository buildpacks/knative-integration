package test_test

import (
	"net/http"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/buildpack/imgutil"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestAcceptance(t *testing.T) {
	spec.Run(t, "acceptance", testAcceptance, spec.Report(report.Terminal{}))
}

func testAcceptance(t *testing.T, when spec.G, it spec.S) {
	var image imgutil.Image
	it.After(func() {
		cmd := exec.Command("kubectl", "delete", "build", "cnb-test-build")
		output, _ := cmd.CombinedOutput()
		println(string(output))

		cmd = exec.Command("kubectl", "delete", "buildtemplate", "buildpacks-cnb")
		output, _ = cmd.CombinedOutput()
		println(string(output))

		deleteImageTag(t)
	})

	when("basic lifecycle", func() {
		it("works", func() {
			var (
				output []byte
				cmd    *exec.Cmd
				err    error
			)
			t.Log("Apply build template")
			cmd = exec.Command("kubectl", "apply", "-f", "../cnb.yaml")
			output, err = cmd.CombinedOutput()
			println(string(output))
			AssertNil(t, err)
			AssertContains(t, string(output), "buildtemplate.build.knative.dev/buildpacks-cnb created")

			t.Log("Creating builder")
			builderName := "gcr.io/cncf-buildpacks-ci/test/builder"
			cmd = exec.Command("pack", "create-builder", builderName, "-b", "fixtures/builder.toml", "--publish")
			output, err = cmd.CombinedOutput()
			println(string(output))
			AssertNil(t, err)
			AssertContains(t, string(output), "Successfully created builder image")

			t.Log("Create a new build")
			cmd = exec.Command("kubectl", "apply", "-f", "fixtures/build.yaml")
			output, err = cmd.CombinedOutput()
			println(string(output))
			AssertNil(t, err)
			AssertContains(t, string(output), "build.build.knative.dev/cnb-test-build created")

			image, err = imgutil.NewRemoteImage("gcr.io/cncf-buildpacks-ci/test/app", authn.DefaultKeychain)
			Eventually(t, func() bool {
				result, err := image.Found()
				AssertNil(t, err)
				return result
			}, time.Second, 5*time.Minute)

			time.Sleep(10 * time.Second)
		})
	})
}

func deleteImageTag(t *testing.T) {
	reference, err := name.ParseReference("gcr.io/cncf-buildpacks-ci/test/app", name.WeakValidation)
	AssertNil(t, err)
	registry, err := name.NewRegistry("gcr.io", name.WeakValidation)
	AssertNil(t, err)
	authenticator, err := authn.DefaultKeychain.Resolve(registry)
	AssertNil(t, err)
	err = remote.Delete(reference, authenticator, http.DefaultTransport)
	AssertNil(t, err)
}

func AssertContains(t *testing.T, actual, expected string) {
	t.Helper()
	if !strings.Contains(actual, expected) {
		t.Fatalf("Expected: '%s' to contain '%s'", actual, expected)
	}
}

func AssertNil(t *testing.T, actual interface{}) {
	t.Helper()
	if !isNil(actual) {
		t.Fatalf("Expected nil: %s", actual)
	}
}

func isNil(value interface{}) bool {
	return value == nil || (reflect.TypeOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil())
}

func Eventually(t *testing.T, test func() bool, every time.Duration, timeout time.Duration) {
	t.Helper()

	ticker := time.NewTicker(every)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			if test() {
				return
			}
		case <-timer.C:
			t.Fatalf("timeout on eventually: %v", timeout)
		}
	}
}
