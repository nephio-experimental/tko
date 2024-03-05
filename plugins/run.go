package plugins

import (
	"bytes"
	contextpkg "context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Run(context contextpkg.Context, stdin io.Reader, command ...string) ([]byte, error) {
	if len(command) == 0 {
		return nil, errors.New("command must have at least one argument")
	}

	var pythonPath string
	var err error
	if pythonPath, err = getPythonPath(); err != nil {
		return nil, err
	}

	name := command[0]
	arguments := command[1:]
	cmd := exec.CommandContext(context, name, arguments...)

	var stdout, stderr bytes.Buffer
	cmd.Stdin = stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(cmd.Env, appendEnvPath(pythonPath))

	if err := cmd.Run(); err == nil {
		return stdout.Bytes(), nil
	} else {
		return nil, withStderr(err, stderr.String())
	}
}

// Utils

func withStderr(err error, stderr string) error {
	stderr = strings.Trim(stderr, "\r\n")
	if stderr != "" {
		return errors.New(err.Error() + "\n" + stderr)
	} else {
		return err
	}
}

func appendEnvPath(path string) string {
	return "PATH=" + path + string(os.PathListSeparator) + os.Getenv("PATH")
}

func getPythonPath() (string, error) {
	pythonPath := os.Getenv("PYTHON_ENV")
	if pythonPath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			pythonPath = filepath.Join(home, "tko-python-env")
		} else {
			return "", err
		}
	}
	return filepath.Join(pythonPath, "bin"), nil
}
