package gql_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGql(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gql Suite")
}
