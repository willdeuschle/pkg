// Copyright 2016 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package matcher

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
)

type Matcher interface {
	Match(relPath string) bool
}

type allMatcher []Matcher

func (m allMatcher) Match(relPath string) bool {
	nonNilMatcherExists := false
	for _, currMatcher := range []Matcher(m) {
		if currMatcher != nil {
			nonNilMatcherExists = true
			if !currMatcher.Match(relPath) {
				return false
			}
		}
	}
	return nonNilMatcherExists
}

// All returns a compound Matcher that returns true if all of its provided non-nil Matchers return true. Returns false
// if no matchers are provided or if all of the provided matchers are nil.
func All(matchers ...Matcher) Matcher {
	return allMatcher(append([]Matcher{}, matchers...))
}

type anyMatcher []Matcher

func (m anyMatcher) Match(relPath string) bool {
	for _, currMatcher := range []Matcher(m) {
		if currMatcher != nil && currMatcher.Match(relPath) {
			return true
		}
	}
	return false
}

// Not returns a matcher that returns the negation of the provided matcher.
func Not(matcher Matcher) Matcher {
	return notMatcher{
		matcher: matcher,
	}
}

type notMatcher struct {
	matcher Matcher
}

func (m notMatcher) Match(relPath string) bool {
	return !m.matcher.Match(relPath)
}

// Any returns a compound Matcher that returns true if any of the provided Matchers return true.
func Any(matchers ...Matcher) Matcher {
	return anyMatcher(append([]Matcher{}, matchers...))
}

// Hidden returns a matcher that matches all hidden files or directories (any path that begins with `.`).
func Hidden() Matcher {
	return Name(`\..+`)
}

// Name returns a Matcher that matches the on the name of all of the components of a path using the provided
// expressions. Each part of the path is tested against the expressions independently (no path separators). The name
// must fully match the expression to be considered a match.
func Name(regexps ...string) Matcher {
	compiled := make([]*regexp.Regexp, len(regexps))
	for i, curr := range regexps {
		compiled[i] = regexp.MustCompile(curr)
	}
	return nameMatcher(compiled)
}

type nameMatcher []*regexp.Regexp

func (m nameMatcher) Match(inputRelPath string) bool {
	for _, currSubpath := range allSubpaths(inputRelPath) {
		currName := path.Base(currSubpath)
		for _, currRegExp := range []*regexp.Regexp(m) {
			matchLoc := currRegExp.FindStringIndex(currName)
			if len(matchLoc) > 0 && matchLoc[0] == 0 && matchLoc[1] == len(currName) {
				return true
			}
		}
	}
	return false
}

// Path returns a Matcher that matches any path that matches or is a subpath of any of the provided paths. For example,
// a value of "foo" would match the relative directory "foo" and all of its sub-paths ("foo/bar", "foo/bar.txt"), but
// not every directory named "foo" (would not match "bar/foo"). Matches are done using glob matching (same as
// filepath.Match). However, unlike filepath.Match, subpath matches will match all of the sub-paths of a given match as
// well (for example, the pattern "foo/*/bar" matches "foo/*/bar/baz").
func Path(paths ...string) Matcher {
	return pathMatcher(paths)
}

type pathMatcher []string

func (m pathMatcher) Match(inputRelPath string) bool {
	subpaths := allSubpaths(inputRelPath)
	for _, currMatcherPathPattern := range []string(m) {
		for _, currSubpath := range subpaths {
			match, err := filepath.Match(currMatcherPathPattern, currSubpath)
			if err != nil {
				// only possible error is bad pattern
				panic(fmt.Sprintf("filepath: Match(%q): %v", currMatcherPathPattern, err))
			}
			if match {
				return true
			}
		}
	}
	return false
}

// allSubpaths returns the provided relative paths and all of its subpaths up to (but not including) ".". For example,
// "foo/bar/baz.txt" return [foo/bar/baz.txt foo/bar foo], while "foo.txt" returns [foo.txt]. Returns nil if the
// provided path is an absolute path.
func allSubpaths(relPath string) []string {
	if path.IsAbs(relPath) {
		return nil
	}
	var subpaths []string
	for currRelPath := relPath; currRelPath != "."; currRelPath = path.Dir(currRelPath) {
		subpaths = append(subpaths, currRelPath)
	}
	return subpaths
}
