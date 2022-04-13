package executables

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	dockerPath      = "docker"
	defaultRegistry = "public.ecr.aws"
)

type Docker struct {
	Executable
}

func NewDocker(executable Executable) *Docker {
	return &Docker{Executable: executable}
}

func (d *Docker) GetDockerLBPort(ctx context.Context, clusterName string) (port string, err error) {
	clusterLBName := fmt.Sprintf("%s-lb", clusterName)
	if stdout, err := d.Execute(ctx, "port", clusterLBName, "6443/tcp"); err != nil {
		return "", err
	} else {
		return strings.Split(stdout.String(), ":")[1], nil
	}
}

func (d *Docker) PullImage(ctx context.Context, image string) error {
	logger.V(2).Info("Pulling docker image", "image", image)
	if _, err := d.Execute(ctx, "pull", image); err != nil {
		return err
	} else {
		return nil
	}
}

func (d *Docker) SetUpCLITools(ctx context.Context, image string) error {
	logger.V(1).Info("Setting up cli docker dependencies")
	if err := d.PullImage(ctx, image); err != nil {
		return err
	} else {
		return nil
	}
}

func (d *Docker) Version(ctx context.Context) (int, error) {
	cmdOutput, err := d.Execute(ctx, "version", "--format", "{{.Client.Version}}")
	if err != nil {
		return 0, fmt.Errorf("please check if docker is installed and running %v", err)
	}
	dockerVersion := strings.TrimSpace(cmdOutput.String())
	versionSplit := strings.Split(dockerVersion, ".")
	installedMajorVersion := versionSplit[0]
	installedMajorVersionInt, err := strconv.Atoi(installedMajorVersion)
	if err != nil {
		return 0, err
	}
	return installedMajorVersionInt, nil
}

func (d *Docker) AllocatedMemory(ctx context.Context) (uint64, error) {
	cmdOutput, err := d.Execute(ctx, "info", "--format", "'{{json .MemTotal}}'")
	if err != nil {
		return 0, fmt.Errorf("please check if docker is installed and running %v", err)
	}
	totalMemory := cmdOutput.String()
	totalMemory = totalMemory[1 : len(totalMemory)-2]
	return strconv.ParseUint(totalMemory, 10, 64)
}

func (d *Docker) CgroupVersion(ctx context.Context) (int, error) {
	cmdOutput, err := d.Execute(ctx, "info", "--format", "'{{json .CgroupVersion}}'")
	if err != nil {
		return 0, fmt.Errorf("please check if docker is installed and running %v", err)
	}
	cgroupVersion := strings.TrimSpace(cmdOutput.String())
	cgroupVersion = strings.Trim(cgroupVersion, "\"'")
	version, err := strconv.Atoi(cgroupVersion)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func (d *Docker) TagImage(ctx context.Context, image string, endpoint string) error {
	localImage := strings.ReplaceAll(image, defaultRegistry, endpoint)
	logger.Info("Tagging image", "image", image, "local image", localImage)
	if _, err := d.Execute(ctx, "tag", image, localImage); err != nil {
		return err
	}
	return nil
}

func (d *Docker) PushImage(ctx context.Context, image string, endpoint string) error {
	localImage := strings.ReplaceAll(image, defaultRegistry, endpoint)
	logger.Info("Pushing", "image", localImage)
	if _, err := d.Execute(ctx, "push", localImage); err != nil {
		return err
	}
	return nil
}

func (d *Docker) Login(ctx context.Context, endpoint, username, password string) error {
	params := []string{"login", endpoint, "--username", username, "--password-stdin"}
	logger.Info(fmt.Sprintf("Logging in to docker registry %s", endpoint))
	_, err := d.ExecuteWithStdin(ctx, []byte(password), params...)
	return err
}

func (d *Docker) LoadFromFile(ctx context.Context, filepath string) error {
	if _, err := d.Execute(ctx, "load", "-i", filepath); err != nil {
		return fmt.Errorf("loading images from file: %v", err)
	}

	return nil
}

func (d *Docker) SaveToFile(ctx context.Context, filepath string, images ...string) error {
	params := make([]string, 0, 3+len(images))
	params = append(params, "save", "-o", filepath)
	params = append(params, images...)

	if _, err := d.Execute(ctx, params...); err != nil {
		return fmt.Errorf("saving images to file: %v", err)
	}

	return nil
}
