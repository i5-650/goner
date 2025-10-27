package cmd

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

var catLayerNum int

// fsCatCmd represents "goner fs cat"
var fsCatCmd = &cobra.Command{
	Use:   "cat [image] [path]",
	Short: "Affiche le contenu d’un fichier dans un layer d’image",
	Long: `Affiche le contenu exact d’un fichier stocké dans un layer Docker/OCI.

Exemples :
  goner fs cat alpine:latest -l 3 /etc/os-release`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef := args[0]
		filePath := strings.TrimPrefix(args[1], "/")

		if catLayerNum <= 0 {
			return fmt.Errorf("vous devez spécifier un numéro de layer >= 1 via --layer ou -l")
		}

		return catFileFromLayer(imageRef, catLayerNum, filePath)
	},
}

func init() {
	filesystemCmd.AddCommand(fsCatCmd)
	fsCatCmd.Flags().IntVarP(&catLayerNum, "layer", "l", 1, "Numéro du layer à explorer (commence à 1)")
}

func catFileFromLayer(refStr string, layerIndex int, filePath string) error {
	ref, err := name.ParseReference(refStr)
	if err != nil {
		return fmt.Errorf("error parsing image: %w", err)
	}

	fmt.Printf("Pulling image %s...\n", ref.Name())
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return fmt.Errorf("error pulling image: %w", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return fmt.Errorf("unable to read layers: %w", err)
	}

	if layerIndex > len(layers) {
		return fmt.Errorf("the image only contains %d layers (you asked for layer %d)", len(layers), layerIndex)
	}

	layer := layers[layerIndex-1]
	rc, err := layer.Uncompressed()
	if err != nil {
		return fmt.Errorf("couldn't uncompress the layer: %w", err)
	}
	defer rc.Close()

	tr := tar.NewReader(rc)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error reading tar: %v", err)
			break
		}

		// Normalise le chemin
		name := strings.TrimPrefix(h.Name, "./")
		name = strings.TrimSuffix(name, "/")

		if name == filePath {
			if h.FileInfo().IsDir() {
				return fmt.Errorf("%s is a dir", filePath)
			}

			fmt.Printf("=== %s (layer #%d) ===\n", filePath, layerIndex)
			_, err := io.Copy(os.Stdout, tr)
			if err != nil {
				return fmt.Errorf("error reading file content: %w", err)
			}
			fmt.Println()
			return nil
		}
	}

	return fmt.Errorf("%s not found in layer #%d", filePath, layerIndex)
}
