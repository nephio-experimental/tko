package plugins

import (
	"bytes"
	contextpkg "context"
	"errors"
	"io"
	"os/exec"
	"strings"
)

func ExecuteLocal(context contextpkg.Context, stdin io.Reader, command ...string) ([]byte, error) {
	if len(command) == 0 {
		return nil, errors.New("command must have at least one argument")
	}

	log.Infof("execute local: %s", command)

	name := command[0]
	arguments := command[1:]
	cmd := exec.CommandContext(context, name, arguments...)

	var stdout, stderr bytes.Buffer
	cmd.Stdin = stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if strings.HasSuffix(name, ".py") {
		if pythonPath, err := getPythonPath(); err == nil {
			cmd.Env = append(cmd.Env, appendEnvPath(pythonPath))
		} else {
			return nil, err
		}
	}

	if err := cmd.Run(); err == nil {
		return stdout.Bytes(), nil
	} else {
		return nil, errorWithStderr(err, stderr.String())
	}
}
