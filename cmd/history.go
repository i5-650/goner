package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "A brief description of your command",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		imageRef := args[0]
		listHistory(imageRef)
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
}

func listHistory(imageRef string) {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		log.Fatalf("Invalid image reference: %v", err)
	}

	img, err := remote.Image(ref)
	if err != nil {
		log.Fatalf("Failed to retrieve image: %v", err)
	}

	cfgFile, err := img.ConfigFile()
	if err != nil {
		log.Fatalf("Failed to retrieve config file: %v", err)
	}

	layers, err := img.Layers()
	if err != nil {
		log.Fatalf("Error reading layers: %v", err)
	}

	var totalCompressed int64
	layerIndex := 0
	for i, h := range cfgFile.History {
		marker := " "
		if !h.EmptyLayer {
			marker = "•" // true layer
		}

		cmd := prettifyCommand(h.CreatedBy)
		if cmd == "" {
			cmd = "(no command)"
		}

		fmt.Printf("[%02d] %s %s\n", i+1, marker, cmd)

		if !h.EmptyLayer {
			if layerIndex < len(layers) {
				digest, _ := layers[layerIndex].Digest()
				size, _ := layers[layerIndex].Size()
				totalCompressed += size
				fmt.Printf("\t↳ Layer %d %s (%.2f MB)\n", layerIndex+1, digest, float64(size)/(1024*1024))
			}
			layerIndex++
		}
	}
	fmt.Printf("Total size (compressed): %.2f MB\n", float64(totalCompressed)/(1024*1024))
}

func prettifyCommand(cmd string) string {
	cmd = strings.TrimPrefix(cmd, "/bin/sh -c ")
	cmd = strings.TrimSpace(cmd)

	// replace multiples space by one
	re := regexp.MustCompile(`\s+`)
	cmd = re.ReplaceAllString(cmd, " ")

	if strings.Contains(cmd, " --") {
		cmd = strings.ReplaceAll(cmd, " --", " \\\n\t\t--")
	}

	cmd = strings.ReplaceAll(cmd, ";", ";\n\t")

	return cmd
}
