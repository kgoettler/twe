package cmd

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CmdSuite struct {
	suite.Suite
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCmdSuite(t *testing.T) {
	suite.Run(t, new(CmdSuite))
}

func (suite *CmdSuite) TestTimecard_FromFile() {
	cwd, err := os.Getwd()
	suite.Require().NoError(err)
	actual := new(bytes.Buffer)
	RootCmd.SetOut(actual)
	RootCmd.SetErr(actual)
	RootCmd.SetArgs([]string{"timecard", "-f", filepath.Join(cwd, "testdata", "sample.input")})
	err = RootCmd.Execute()
	suite.Require().NoError(err)
	RootCmd.SetArgs([]string{})
}

func (suite *CmdSuite) TestTimecard_WithArgs() {
	actual := new(bytes.Buffer)
	RootCmd.SetOut(actual)
	RootCmd.SetErr(actual)
	RootCmd.SetArgs([]string{"timecard", "2024-01-01", "-", "2024-01-08"})
	err := RootCmd.Execute()
	suite.Require().NoError(err)
}

func (suite *CmdSuite) TestLast() {
	actual := new(bytes.Buffer)
	RootCmd.SetOut(actual)
	RootCmd.SetArgs([]string{"last"})
	err := RootCmd.Execute()
	suite.Equal("20240108T040000\n", actual.String())
	suite.Require().NoError(err)
}

func (suite *CmdSuite) TestImport() {
	input := bytes.NewReader([]byte(`[
{"id":5,"start":"20250415T040000Z","end":"20250415T100000Z","tags":["Sleep"]},
{"id":4,"start":"20250415T100000Z","end":"20250415T110000Z","tags":["Shower"]},
{"id":3,"start":"20250415T110000Z","end":"20250415T120000Z","tags":["Breakfast"]},
{"id":2,"start":"20250415T120000Z","end":"20250415T130000Z","tags":["'Commuting to Work'"]},
{"id":1,"start":"20250415T130000Z","end":"20250415T210000Z","tags":["Foo"]}
]`))

	actual := new(bytes.Buffer)
	RootCmd.SetOut(actual)
	RootCmd.SetArgs([]string{"import"})
	RootCmd.SetIn(input)
	err := RootCmd.Execute()
	suite.Require().NoError(err)
}
