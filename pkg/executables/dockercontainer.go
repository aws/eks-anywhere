package executables

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

type DockerClient interface {
	Login(ctx context.Context, endpoint, username, password string) error
	PullImage(ctx context.Context, image string) error
	Execute(ctx context.Context, args ...string) (stdout bytes.Buffer, err error)
}

type dockerContainer struct {
	image               string
	workingDir          string
	mountDirs           []string
	containerName       string
	dockerClient        DockerClient
	initOnce, closeOnce sync.Once
	*retrier.Retrier
}

func newDockerContainer(image, workingDir string, mountDirs []string, dockerClient DockerClient) *dockerContainer {
	return &dockerContainer{
		image:         image,
		workingDir:    workingDir,
		mountDirs:     mountDirs,
		containerName: containerNamePrefix + strconv.FormatInt(time.Now().UnixNano(), 10),
		dockerClient:  dockerClient,
		Retrier:       retrier.NewWithMaxRetries(maxRetries, backOffPeriod),
	}
}

func NewDockerContainerCustomBinary(docker DockerClient) *dockerContainer {
	return &dockerContainer{
		dockerClient: docker,
	}
}

func (d *dockerContainer) Init(ctx context.Context) error {
	var err error
	d.initOnce.Do(func() {
		err = d.Retry(func() error {
			return d.dockerClient.PullImage(ctx, d.image)
		})
		if err != nil {
			return
		}

		var absWorkingDir string
		absWorkingDir, err = filepath.Abs(d.workingDir)
		if err != nil {
			err = fmt.Errorf("getting abs path for mount dir: %v", err)
			return
		}

		params := []string{"run", "-d", "--name", d.containerName, "--network", "host", "-w", absWorkingDir, "-v", "/var/run/docker.sock:/var/run/docker.sock"}

		for _, m := range d.mountDirs {
			var absMountDir string
			absMountDir, err = filepath.Abs(m)
			if err != nil {
				err = fmt.Errorf("getting abs path for mount dir: %v", err)
				return
			}
			params = append(params, "-v", fmt.Sprintf("%[1]s:%[1]s", absMountDir))
		}

		// start container and keep it running in the background
		logger.V(3).Info("Initializing long running container", "name", d.containerName, "image", d.image)
		params = append(params, "--entrypoint", "sleep", d.image, "infinity")
		_, err = d.dockerClient.Execute(ctx, params...)
	})

	return err
}

func (d *dockerContainer) ContainerName() string {
	return d.containerName
}

func (d *dockerContainer) Close(ctx context.Context) error {
	if d == nil {
		return nil
	}

	var err error
	d.closeOnce.Do(func() {
		logger.V(3).Info("Cleaning up long running container", "name", d.containerName)
		_, err = d.dockerClient.Execute(ctx, "rm", "-f", "-v", d.containerName)
	})

	return err
}
