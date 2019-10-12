package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const versionPkg = "github.com/fhofherr/acmeproxy/pkg/version"

var (
	static = flag.Bool("static", false, "Build a statically linked binary.")
)

func main() {
	flag.Parse()

	buildTime := time.Now().UTC().Format("2006-01-02:15:04:05.000")
	gitHash, err := git("rev-parse", "HEAD")
	if err != nil {
		fmt.Printf("Failed to determine git hash: %v\n", err)
		os.Exit(1)
	}
	gitTag, err := git("describe", "--tags")
	if err != nil {
		fmt.Println("Failed to determine git tag")
		// A missing git tag is not a fatal error. Do not exit here.
	}

	outpath := flag.Arg(0)
	if outpath == "" {
		fmt.Printf("Usage:\n\t%s [flags] <outpath>\n", os.Args[0])
		os.Exit(1)
	}
	env := appendOSandArch(nil, outpath)
	if *static {
		env = append(env, "CGO_ENABLED=0")
	}
	args := []string{"build", "-a", "-o", outpath}
	args = appendLDFlags(args, buildTime, gitHash, gitTag)
	fmt.Printf("%s go %s\n", strings.Join(env, " "), strings.Join(args, " "))
	stdout, stderr, err := runCmd(env, "go", args...)
	if err != nil {
		fmt.Printf("Could not execute go: %v\n", err)
		fmt.Printf("Stdout:\n\n%s\n", stdout)
		fmt.Printf("Stderr:\n\n%s\n", stderr)
	}
}

func appendOSandArch(env []string, outpath string) []string {
	parts := strings.Split(outpath, string(os.PathSeparator))
	if len(parts) < 3 {
		fmt.Printf("Expected at least three components in outpath; got: %d\n", len(parts))
		os.Exit(1)
	}
	goos := parts[1]
	if goos != "local" && len(parts) == 4 {
		goarch := parts[2]
		env = append(env, "GOOS="+goos, "GOARCH="+goarch)
	}
	return env
}

func appendLDFlags(args []string, buildTime, gitHash, gitTag string) []string {
	ldflags := []string{
		"-s",
		"-X", fmt.Sprintf("%s.BuildTime=%s", versionPkg, buildTime),
		"-X", fmt.Sprintf("%s.GitHash=%s", versionPkg, gitHash),
	}
	if gitTag != "" {
		ldflags = append(ldflags, "-X", fmt.Sprintf("%s.GitTag=%s", versionPkg, gitTag))
	}
	args = append(args, "-ldflags", strings.Join(ldflags, " "))
	return args
}

func runCmd(env []string, cmdPath string, args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command(cmdPath, args...)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, env...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return fmt.Sprintf("%s", stdout.Bytes()), fmt.Sprintf("%s", stdout.Bytes()), err
}

func git(args ...string) (string, error) {
	stdout, _, err := runCmd(nil, "git", args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}
