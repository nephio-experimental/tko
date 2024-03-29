package commands

import (
	contextpkg "context"
	"fmt"
	"os"
	pathpkg "path"
	"strings"

	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/exturl"
	"github.com/tliron/kutil/streampackage"
	"github.com/tliron/kutil/util"
)

func readPackage(context contextpkg.Context, url string, stdin bool) (tkoutil.Package, error) {
	if stdin && (url != "") {
		util.Fail("cannot specify both --stdin=true and --url=")
	}

	var package_ tkoutil.Package

	var err error
	if stdin {
		package_, err = readPackageFromStdin()
	} else {
		if url == "" {
			util.Fail("must specify either --stdin=true or --url=")
		}
		package_, err = readPackageFromUrl(context, url)
	}
	if err != nil {
		return nil, err
	}

	return package_, nil
}

func readPackageFromStdin() (tkoutil.Package, error) {
	log.Info("reading package from stdin")
	return tkoutil.ReadPackage("yaml", os.Stdin)
}

func readPackageFromUrl(context contextpkg.Context, url string) (tkoutil.Package, error) {
	urlContext := exturl.NewContext()
	util.OnExitError(urlContext.Release)

	base, err := urlContext.NewWorkingDirFileURL()
	util.OnExitError(urlContext.Release)

	url_, err := urlContext.NewValidAnyOrFileURL(context, url, []exturl.URL{base})
	if err != nil {
		return nil, err
	}

	log.Infof("reading package from URL: %s", url_)

	var unpack string
	if strings.HasSuffix(url, ".tar") {
		unpack = "tar"
	} else if strings.HasSuffix(url, ".tar.gz") || strings.HasSuffix(url, ".tgz") {
		unpack = "tgz"
	} else if strings.HasSuffix(url, ".zip") {
		unpack = "zip"
	}

	streamPackage, err := streampackage.NewStreamPackage(context, url_, unpack)
	if err != nil {
		return nil, err
	}

	var package_ tkoutil.Package

	for {
		if stream, err := streamPackage.Next(); err == nil {
			if stream == nil {
				break
			}

			if reader, path, _, err := stream.Open(context); err == nil {
				if ext := pathpkg.Ext(path); (ext == ".yaml") || (ext == ".yml") {
					reader = util.NewContextualReadCloser(context, reader)

					if list_, err := tkoutil.ReadPackage("yaml", reader); err == nil {
						package_ = append(package_, list_...)
					} else {
						reader.Close()
						return nil, fmt.Errorf("%s: %s", path, err.Error())
					}

					if err := reader.Close(); err != nil {
						return nil, err
					}
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return package_, nil
}
