package cmd

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

var layerNum int

// filesystemCmd represents the filesystem command
var filesystemCmd = &cobra.Command{
	Use:     "filesystem [image]",
	Aliases: []string{"fs"},
	Short:   "Explore the filesystem of a specific layer of an image",
	Long: `Displays the content of a layer in an OCI/Docker image.

Examples:
  goner filesystem alpine:latest --layer 2
  goner fs alpine:latest -l 2`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef := args[0]

		if layerNum <= 0 {
			return fmt.Errorf("you must specify a layer number >= 1 using --layer or -l")
		}

		return exploreFilesystem(imageRef, layerNum)
	},
}

func init() {
	rootCmd.AddCommand(filesystemCmd)

	// Short and long flag
	filesystemCmd.Flags().IntVarP(&layerNum, "layer", "l", 1, "Layer number to explore (starting at 1)")
	filesystemCmd.AddCommand(fsCatCmd)
}

// exploreFilesystem downloads the image and displays the content of the selected layer
func exploreFilesystem(refStr string, layerIndex int) error {
	ref, err := name.ParseReference(refStr)
	if err != nil {
		return fmt.Errorf("error parsing image: %w", err)
	}

	fmt.Printf("Pulling image %s...\n", ref.Name())
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return fmt.Errorf("unable to read layers: %w", err)
	}

	if layerIndex > len(layers) {
		return fmt.Errorf("the image only contains %d layers (you requested layer %d)", len(layers), layerIndex)
	}

	layer := layers[layerIndex-1]
	rc, err := layer.Uncompressed()
	if err != nil {
		return fmt.Errorf("failed to uncompress layer: %w", err)
	}
	defer rc.Close()

	tr := tar.NewReader(rc)
	fmt.Printf("Content of layer #%d of %s:\n", layerIndex, ref.Name())

	count := 0
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading tar: %v\n", err)
			break
		}

		// Ignore whiteout files
		if strings.HasPrefix(header.Name, ".wh.") {
			continue
		}

		mode := header.FileInfo().Mode()
		size := header.FileInfo().Size()
		fmt.Printf("%-10s %8d  %s\n", mode, size, header.Name)
		count++
	}

	if count == 0 {
		fmt.Println("(No files found in this layer)")
	}
	return nil
}
