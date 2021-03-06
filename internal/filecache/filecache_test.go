package filecache

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/txtar"

	"github.com/Helcaraxan/modularise/internal/testlib"
)

func TestModulePath(t *testing.T) {
	t.Parallel()

	tc, err := ioutil.ReadFile("./testdata/module_info.txtar")
	testlib.NoError(t, true, err)

	a := txtar.Parse(tc)

	mi := struct {
		Dir  string `json:"Dir"`
		Path string `json:"Path"`
	}{}
	err = json.Unmarshal(a.Comment, &mi)
	testlib.NoError(t, true, err)

	parallelTestAllCacheTypes(t, a, func(t *testing.T, fc FileCache) {
		t.Parallel()

		testlib.Equal(t, false, mi.Path, fc.ModulePath())
	})
}

func TestPkgs(t *testing.T) {
	t.Parallel()

	tc, err := ioutil.ReadFile("./testdata/pkgs.txtar")
	testlib.NoError(t, true, err)

	a := txtar.Parse(tc)

	eps := map[string]bool{}
	for _, p := range strings.Split(strings.TrimSpace(string(a.Comment)), "\n") {
		if strings.HasPrefix(p, "#") || p == "" {
			continue
		}
		eps[p] = true
	}

	parallelTestAllCacheTypes(t, a, func(t *testing.T, fc FileCache) {
		t.Parallel()

		pkgs, err := fc.Pkgs()
		testlib.NoError(t, true, err)
		testlib.Equal(t, false, eps, pkgs)
	})
}

func TestFiles(t *testing.T) {
	t.Parallel()

	tc, err := ioutil.ReadFile("./testdata/files.txtar")
	testlib.NoError(t, true, err)

	a := txtar.Parse(tc)

	efs := map[string]bool{}
	for _, ef := range strings.Split(strings.TrimSpace(string(a.Comment)), "\n") {
		if strings.HasPrefix(ef, "#") || ef == "" {
			continue
		}
		efs[ef] = true
	}

	parallelTestAllCacheTypes(t, a, func(t *testing.T, fc FileCache) {
		t.Parallel()

		fs, err := fc.Files()
		testlib.NoError(t, true, err)
		testlib.Equal(t, false, efs, fs)
	})
}

func TestFilesInPkg(t *testing.T) {
	t.Parallel()

	tc, err := ioutil.ReadFile("./testdata/files_in_pkg.txtar")
	testlib.NoError(t, true, err)

	a := txtar.Parse(tc)

	efs := map[string]bool{}
	for _, ef := range strings.Split(strings.TrimSpace(string(a.Comment)), "\n") {
		if strings.HasPrefix(ef, "#") || ef == "" {
			continue
		}
		efs[ef] = true
	}

	parallelTestAllCacheTypes(t, a, func(t *testing.T, fc FileCache) {
		t.Parallel()

		fs, err := fc.FilesInPkg(filepath.Join(fc.ModulePath(), "lib"))
		testlib.NoError(t, true, err)
		testlib.Equal(t, false, efs, fs)
	})
}

func TestReadFile(t *testing.T) {
	t.Parallel()

	tc, err := ioutil.ReadFile("./testdata/read_file.txtar")
	testlib.NoError(t, true, err)

	a := txtar.Parse(tc)

	tfs := map[string][]byte{}
	for _, f := range strings.Split(strings.TrimSpace(string(a.Comment)), "\n") {
		if strings.HasPrefix(f, "#") || f == "" {
			continue
		}
		tfs[f] = nil
	}
	testlib.True(t, true, len(tfs) > 0)
	for _, f := range a.Files {
		if _, ok := tfs[f.Name]; ok {
			tfs[f.Name] = f.Data
		}
	}

	parallelTestAllCacheTypes(t, a, func(t *testing.T, fc FileCache) {
		t.Parallel()

		for tf, ec := range tfs {
			c, err := fc.ReadFile(tf)
			testlib.NoError(t, true, err)
			testlib.Equal(t, false, ec, c)
		}
	})
}

func TestReadGoFile(t *testing.T) {
	t.Parallel()

	tc, err := ioutil.ReadFile("./testdata/read_go_file.txtar")
	testlib.NoError(t, true, err)

	a := txtar.Parse(tc)

	tfs := map[string]bool{}
	for _, f := range strings.Split(strings.TrimSpace(string(a.Comment)), "\n") {
		if strings.HasPrefix(f, "#") || f == "" {
			continue
		}
		tfs[f] = true
	}
	testlib.True(t, true, len(tfs) > 0)

	parallelTestAllCacheTypes(t, a, func(t *testing.T, fc FileCache) {
		t.Parallel()

		var astInvarianceTested bool
		for tf := range tfs {
			fa, fs, err := fc.ReadGoFile(tf)
			// We deliberately do not try to compare the resulting AST or FileSet as we've already
			// assert in TestReadFile above that the content we get for the file is correct. Hence
			// we only need to ensure that we do not error out and return actual non-nil objects.

			testlib.NoError(t, true, err)
			testlib.NotNil(t, false, fa)
			testlib.NotNil(t, false, fs)

			if len(fa.Imports) > 0 {
				astInvarianceTested = true
				fa2, _, err := fc.ReadGoFile(tf)
				testlib.NoError(t, true, err)
				for _, imp := range fa2.Imports {
					imp.Path.Value = "invalid-import"
				}

				fa3, _, err := fc.ReadGoFile(tf)
				testlib.NoError(t, true, err)
				testlib.Equal(t, false, fa3, fa)
			}
		}
		testlib.True(t, false, astInvarianceTested)
	})
}

func parallelTestAllCacheTypes(t *testing.T, a *txtar.Archive, test func(*testing.T, FileCache)) {
	cacheTypes := map[string]Type{
		"TestCache": TestCache,
		"Uncache":   Uncache,
	}

	for cn := range cacheTypes {
		ct := cacheTypes[cn]
		t.Run(cn, func(t *testing.T) {
			cache, cleanup := initFileCache(t, ct, a)

			defer cleanup()

			test(t, cache)
		})
	}
}
