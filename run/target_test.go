package run

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPathMissingDest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "source")
	err := os.WriteFile(src, []byte("hi!"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "missing")
	rebuild, err := Path(dst, src)
	if err != nil {
		t.Fatal("Expected no error, but got", err)
	}
	if !rebuild {
		t.Fatal("expected to be told to rebuild, but got false")
	}
}

func TestPathMissingSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dst := filepath.Join(dir, "dst")
	err := os.WriteFile(dst, []byte("hi!"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "missing")
	_, err = Path(dst, src)
	if !os.IsNotExist(err) {
		t.Fatal("Expected os.IsNotExist(err), but got", err)
	}
}

func TestGlobEmptyGlob(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dst := filepath.Join(dir, "dst")
	err := os.WriteFile(dst, []byte("hi!"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "src*")
	_, err = Glob(dst, src)
	if err == nil {
		t.Fatal("Expected error, but got nil")
	}
}

func TestDirMissingSrc(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dst := filepath.Join(dir, "dst")
	err := os.WriteFile(dst, []byte("hi!"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "missing")
	_, err = Dir(dst, src)
	if !os.IsNotExist(err) {
		t.Fatal("Expected os.IsNotExist(err), but got", err)
	}
}

func TestDirMissingDest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "source")
	err := os.Mkdir(src, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(src, "somefile"), []byte("hi!"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "missing")
	rebuild, err := Dir(dst, src)
	if err != nil {
		t.Fatal("Expected no error, but got", err)
	}
	if !rebuild {
		t.Fatal("expected to be told to rebuild, but got false")
	}
}

func TestGlob(t *testing.T) {
	dir := t.TempDir()
	err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash("dir/dir2")), 0o777)
	if err != nil {
		t.Fatal(err)
	}
	// files are created in order so we know how to expect
	files := []string{
		"old_executable",
		"file_one.src",
		"dir/file_two.src",
		"middle_executable",
		"file_three.src",
		"dir/dir2/file_four.src",
		"built_executable",
	}
	for _, v := range files {
		time.Sleep(10 * time.Millisecond)
		f := filepath.Join(dir, filepath.FromSlash(v))
		err := os.WriteFile(f, []byte(v), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// force an environment variable for testing
	t.Setenv("MYVAR", "file")
	t.Setenv("THREE", "three")

	table := []struct {
		desc    string
		target  string
		sources []string
		expect  bool
	}{
		{
			desc:    "Missing target",
			target:  "missing_file",
			sources: []string{"file*.src"},
			expect:  true,
		},
		{
			desc:    "Target is newer",
			target:  "built_executable",
			sources: []string{"*.src", "dir/*.src"},
			expect:  false,
		},
		{
			desc:    "No actual globs",
			target:  "built_executable",
			sources: []string{"file_one.src", "file_three.src"},
			expect:  false,
		},
		{
			desc:    "Target is older",
			target:  "old_executable",
			sources: []string{"f*.src"},
			expect:  true,
		},
		{
			desc:    "Target is in the middle of files in the glob",
			target:  "middle_executable",
			sources: []string{"file*"},
			expect:  true,
		},
		{
			desc:    "Globs work for dirs",
			target:  "older_executable",
			sources: []string{"d*"},
			expect:  true,
		},
	}

	for _, c := range table {
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			for i := range c.sources {
				c.sources[i] = filepath.Join(dir, c.sources[i])
			}
			c.target = filepath.Join(dir, c.target)
			v, err := Glob(c.target, c.sources...)
			if err != nil {
				t.Fatal(err)
			}
			if v != c.expect {
				t.Errorf("expecting %v got %v", c.expect, v)
			}
		})
	}
}

