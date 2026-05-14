package characterization_test

import (
	"backend-v2/test/characterization/helpers"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	helpers.Boot()
	code := m.Run()
	helpers.Shutdown()
	os.Exit(code)
}
