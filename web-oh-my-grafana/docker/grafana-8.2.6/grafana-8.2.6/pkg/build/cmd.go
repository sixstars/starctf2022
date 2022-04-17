package build

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	GoOSWindows = "windows"
	GoOSLinux   = "linux"

	ServerBinary = "grafana-server"
	CLIBinary    = "grafana-cli"
)

var binaries = []string{ServerBinary, CLIBinary}

func logError(message string, err error) int {
	log.Println(message, err)

	return 1
}

// RunCmd runs the build command and returns the exit code
func RunCmd() int {
	opts := BuildOptsFromFlags()

	wd, err := os.Getwd()
	if err != nil {
		return logError("Error getting working directory", err)
	}

	packageJSON, err := OpenPackageJSON(wd)
	if err != nil {
		return logError("Error opening package json", err)
	}

	opts.version = packageJSON.Version

	version, iteration := LinuxPackageVersion(packageJSON.Version, opts.buildID)

	if opts.printGenVersion {
		fmt.Print(genPackageVersion(version, iteration))
		return 0
	}

	log.Printf("Version: %s, Linux Version: %s, Package Iteration: %s\n", version, version, iteration)

	if flag.NArg() == 0 {
		log.Println("Usage: go run build.go build")
		return 1
	}

	for _, cmd := range flag.Args() {
		switch cmd {
		case "setup":
			setup(opts.goos)

		case "build-srv", "build-server":
			if !opts.isDev {
				clean(opts)
			}

			if err := doBuild("grafana-server", "./pkg/cmd/grafana-server", opts); err != nil {
				log.Println(err)
				return 1
			}

		case "build-cli":
			clean(opts)
			if err := doBuild("grafana-cli", "./pkg/cmd/grafana-cli", opts); err != nil {
				log.Println(err)
				return 1
			}

		case "build":
			//clean()
			for _, binary := range binaries {
				log.Println("building binaries", cmd)
				// Can't use filepath.Join here because filepath.Join calls filepath.Clean, which removes the `./` from this path, which upsets `go build`
				if err := doBuild(binary, fmt.Sprintf("./pkg/cmd/%s", binary), opts); err != nil {
					log.Println(err)
					return 1
				}
			}

		case "build-frontend":
			yarn("build")

		case "sha-dist":
			if err := shaDir("dist"); err != nil {
				return logError("error packaging dist directory", err)
			}

		case "latest":
			makeLatestDistCopies()

		case "clean":
			clean(opts)

		default:
			log.Println("Unknown command", cmd)
			return 1
		}
	}

	return 0
}

func makeLatestDistCopies() {
	files, err := ioutil.ReadDir("dist")
	if err != nil {
		log.Fatalf("failed to create latest copies. Cannot read from /dist")
	}

	latestMapping := map[string]string{
		"_amd64.deb":               "dist/grafana_latest_amd64.deb",
		".x86_64.rpm":              "dist/grafana-latest-1.x86_64.rpm",
		".linux-amd64.tar.gz":      "dist/grafana-latest.linux-x64.tar.gz",
		".linux-amd64-musl.tar.gz": "dist/grafana-latest.linux-x64-musl.tar.gz",
		".linux-armv7.tar.gz":      "dist/grafana-latest.linux-armv7.tar.gz",
		".linux-armv7-musl.tar.gz": "dist/grafana-latest.linux-armv7-musl.tar.gz",
		".linux-armv6.tar.gz":      "dist/grafana-latest.linux-armv6.tar.gz",
		".linux-arm64.tar.gz":      "dist/grafana-latest.linux-arm64.tar.gz",
		".linux-arm64-musl.tar.gz": "dist/grafana-latest.linux-arm64-musl.tar.gz",
	}

	for _, file := range files {
		for extension, fullName := range latestMapping {
			if strings.HasSuffix(file.Name(), extension) {
				if _, err := runError("cp", path.Join("dist", file.Name()), fullName); err != nil {
					log.Println("error running cp command:", err)
				}
			}
		}
	}
}

func yarn(params ...string) {
	runPrint(`yarn run`, params...)
}

func genPackageVersion(version string, iteration string) string {
	if iteration != "" {
		return fmt.Sprintf("%v-%v", version, iteration)
	} else {
		return version
	}
}

