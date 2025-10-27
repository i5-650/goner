package cmd

import (
	"fmt"
	"log"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

// layerCmd represents the layer command
var layerCmd = &cobra.Command{
	Use:   "layer <image:tag>",
	Short: "Show layer of the container",
	Long:  `Show layer of the container that will be pulled from the registry`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		imageRef := args[0]
		listLayers(imageRef)
	},
}

func init() {
	rootCmd.AddCommand(layerCmd)
}

func listLayers(imageRef string) {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		log.Fatalf("Invalid image reference: %v", err)
	}

	img, err := remote.Image(ref)
	if err != nil {
		log.Fatalf("Failed to retrieve image: %v", err)
	}

	layers, err := img.Layers()
	if err != nil {
		log.Fatalf("Error reading layers: %v", err)
	}
	var totalCompressed int64

	fmt.Printf("Image %s contains %d layers:\n", ref.Name(), len(layers))
	for i, layer := range layers {
		digest, _ := layer.Digest()
		size, _ := layer.Size()
		totalCompressed += size
		fmt.Printf("  • Layer %d : %s (%.2f MB)\n", i+1, digest.String(), float64(size)/(1024*1024))
	}
	fmt.Printf("Total size (compressed): %.2f MB\n", float64(totalCompressed)/(1024*1024))
}
