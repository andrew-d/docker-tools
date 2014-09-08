package main

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/fsouza/go-dockerclient"
	flag "github.com/ogier/pflag"
)

var (
	flagNoCache  bool
	flagRm       bool
	flagForceRm  bool
	flagEndpoint string
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
}

func usage() {
	fmt.Println("Usage: dbuild [options] <Dockerfile> <root path> <output file>")
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

	imageName := "FIXME"

	// Set up build options.
	output := NewLineStreamer(os.Stdout, "   [build] ", "")
	opts := docker.BuildImageOptions{
		Name:         imageName,
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
	img, err := client.InspectImage(imageName)
	if err != nil {
		log.Printf("Error inspecting image: %s", err)
		return
	}

	log.Printf("Image built (size = %d)", img.Size)

	// Export the image to our output file.
	exportOpts := docker.ExportImageOptions{
		Name:         imageName,
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
