package main

import (
	"archive/tar"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsouza/go-dockerclient"
	flag "github.com/ogier/pflag"
)

var (
	flagNoCache  bool
	flagRm       bool
	flagForceRm  bool
	flagEndpoint string
	flagImageName string
)

func init() {
	flag.BoolVar(&flagNoCache, "no-cache", false,
		"Do not use cache when building the image")
	flag.BoolVar(&flagRm, "rm", false,
		"Remove intermediate containers after a successful build")
	flag.BoolVar(&flagForceRm, "force-rm", false,
		"Always remove intermediate containers, even after unsuccessful builds")
	flag.StringVarP(&flagEndpoint, "endpoint", "e", "unix:///var/run/docker.sock",
		"How to connect to the Docker service")
	flag.StringVarP(&flagImageName, "name", "n", "",
		"The name to give the built image")
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

	dockerfilePath := flag.Arg(0)
	rootPath := flag.Arg(1)
	outputPath := flag.Arg(2)

	log.Println("Started")

	client, err := docker.NewClient(flagEndpoint)
	if err != nil {
		log.Printf("Error creating Docker client: %s", err)
		return
	}

	log.Println("Connected to Docker client")

	// Create the output buffer.
	outf, err := os.Create(outputPath)
	if err != nil {
		log.Printf("Error creating output file: %s", err)
		return
	}
	defer outf.Close()

	// Create our build context tar file.
	buildctx, err := ioutil.TempFile("", "dbuild-ctx")
	if err != nil {
		log.Printf("Error creating temporary build context file: %s", err)
		return
	}
	defer buildctx.Close()

	tr := tar.NewWriter(buildctx)

	// Write the Dockerfile into the build context
	dockerfile, err := os.Open(dockerfilePath)
	if err != nil {
		log.Printf("Error opening Dockerfile: %s", err)
		return
	}

	err = writeFileTo(tr, dockerfile)
	if err != nil {
		log.Printf("Error writing Dockerfile to build context: %s", err)
		return
	}

	// Recursively search the root for other files and add those.
	log.Println("Adding files to build context...")

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		// If there's an error, we just return it and abort the walk.
		if err != nil {
			return err
		}

		// TODO: add this file to the tar file
		// TODO: ensure that the file we're adding isn't our Dockerfile (don't want
		//		 to add it twice)

		// log.Println(path)

		return nil
	})

	log.Println("Finished adding build context")

	err = tr.Close()
	if err != nil {
		log.Printf("Error finalizing build context: %s", err)
		return
	}

	// Need to rewind our tar file handle to the beginning.
	_, err = buildctx.Seek(0, 0)
	if err != nil {
		log.Printf("Error seeking to beginning of build context: %s", err)
		return
	}

	// Get the image name.
	if len(flagImageName) == 0 {
		flagImageName = randString(20)
	}
	log.Printf("Using image name: %s", flagImageName)

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
	log.Println("Starting to build image, please wait...")
	err = client.BuildImage(opts)
	if err != nil {
		log.Printf("Error building image: %s", err)
		return
	}
	log.Println("Finished building image")

	// Inspect the image to get information.
	img, err := client.InspectImage(flagImageName)
	if err != nil {
		log.Printf("Error inspecting image: %s", err)
		return
	}

	log.Printf("Image built (size = %d)", img.Size)

	// Export the image to our output file.
	exportOpts := docker.ExportImageOptions{
		Name:         flagImageName,
		OutputStream: outf,
	}

	log.Println("Exporting built image, please wait...")
	err = client.ExportImage(exportOpts)
	if err != nil {
		log.Printf("Error exporting image: %s", err)
		return
	}
	log.Println("Finished exporting")

	log.Println("All done!")
}

// Write the contents of a file to a TAR file.
func writeFileTo(tarfile *tar.Writer, f *os.File) error {
	info, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

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