func TestPath(t *testing.T) {
	dir := t.TempDir()
	err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash("dir/dir2")), 0o777)
	if err != nil {
		t.Fatal(err)
	}
	// files are created in order so we know how to expect
	files := []string{
		"file_one",
		"dir/file_two",
		"file_three",
		"dir/dir2/file_four",
	}
	for _, v := range files {
		time.Sleep(10 * time.Millisecond)
		f := filepath.Join(dir, filepath.FromSlash(v))
		err := os.WriteFile(f, []byte(v), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// force an environment variable for testing
	t.Setenv("MYVAR", "file")
	t.Setenv("THREE", "three")

	table := []struct {
		desc    string
		target  string
		sources []string
		expect  bool
	}{
		{
			desc:    "Missing target",
			target:  "missing_file",
			sources: []string{"file_one"},
			expect:  true,
		},
		{
			desc:    "Target is newer",
			target:  "file_three",
			sources: []string{"file_one"},
			expect:  false,
		},
		{
			desc:    "Target is older",
			target:  "file_one",
			sources: []string{"file_three"},
			expect:  true,
		},
		{
			// note that even though file_four is in dir/dir2 ... the modtimes
			// only get propagated up to the parent directory of the folder, not
			// propagated all the way up.
			desc:    "Source is older dir",
			target:  "file_three",
			sources: []string{"dir"},
			expect:  false,
		},
		{
			desc:    "Source is newer dir2",
			target:  "file_three",
			sources: []string{"dir/dir2"},
			expect:  true,
		},
		{
			desc:    "Source is newer dir",
			target:  "file_one",
			sources: []string{"dir"},
			expect:  true,
		},
		{
			desc:    "Target is newer; expand source",
			target:  "${MYVAR}_$THREE",
			sources: []string{"file_one"},
			expect:  false,
		},
		{
			desc:    "Target is older; expand dest",
			target:  "file_one",
			sources: []string{"${MYVAR}_$THREE"},
			expect:  true,
		},
	}

	for _, c := range table {
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			for i := range c.sources {
				c.sources[i] = filepath.Join(dir, c.sources[i])
			}
			c.target = filepath.Join(dir, c.target)
			v, err := Path(c.target, c.sources...)
			if err != nil {
				t.Fatal(err)
			}
			if v != c.expect {
				t.Errorf("expecting %v got %v", c.expect, v)
			}
		})
	}
}

func TestDir(t *testing.T) {
	dir := t.TempDir()
	err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash("dir/dir2")), 0o777)
	if err != nil {
		t.Fatal(err)
	}
	// files are created in order so we know which one is newer
	files := []string{
		"file_one",
		"dir/file_two",
		"file_three",
		"dir/dir2/file_four",
		"file_five",
	}
	for _, v := range files {
		time.Sleep(10 * time.Millisecond)
		f := filepath.Join(dir, filepath.FromSlash(v))
		err := os.WriteFile(f, []byte(v), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// force environment variables for testing
	t.Setenv("MYFILE", "file")
	t.Setenv("MYDIR", "dir")
	t.Setenv("X1", "one")

	table := []struct {
		desc    string
		target  string
		sources []string
		expect  bool
	}{
		{
			desc:    "Missing target",
			target:  "missing_file",
			sources: []string{"file_one"},
			expect:  true,
		},
		{
			desc:    "Target is newer",
			target:  "file_three",
			sources: []string{"file_one"},
			expect:  false,
		},
		{
			desc:    "Target is older",
			target:  "file_one",
			sources: []string{"file_three"},
			expect:  true,
		},
		{
			desc:    "Source is older dir",
			target:  "file_five",
			sources: []string{"dir"},
			expect:  false,
		},
		{
			desc:    "Source is newer dir",
			target:  "file_one",
			sources: []string{"dir"},
			expect:  true,
		},
		{
			// This is the tricky one. The modtime on "dir" will be the same
			// as the modtime on dir/file_two, but the modtime on the subdir
			// will be the same as the modtime on dir/dir2/file_four
			// and therefor the should say the source is newer.
			desc:    "Source is newer subdir",
			target:  "file_three",
			sources: []string{"dir"},
			expect:  true,
		},
		{
			desc:    "Target is newer (with env expansion)",
			target:  "${MYFILE}_three",
			sources: []string{"${MYFILE}_$X1"},
			expect:  false,
		},
		{
			desc:    "Target is older (with env expansion)",
			target:  "${MYFILE}_one",
			sources: []string{"$MYFILE_three"},
			expect:  true,
		},
		{
			desc:    "Source is older dir (with env expansion)",
			target:  "${MYFILE}_five",
			sources: []string{"${MYDIR}"},
			expect:  false,
		},
		{
			desc:    "Source is newer dir (with env expansion)",
			target:  "${MYFILE}_$X1",
			sources: []string{"$MYDIR"},
			expect:  true,
		},
		{
			desc:    "Source file is newer than dst dir",
			target:  "dir/dir2",
			sources: []string{"file_five"},
			expect:  true,
		},
		{
			desc:    "Source file is not newer than dst dir",
			target:  "dir/dir2",
			sources: []string{"file_one"},
			expect:  false,
		},
	}

	for _, c := range table {
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			sources := make([]string, len(c.sources))
			for i := range c.sources {
				sources[i] = filepath.Join(dir, c.sources[i])
			}
			target := filepath.Join(dir, c.target)
			v, err := Dir(target, sources...)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if v != c.expect {
				t.Errorf("expecting %v got %v", c.expect, v)
			}
		})
	}
}
