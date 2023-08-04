package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"io"
	"os"
	"path/filepath"
)

func main() {
	app := fiber.New()
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	rb := NewRingBuffer(20000)
	app.Static("/code/", "./code")

	err := cloneRepo("https://github.com/heroku/node-js-getting-started", "./code")
	if err != nil {
	}

	// POST Static Resources
	app.Post("/code/+", func(ctx *fiber.Ctx) error {
		// The wildcard value (directory or file)
		wildcard := ctx.Params("+")

		// Ensure wildcard is safe (e.g., not attempting to write outside allowed directory)
		if filepath.Clean(wildcard) != wildcard || filepath.IsAbs(wildcard) {
			return ctx.Status(fiber.StatusBadRequest).SendString("Invalid file path provided.")
		}

		// Get the POST request body (gzipped content)
		gzippedContent := ctx.Body()

		// Decompress the gzipped content
		reader, err := gzip.NewReader(bytes.NewReader(gzippedContent))
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).SendString("Failed to decompress content.")
		}
		defer reader.Close()

		// Create the destination directory if it doesn't exist
		destDir := filepath.Dir(wildcard)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString("Failed to create directory.")
		}

		// Write the decompressed content to the desired location
		file, err := os.Create(wildcard)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString("Failed to create file.")
		}
		defer file.Close()

		_, err = io.Copy(file, reader)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString("Failed to write file.")
		}

		return ctx.SendString("File successfully uploaded and extracted.")
	})

	//app.Post("/clone", func(ctx *fiber.Ctx) error {
	//	// Clone the repo
	//
	//
	//	// Build the Docker image
	//	if err := buildDockerImage("./temp_repo", "swizzle_dev"); err != nil {
	//		return ctx.Status(fiber.StatusInternalServerError).SendString(err.Error())
	//	}
	//
	//	return ctx.SendString("Done!")
	//})

	// Endpoint to fetch the logs
	app.Get("/logs", func(ctx *fiber.Ctx) error {
		return ctx.JSON(rb.Get())
	})
}

// GetLogs return logs from the container io.ReadCloser. It's the caller duty
// duty to do a stdcopy.StdCopy. Any other method might render unknown
// unicode character as log output has both stdout and stderr. That starting
// has info if that line is stderr or stdout.
// IMPORTANT: https://stackoverflow.com/questions/45756681/grab-output-from-container-in-docker-sdk
func GetLogs(ctx context.Context, cli *client.Client, contName string) (logOutput io.ReadCloser) {
	options := types.ContainerLogsOptions{ShowStdout: true}

	out, err := cli.ContainerLogs(ctx, contName, options)
	if err != nil {
		panic(err)
	}

	return out
}

func buildDockerImage(dockerfileDir, imageName string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	defer cli.Close()

	buildContext, _ := os.Open(dockerfileDir)
	defer buildContext.Close()

	buildOptions := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{imageName},
	}
	buildResponse, err := cli.ImageBuild(context.Background(), buildContext, buildOptions)
	if err != nil {
		return err
	}
	defer buildResponse.Body.Close()

	// This is to read the build response for logging or printing purposes.
	// It streams the build process logs.
	_, err = io.Copy(os.Stdout, buildResponse.Body)
	return err
}

func cloneRepo(url, path string) error {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	return err
}
