// ABOUTME: Ginkgo test suite bootstrap for Pivnet acceptance tests.
// ABOUTME: Configures test environment and shared test fixtures.
package test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pivnet Acceptance Suite")
}
