package main

import (
	"flag"
	"fmt"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {

	inputDir := flag.String("input", "", "Input directory containing photos (required)")
	outputDir := flag.String("output", "", "Output directory containing photos (required)")
	recursive := flag.Bool("recursive", false, "Process subdirectories recursively")
	threads := flag.Int("threads", 4, "Number of concurrent processing threads")
	flag.Parse()

	// Validate required flags
	if *inputDir == "" || *outputDir == "" {
		fmt.Println("Error: Both input and output directories are required")
		fmt.Println("\nUsage:")
		fmt.Println("  stripper -input <input_dir> -output <output_dir> [-recursive] [-threads N]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	//Create out put directory if it does not exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Create a channel for image processing tasks
	tasks := make(chan string)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go worker(tasks, *outputDir, &wg)
	}

	// Process count tracking
	var processed, failed int
	var processLock sync.Mutex

	err := filepath.Walk(*inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories unless recursive flag is set
		if info.IsDir() {
			if path != *inputDir && !*recursive {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file is an image
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			tasks <- path
			processLock.Lock()
			processed++
			processLock.Unlock()
		}

		return nil
	})

	// Close the tasks channel after all files are queued
	close(tasks)

	// Wait for all workers to finish
	wg.Wait()

	// Print summary
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
	}
	fmt.Printf("\nProcessing complete:\n")
	fmt.Printf("Processed: %d images\n", processed)
	fmt.Printf("Failed: %d images\n", failed)
}

func worker(tasks <-chan string, outputDir string, wg *sync.WaitGroup) {
	defer wg.Done()

	for inputPath := range tasks {
		// Simply use the base filename with clean_ prefix
		outputPath := filepath.Join(outputDir, "clean_"+filepath.Base(inputPath))

		fmt.Printf("Processing: %s\n", inputPath)

		if err := stripMetadata(inputPath, outputPath); err != nil {
			fmt.Printf("Error processing %s: %v\n", inputPath, err)
			continue
		}
	}
}