func setup(goos string) {
	args := []string{"install", "-v"}
	if goos == GoOSWindows {
		args = append(args, "-buildmode=exe")
	}
	args = append(args, "./pkg/cmd/grafana-server")
	runPrint("go", args...)
}

func doBuild(binaryName, pkg string, opts BuildOpts) error {
	log.Println("building", binaryName, pkg)
	libcPart := ""
	if opts.libc != "" {
		libcPart = fmt.Sprintf("-%s", opts.libc)
	}
	binary := fmt.Sprintf("./bin/%s", binaryName)

	//don't include os/arch/libc in output path in dev environment
	if !opts.isDev {
		binary = fmt.Sprintf("./bin/%s-%s%s/%s", opts.goos, opts.goarch, libcPart, binaryName)
	}

	if opts.goos == GoOSWindows {
		binary += ".exe"
	}

	if !opts.isDev {
		rmr(binary, binary+".md5")
	}

	lf, err := ldflags(opts)
	if err != nil {
		return err
	}

	args := []string{"build", "-ldflags", lf}

	if opts.goos == GoOSWindows {
		// Work around a linking error on Windows: "export ordinal too large"
		args = append(args, "-buildmode=exe")
	}

	if len(opts.buildTags) > 0 {
		args = append(args, "-tags", strings.Join(opts.buildTags, ","))
	}

	if opts.race {
		args = append(args, "-race")
	}

	args = append(args, "-o", binary)
	args = append(args, pkg)

	runPrint("go", args...)

	if opts.isDev {
		return nil
	}

	if err := setBuildEnv(opts); err != nil {
		return err
	}
	runPrint("go", "version")
	libcPart = ""
	if opts.libc != "" {
		libcPart = fmt.Sprintf("/%s", opts.libc)
	}
	fmt.Printf("Targeting %s/%s%s\n", opts.goos, opts.goarch, libcPart)

	// Create an md5 checksum of the binary, to be included in the archive for
	// automatic upgrades.
	return md5File(binary)
}

func ldflags(opts BuildOpts) (string, error) {
	buildStamp, err := buildStamp()
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	b.WriteString("-w")
	b.WriteString(fmt.Sprintf(" -X main.version=%s", opts.version))
	b.WriteString(fmt.Sprintf(" -X main.commit=%s", getGitSha()))
	b.WriteString(fmt.Sprintf(" -X main.buildstamp=%d", buildStamp))
	b.WriteString(fmt.Sprintf(" -X main.buildBranch=%s", getGitBranch()))
	if v := os.Getenv("LDFLAGS"); v != "" {
		b.WriteString(fmt.Sprintf(" -extldflags \"%s\"", v))
	}

	return b.String(), nil
}

func setBuildEnv(opts BuildOpts) error {
	if err := os.Setenv("GOOS", opts.goos); err != nil {
		return err
	}

	if opts.goos == GoOSWindows {
		// require windows >=7
		if err := os.Setenv("CGO_CFLAGS", "-D_WIN32_WINNT=0x0601"); err != nil {
			return err
		}
	}

	if opts.goarch != "amd64" || opts.goos != GoOSLinux {
		// needed for all other archs
		opts.cgo = true
	}

	if strings.HasPrefix(opts.goarch, "armv") {
		if err := os.Setenv("GOARCH", "arm"); err != nil {
			return err
		}

		if err := os.Setenv("GOARM", opts.goarch[4:]); err != nil {
			return err
		}
	} else {
		if err := os.Setenv("GOARCH", opts.goarch); err != nil {
			return err
		}
	}

	if opts.cgo {
		if err := os.Setenv("CGO_ENABLED", "1"); err != nil {
			return err
		}
	}

	if opts.gocc == "" {
		return nil
	}

	return os.Setenv("CC", opts.gocc)
}

func buildStamp() (int64, error) {
	// use SOURCE_DATE_EPOCH if set.
	if v, ok := os.LookupEnv("SOURCE_DATE_EPOCH"); ok {
		return strconv.ParseInt(v, 10, 64)
	}

	bs, err := runError("git", "show", "-s", "--format=%ct")
	if err != nil {
		return time.Now().Unix(), nil
	}

	return strconv.ParseInt(string(bs), 10, 64)
}

func clean(opts BuildOpts) {
	rmr("dist")
	rmr("tmp")
	rmr(filepath.Join(build.Default.GOPATH, fmt.Sprintf("pkg/%s_%s/github.com/grafana", opts.goos, opts.goarch)))
}
