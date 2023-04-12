package executables

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
)

// Temporary: Curated packages dev and prod accounts are currently hard coded
// This is because there is no mechanism to extract these values as of now.
const (
	dockerPath        = "docker"
	defaultRegistry   = "public.ecr.aws"
	packageProdDomain = "783794618700.dkr.ecr.us-west-2.amazonaws.com"
	packageDevDomain  = "857151390494.dkr.ecr.us-west-2.amazonaws.com"
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

func (d *Docker) TagImage(ctx context.Context, image string, endpoint string) error {
	replacer := strings.NewReplacer(defaultRegistry, endpoint, packageProdDomain, endpoint, packageDevDomain, endpoint)
	localImage := replacer.Replace(image)
	logger.Info("Tagging image", "image", image, "local image", localImage)
	if _, err := d.Execute(ctx, "tag", image, localImage); err != nil {
		return err
	}
	return nil
}

func (d *Docker) PushImage(ctx context.Context, image string, endpoint string) error {
	replacer := strings.NewReplacer(defaultRegistry, endpoint, packageProdDomain, endpoint, packageDevDomain, endpoint)
	localImage := replacer.Replace(image)
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

func (d *Docker) Run(ctx context.Context, image string, name string, cmd []string, flags ...string) error {
	params := []string{"run", "-d", "-i"}
	params = append(params, flags...)
	params = append(params, "--name", name, image)
	params = append(params, cmd...)

	if _, err := d.Execute(ctx, params...); err != nil {
		return fmt.Errorf("running docker container %s with image %s: %v", name, image, err)
	}
	return nil
}

func (d *Docker) ForceRemove(ctx context.Context, name string) error {
	params := []string{"rm", "-f", name}

	if _, err := d.Execute(ctx, params...); err != nil {
		return fmt.Errorf("force removing docker container %s: %v", name, err)
	}
	return nil
}

// CheckContainerExistence checks whether a Docker container with the provided name exists
// It returns true if a container with the name exists, false if it doesn't and an error if it encounters some other error.
func (d *Docker) CheckContainerExistence(ctx context.Context, name string) (bool, error) {
	params := []string{"container", "inspect", name}

	_, err := d.Execute(ctx, params...)
	if err == nil {
		return true, nil
	} else if strings.Contains(err.Error(), "No such container") {
		return false, nil
	}

	return false, fmt.Errorf("checking if a docker container with name %s exists: %v", name, err)
}
