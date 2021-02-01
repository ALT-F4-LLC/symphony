package cli

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/sirupsen/logrus"
)

// ImageNewOptions : options for creating an image
type ImageNewOptions struct {
	Description *string
	ManagerAddr *string
	Name        *string
	Path        *string
}

// ImageNew : initializes a service for use
func ImageNew(opts ImageNewOptions) {
	if opts.ManagerAddr == nil {
		logrus.Fatal("Missing --manager-addr option. Check --help for more.")
	}

	if opts.Description == nil {
		logrus.Fatal("Missing description argument. Check --help for more.")
	}

	if opts.Name == nil {
		logrus.Fatal("Missing name argument. Check --help for more.")
	}

	if opts.Path == nil {
		logrus.Fatal("Missing path argument. Check --help for more.")
	}

	path, err := filepath.Abs(*opts.Path)

	if err != nil {
		logrus.Fatal("Missing file from path provided.")
	}

	file, err := os.Open(path)

	if err != nil {
		logrus.Fatal(err)
	}

	defer file.Close()

	fileBase := filepath.Base(path)

	conn, err := utils.NewClientConnTcp(*opts.ManagerAddr)

	if err != nil {
		logrus.Fatal(err)
	}

	defer conn.Close()

	timeout := utils.ContextTimeout * 12

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	c := api.NewManagerClient(conn)

	stream, err := c.NewImage(ctx)

	if err != nil {
		logrus.Fatal(err)
	}

	details := &api.RequestNewImage{
		Data: &api.RequestNewImage_Details{
			Details: &api.RequestNewImageDetails{
				Description: *opts.Description,
				File:        fileBase,
				Name:        *opts.Name,
			},
		},
	}

	sendErr := stream.Send(details)

	if sendErr != nil {
		logrus.Fatal(sendErr)
	}

	buffer := make([]byte, 1024)

	reader := bufio.NewReader(file)

	for {
		n, err := reader.Read(buffer)

		if err == io.EOF {
			break
		}

		if err != nil {
			logrus.Fatal(err)
		}

		chunk := &api.RequestNewImage{
			Data: &api.RequestNewImage_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		sendErr := stream.Send(chunk)

		if sendErr != nil {
			logrus.Fatal(sendErr)
		}
	}

	res, err := stream.CloseAndRecv()

	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info(res)
}
