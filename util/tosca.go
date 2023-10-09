package util

import (
	contextpkg "context"

	"github.com/tliron/exturl"
	"github.com/tliron/kutil/problems"
	"github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
	"github.com/tliron/puccini/normal"
	parserpkg "github.com/tliron/puccini/tosca/parser"
)

var parser = parserpkg.NewParser()

//
// ToscaParser
//

type ToscaParser struct {
	Clout *clout.Clout

	urlContext *exturl.Context
	problems   *problems.Problems
}

func NewToscaParser() *ToscaParser {
	return &ToscaParser{
		urlContext: exturl.NewContext(),
	}
}

func (self *ToscaParser) Release() error {
	return self.urlContext.Release()
}

func (self *ToscaParser) Parse(context contextpkg.Context, url string) error {
	base, err := self.urlContext.NewWorkingDirFileURL()
	if err != nil {
		return err
	}

	if url_, err := self.urlContext.NewValidAnyOrFileURL(context, url, []exturl.URL{base}); err == nil {
		parserContext := parser.NewContext()
		parserContext.URL = url_
		var serviceTemplate *normal.ServiceTemplate

		if serviceTemplate, err = parserContext.Parse(context); err == nil {
			if self.Clout, err = serviceTemplate.Compile(); err == nil {
				self.problems = parserContext.GetProblems()

				self.execContext().Resolve()

				if !self.problems.Empty() {
					return self.problems.ToError(false)
				}

				return nil
			} else {
				if !self.problems.Empty() {
					return self.problems.ToError(false)
				}

				return err
			}
		} else {
			if (self.problems != nil) && !self.problems.Empty() {
				return self.problems.ToError(false)
			}

			return err
		}
	} else {
		return err
	}
}

func (self *ToscaParser) Coerce() error {
	self.execContext().Coerce()

	if !self.problems.Empty() {
		return self.problems.ToError(false)
	}

	return nil
}

func (self *ToscaParser) execContext() *js.ExecContext {
	return &js.ExecContext{
		Clout:      self.Clout,
		Problems:   self.problems,
		URLContext: self.urlContext,
		History:    true,
		Format:     "yaml",
		Strict:     true,
		Pretty:     false,
	}
}
