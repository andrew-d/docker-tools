package main

import (
	"archive/tar"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrew-d/docker-tools/log"
	"github.com/fsouza/go-dockerclient"
	flag "github.com/ogier/pflag"
)

var (
	flagNoCache   bool
	flagRm        bool
	flagForceRm   bool
	flagRmAfter   bool
	flagEndpoint  string
	flagImageName string
)

func init() {
	flag.BoolVar(&flagNoCache, "no-cache", false,
		"Do not use cache when building the image")
	flag.BoolVar(&flagRm, "rm", false,
		"Remove intermediate containers after a successful build")
	flag.BoolVar(&flagForceRm, "force-rm", false,
		"Always remove intermediate containers, even after unsuccessful builds")
	flag.BoolVar(&flagRmAfter, "rm-after", false,
		"Remove the image from Docker after it's built and exported")
	flag.StringVarP(&flagEndpoint, "endpoint", "e", "unix:///var/run/docker.sock",
		"How to connect to the Docker service")
	flag.StringVarP(&flagImageName, "name", "n", "",
		"The name to give the built image (default: randomly generated)")
}

func usage() {
	fmt.Println(strings.TrimSpace(`
Usage: dbuild [options] <Dockerfile> <root path> <output file>

Builds a Docker image from the given Dockerfile, with the root of the build
context at the given root path.  The built image is then exported into the
given output file.

Options:`))
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()

	if flag.NArg() < 3 {
		usage()
	}

	dockerfilePath, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Errorf("Error finding absolute path (1): %s", err)
		return
	}

	rootPath, err := filepath.Abs(flag.Arg(1))
	if err != nil {
		log.Errorf("Error finding absolute path (2): %s", err)
		return
	}

	outputPath := flag.Arg(2)

	log.Infof("Started")

	client, err := docker.NewClient(flagEndpoint)
	if err != nil {
		log.Errorf("Error creating Docker client: %s", err)
		return
	}

	err = client.Ping()
	if err != nil {
		log.Errorf("Error pinging Docker client: %s", err)
		return
	}

	log.Infof("Connected to Docker client")

	// Create the output buffer.
	outf, err := os.Create(outputPath)
	if err != nil {
		log.Errorf("Error creating output file: %s", err)
		return
	}
	defer outf.Close()

	// Create our build context tar file.
	buildctx, err := ioutil.TempFile("", "dbuild-ctx")
	if err != nil {
		log.Errorf("Error creating temporary build context file: %s", err)
		return
	}
	defer buildctx.Close()

	tr := tar.NewWriter(buildctx)

	// Write the Dockerfile into the build context
	dockerfile, err := os.Open(dockerfilePath)
	if err != nil {
		log.Errorf("Error opening Dockerfile: %s", err)
		return
	}

	err = writeFileTo(tr, dockerfile, "Dockerfile")
	if err != nil {
		log.Errorf("Error writing Dockerfile to build context: %s", err)
		return
	}

	// Recursively search the root for other files and add those.
	log.Infof("Adding files to build context...")

	rootDockerfilePath := filepath.Join(rootPath, "Dockerfile")
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		// If there's an error, we just return it and abort the walk.
		if err != nil {
			return err
		}

		// Ignore paths that start with '.'
		if len(path) > 0 && path[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Just descend into directories.
		if info.IsDir() {
			return nil
		}

		// We skip this file if the path is the same as our Dockerfile, and if
		// it's in the root directory.  This is to avoid having two Dockerfiles
		// in the root.
		if path == rootDockerfilePath {
			return nil
		}

		// Find the path relative to the root.
		rel, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		// Open the file.
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		// This is the VT100 escape sequence for "clear line".
		fmt.Printf("\r\033[2KAdding file: %s", rel)

		// Add this file to the TAR file.
		err = writeFileTo(tr, f, rel)

		// TODO: ensure that the file we're adding isn't our Dockerfile?

		// log.Infof(path)

		return err
	})

	// Clear line
	fmt.Printf("\r\033[2K")
	log.Infof("Finished adding build context")

	err = tr.Close()
	if err != nil {
		log.Errorf("Error finalizing build context: %s", err)
		return
	}

	// Need to rewind our tar file handle to the beginning.
	_, err = buildctx.Seek(0, 0)
	if err != nil {
		log.Errorf("Error seeking to beginning of build context: %s", err)
		return
	}

	// Get the image name.
	if len(flagImageName) == 0 {
		flagImageName = randString(20)
	}
	log.Infof("Using image name: %s", flagImageName)

	// Set up build options.  Note that the escape at the end resets the
	// terminal color.
	output := NewLineStreamer(os.Stdout, "   [build] ", "\x1b[0m")
	opts := docker.BuildImageOptions{
		Name:         flagImageName,
		InputStream:  buildctx,
		OutputStream: output,

		// From program options.
		NoCache:             flagNoCache,
		RmTmpContainer:      flagRm,
		ForceRmTmpContainer: flagForceRm,
	}

	// Send everything off for building
	log.Infof("Starting to build image, please wait...")
	err = client.BuildImage(opts)
	if err != nil {
		log.Errorf("Error building image: %s", err)
		return
	}
	log.Infof("Finished building image")

	// Inspect the image to get information.
	img, err := client.InspectImage(flagImageName)
	if err != nil {
		log.Errorf("Error inspecting image: %s", err)
		return
	}

	log.Infof("Image built (size = %d)", img.Size)

	// Export the image to our output file.
	exportOpts := docker.ExportImageOptions{
		Name:         flagImageName,
		OutputStream: outf,
	}

	log.Infof("Exporting built image, please wait...")
	err = client.ExportImage(exportOpts)
	if err != nil {
		log.Errorf("Error exporting image: %s", err)
		return
	}
	log.Infof("Finished exporting")

	// Optionally remove the image.
	if flagRmAfter {
		log.Infof("Removing image...")
		err = client.RemoveImage(flagImageName)
		if err != nil {
			log.Errorf("Error removing image: %s", err)
			return
		}
		log.Infof("Image removed")
	}

	log.Infof("Completed successfully")
}

// Write the contents of a file to a TAR file.
func writeFileTo(tarfile *tar.Writer, f *os.File, name string) error {
	info, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

	header.Name = name

	err = tarfile.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tarfile, f)
	return err
}

func randString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
