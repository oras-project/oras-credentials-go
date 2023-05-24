package client

import (
	"context"
	"reflect"
	"testing"
)

func TestNewShellProgramFunc(t *testing.T) {
	tests := []struct {
		testName        string
		programFuncName string
		programName     string
	}{
		{
			testName:        "use NewShellProgramFunc() and create a Program",
			programFuncName: "test-credential-helper",
			programName:     "get",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			programFunc := NewShellProgramFunc(tt.programFuncName)
			program := programFunc(context.Background(), tt.programName)
			shell, ok := program.(*Shell)
			if !ok {
				t.Error("Incorrect type of program")
			}
			gotCmd := shell.cmd
			wantCmd := createProgramCmdRedirectErr(context.Background(), tt.programFuncName, []string{tt.programName}, nil)
			if !reflect.DeepEqual(gotCmd.String(), wantCmd.String()) {
				t.Errorf("gotCmd.String() = %v, want %v", gotCmd, wantCmd)
			}
		})
	}
}
