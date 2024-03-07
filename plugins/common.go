package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tliron/commonlog"
)

var log = commonlog.GetLogger("plugins")

func errorWithStderr(err error, stderr string) error {
	stderr = strings.Trim(stderr, "\r\n")
	if stderr != "" {
		return fmt.Errorf("%w\n%s", err, stderr)
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
