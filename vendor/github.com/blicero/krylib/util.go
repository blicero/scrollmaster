// /home/krylon/go/src/krylib/util.go
// -*- coding: utf-8; -*-
// Created on 28. 02. 2013 by Benjamin Walkenhorst
// (c) 2013 Benjamin Walkenhorst
// -*- mode: go; -*-
// Time-stamp: <2023-09-05 19:11:50 krylon>

package krylib

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// Units of measurement, used for formatting amounts of data.
var units = []string{
	"Bytes",
	"KB",
	"MB",
	"GB",
	"TB",
	"PB",
}

// FmtBytes formats a number of bytes into a human-readable string,
// i.e. 1536 becomes "1.5 KB".
func FmtBytes(bytes int64) string {
	idx := 0
	maxIdx := len(units) - 1
	var amt = float64(bytes)

	for (amt > 1024) && (idx < maxIdx) {
		idx++
		amt /= 1024
	}

	return fmt.Sprintf("%.1f %s", amt, units[idx])
} // func FmtBytes(bytes int64) string

// Fexists checks if a file exists.
func Fexists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
} // func Fexists(path string) (bool, error)

// IsDir checks if the given path exists and is a Directory.
// If the given path does not exists, it returns false, nil
func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return ((info.Mode() & os.ModeDir) != 0), nil
} // func IsDir(path string) (bool, error)

// FileSize returns the size of the given file in bytes.
// If the file does not exist, it returns an error
func FileSize(path string) (int64, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return -1, err
	}

	return stat.Size(), nil
} // func FileSize(path string) (int64, error)

// Mtime returns the mtime stamp of the given file.
// If the path does not exist, it returns an error
func Mtime(path string) (time.Time, error) {
	var info os.FileInfo
	var err error

	if info, err = os.Stat(path); err != nil {
		return time.Unix(0, 0), err
	}

	return info.ModTime(), nil
} // func Mtime(path string) (time.Time, error)

// Fibonacci returns the nth Fibonacci number.
func Fibonacci(n int) int64 {
	var f1, f2, cnt int64

	if n < 2 {
		return 1
	}

	f2 = 1
	cnt = 2

	for i := 1; i < n; i++ {
		f1 = f2
		f2 = cnt
		cnt = f1 + f2
	}

	return f2
} // func Fibonacci(n int) int64

// ParseURL parses a string containing (hopefully) a URL.
// Returns the url.URL pointer if successful, panics otherwise.
// The point is to be able to use this function to initialize
// "global" variables.
func ParseURL(s string) *url.URL {
	var err error
	var addr *url.URL
	if addr, err = url.Parse(s); err != nil {
		panic(err)
	} else {
		return addr
	}
} // func ParseURL(s string) *url.URL

// Trace emits info about the CALLING function
// I found this code on StackOverflow, so the copyright
// situation is a bit hazy.
// Then again, if you didn't want people to use your
// code like I do here, you'd probably not post it in
// StackOverflow.
// Ergo I consider this public domain code, since I am now also
// unable to find the thread where I saw this code.
// My point is that I did not come up with this myself but
// I use it happily.
//
// PS: I should readlly give this function a do-over so that it returns
//
//	the data it obtains, rather than writing it to Stdout.
func Trace() {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fmt.Printf("%s:%d %s\n", file, line, f.Name())
} // func Trace()

// TraceInfo returns a string that contains the name of the source file,
// the line number in the file and the name of the function where TraceInfo
// is called.
func TraceInfo() string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	var s = fmt.Sprintf("%s:%d %s\n", file, line, f.Name())
	return s
} // func TraceInfo() string

// StringSlice takes a slice of Stringers (i.e. objects that implement the
// String() method), creates a slice of strings of the same size as the input,
// then calls String() on each member of the input slice and stores the
// result in the output slice which it returns at the end.
func StringSlice(arr []fmt.Stringer) []string {
	var out = make([]string, len(arr))

	for idx, val := range arr {
		out[idx] = val.String()
	}

	return out
} // func StringSlice(arr []fmt.Stringer) []String

// Date returns a Time value that 00:00:00 local time on the given date.
func Date(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local)
} // func Date(y, m, d int) time.Time

var eolPat = regexp.MustCompile("[\r\n]+$")

// Chomp returns a copy of the argument string with all line-ending characters
// removed.
func Chomp(s string) string {
	return eolPat.ReplaceAllString(s, "")
} // func Chomp(s string) string

// CopyFile copies the file at src to dst.
// It copies only the content, not metadata liker ownership, access right, etc.
//
// In case of an error, an incomplete or corrupt file at dst  might remain.
func CopyFile(src, dst string) error {
	var (
		in, out *os.File
		err     error
	)

	if in, err = os.Open(src); err != nil { // nolint: gosec
		return err
	}
	defer in.Close() // nolint: errcheck

	if out, err = os.Create(dst); err != nil {
		return err
	}
	defer out.Close() // nolint: errcheck

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	var info os.FileInfo

	if info, err = in.Stat(); err != nil {
		return err
	} else if err = out.Chmod(info.Mode()); err != nil {
		return err
	}

	return nil
} // func CopyFile(src, dst string) error

// GetHomeDirectory determines the user's home directory.
// The reason this is even a thing is that Unix-like systems
// store this path in the environment variable "HOME",
// whereas Windows uses the environment variable "USERPROFILE",
// Hence this function.
func GetHomeDirectory() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}

	return os.Getenv("HOME")
} // func GetHomeDirectory() string

// ExpandTilde replaces a leading "~" in a path
// with the current user's home directory.
func ExpandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	var fullPath = filepath.Join(
		GetHomeDirectory(),
		path[1:],
	)

	return fullPath
} // func expandTilde(path string) string

// IntRange returns a slice of integers in the range of [0,n)
// If that sounds primitive, that is because it *is* primitive -- languages like
// Perl or Ruby offer language primitives for this kind of feature.
// But Go likes to keep it simple, so here we go.
func IntRange(n int) []int {
	var r = make([]int, n)

	for i := 0; i < n; i++ {
		r[i] = i
	}

	return r
} // func IntRange(n int) []int

// Number is the sum type of all integer and floating point types that are strictly ordered.
type Number interface {
	int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64 | float32 | float64
}

// Min returns the smaller of two values.
func Min[T Number](x, y T) T {
	if x < y {
		return x
	} else {
		return y
	}
} // func Min[T Number](x, y T) T

// Max returns the greater of two values
func Max[T Number](x, y T) T {
	if x > y {
		return x
	} else {
		return y
	}
} // func Max[T Number](x, y T) T

var whitespacePat = regexp.MustCompile("[[:space:]]+")

// SplitOnWhitespace splits a string on whitespace and returns a slice of the resulting substrings.
func SplitOnWhitespace(s string) []string {
	return whitespacePat.Split(s, -1)
} // func SplitOnWhitespace(s string) []string
