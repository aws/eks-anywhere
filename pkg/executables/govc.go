package executables

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	govcPath             = "govc"
	govcUsernameKey      = "GOVC_USERNAME"
	govcPasswordKey      = "GOVC_PASSWORD"
	govcURLKey           = "GOVC_URL"
	govcInsecure         = "GOVC_INSECURE"
	govcTlsHostsFile     = "govc_known_hosts"
	govcTlsKnownHostsKey = "GOVC_TLS_KNOWN_HOSTS"
	vSphereUsernameKey   = "EKSA_VSPHERE_USERNAME"
	vSpherePasswordKey   = "EKSA_VSPHERE_PASSWORD"
	vSphereServerKey     = "VSPHERE_SERVER"
	byteToGiB            = 1073741824.0
	deployOptsFile       = "deploy-opts.json"
)

var requiredEnvs = []string{govcUsernameKey, govcPasswordKey, govcURLKey, govcInsecure}

//go:embed config/deploy-opts.json
var deployOpts []byte

type FolderType string

const (
	datastore     FolderType = "datastore"
	vm            FolderType = "vm"
	maxRetries               = 5
	backOffPeriod            = 5 * time.Second
)

type Govc struct {
	writer filewriter.FileWriter
	Executable
	retrier      *retrier.Retrier
	requiredEnvs *syncSlice
}

func NewGovc(executable Executable, writer filewriter.FileWriter) *Govc {
	envVars := newSyncSlice()
	envVars.append(requiredEnvs...)

	return &Govc{
		writer:       writer,
		Executable:   executable,
		retrier:      retrier.NewWithMaxRetries(maxRetries, backOffPeriod),
		requiredEnvs: envVars,
	}
}

func (g *Govc) exec(ctx context.Context, args ...string) (stdout bytes.Buffer, err error) {
	envMap, err := g.validateAndSetupCreds()
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed govc validations: %v", err)
	}

	return g.ExecuteWithEnv(ctx, envMap, args...)
}

func (g *Govc) Close(ctx context.Context) error {
	if g == nil {
		return nil
	}

	return g.Logout(ctx)
}

func (g *Govc) Logout(ctx context.Context) error {
	logger.V(3).Info("Logging out from current govc session")
	if _, err := g.exec(ctx, "session.logout"); err != nil {
		return fmt.Errorf("govc returned error when logging out: %v", err)
	}

	// Commands that skip cert verification will have a different session.
	// So we try to destroy it as well here to avoid leaving it orphaned
	if _, err := g.exec(ctx, "session.logout", "-k"); err != nil {
		return fmt.Errorf("govc returned error when logging out from session without cert verification: %v", err)
	}

	return nil
}

func (g *Govc) SearchTemplate(ctx context.Context, datacenter string, machineConfig *v1alpha1.VSphereMachineConfig) (string, error) {
	params := []string{"find", "-json", "/" + datacenter, "-type", "VirtualMachine", "-name", filepath.Base(machineConfig.Spec.Template)}
	templateResponse, err := g.exec(ctx, params...)
	if err != nil {
		return "", fmt.Errorf("error getting template: %v", err)
	}

	templateJson := templateResponse.String()
	templateJson = strings.TrimSuffix(templateJson, "\n")
	if templateJson == "null" || templateJson == "" {
		logger.V(2).Info(fmt.Sprintf("Template not found: %s", machineConfig.Spec.Template))
		return "", nil
	}

	templateInfo := make([]string, 0)
	if err = json.Unmarshal([]byte(templateJson), &templateInfo); err != nil {
		logger.V(2).Info(fmt.Sprintf("Failed unmarshalling govc response: %s, %v", templateJson, err))
		return "", nil
	}

	bTemplateFound := false
	var foundTemplate string
	for _, t := range templateInfo {
		if strings.HasSuffix(t, machineConfig.Spec.Template) {
			if bTemplateFound {
				return "", fmt.Errorf("specified template '%s' maps to multiple paths within the datacenter '%s'", machineConfig.Spec.Template, datacenter)
			}
			bTemplateFound = true
			foundTemplate = t
		}
	}
	if !bTemplateFound {
		logger.V(2).Info(fmt.Sprintf("Template '%s' not found", machineConfig.Spec.Template))
		return "", nil
	}

	return foundTemplate, nil
}

func (g *Govc) LibraryElementExists(ctx context.Context, library string) (bool, error) {
	response, err := g.exec(ctx, "library.ls", library)
	if err != nil {
		return false, fmt.Errorf("govc failed getting library to check if it exists: %v", err)
	}

	return response.Len() > 0, nil
}

type libElement struct {
	ContentVersion string `json:"content_version"`
}

func (g *Govc) GetLibraryElementContentVersion(ctx context.Context, element string) (string, error) {
	response, err := g.exec(ctx, "library.info", "-json", element)
	if err != nil {
		return "", fmt.Errorf("govc failed getting library element info: %v", err)
	}
	elementInfoJson := response.String()
	elementInfoJson = strings.TrimSuffix(elementInfoJson, "\n")
	if elementInfoJson == "null" {
		return "-1", nil
	}

	elementInfo := make([]libElement, 0)
	err = yaml.Unmarshal([]byte(elementInfoJson), &elementInfo)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling library element info: %v", err)
	}

	if len(elementInfo) == 0 {
		return "", fmt.Errorf("govc failed to return element info for library element %v", element)
	}
	return elementInfo[0].ContentVersion, nil
}

func (g *Govc) DeleteLibraryElement(ctx context.Context, element string) error {
	_, err := g.exec(ctx, "library.rm", element)
	if err != nil {
		return fmt.Errorf("govc failed deleting library item: %v", err)
	}

	return nil
}

func (g *Govc) ResizeDisk(ctx context.Context, datacenter, template, diskName string, diskSizeInGB int) error {
	_, err := g.exec(ctx, "vm.disk.change", "-dc", datacenter, "-vm", template, "-disk.name", diskName, "-size", strconv.Itoa(diskSizeInGB)+"G")
	if err != nil {
		return fmt.Errorf("failed to resize disk %s: %v", diskName, err)
	}
	return nil
}

func (g *Govc) DevicesInfo(ctx context.Context, datacenter, template string) (interface{}, error) {
	response, err := g.exec(ctx, "device.info", "-dc", datacenter, "-vm", template, "-json")
	if err != nil {
		return nil, fmt.Errorf("error getting template device information: %v", err)
	}

	var devicesInfo map[string]interface{}
	err = yaml.Unmarshal(response.Bytes(), &devicesInfo)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling devices info: %v", err)
	}
	return devicesInfo["Devices"], nil
}

func (g *Govc) TemplateHasSnapshot(ctx context.Context, template string) (bool, error) {
	envMap, err := g.validateAndSetupCreds()
	if err != nil {
		return false, fmt.Errorf("failed govc validations: %v", err)
	}

	params := []string{"snapshot.tree", "-vm", template}
	snap, err := g.ExecuteWithEnv(ctx, envMap, params...)
	if err != nil {
		return false, fmt.Errorf("failed to get snapshot details: %v", err)
	}
	if snap.String() == "" {
		return false, nil
	}
	return true, nil
}

type datastoreResponse struct {
	Datastores []types.Datastores `json:"Datastores"`
}

func (g *Govc) GetWorkloadAvailableSpace(ctx context.Context, datastore string) (float64, error) {
	envMap, err := g.validateAndSetupCreds()
	if err != nil {
		return 0, fmt.Errorf("failed govc validations: %v", err)
	}

	params := []string{"datastore.info", "-json=true", datastore}
	result, err := g.ExecuteWithEnv(ctx, envMap, params...)
	if err != nil {
		return 0, fmt.Errorf("error getting datastore info: %v", err)
	}

	response := &datastoreResponse{}
	err = json.Unmarshal(result.Bytes(), response)
	if err != nil {
		return -1, nil
	}
	if len(response.Datastores) > 0 {
		freeSpace := response.Datastores[0].Info.FreeSpace
		spaceGiB := freeSpace / byteToGiB
		return spaceGiB, nil
	}
	return 0, fmt.Errorf("error getting datastore available space response: %v", err)
}

func (g *Govc) CreateLibrary(ctx context.Context, datastore, library string) error {
	if _, err := g.exec(ctx, "library.create", "-ds", datastore, library); err != nil {
		return fmt.Errorf("error creating library %s: %v", library, err)
	}
	return nil
}

func (g *Govc) DeployTemplateFromLibrary(ctx context.Context, templateDir, templateName, library, datacenter, datastore, resourcePool string, resizeBRDisk bool) error {
	logger.V(4).Info("Deploying template", "dir", templateDir, "templateName", templateName)
	if err := g.deployTemplate(ctx, library, templateName, templateDir, datacenter, datastore, resourcePool); err != nil {
		return err
	}

	if resizeBRDisk {
		// Get devices information template to identify second disk properly
		logger.V(4).Info("Getting devices info for template")
		devicesInfo, err := g.DevicesInfo(ctx, datacenter, templateName)
		if err != nil {
			return err
		}
		// For 1.22 we switched to using one disk for BR, so it for now as long as the boolean is set, and we only see
		// one disk, we can assume this is for 1.22. This loop would need to change if that assumption changes
		// in the future, but 1.20 and 1.21 are still using dual disks which is why we need to check for the second
		// disk first. Since this loop will get all kinds of devices and not just hard disks, we need to do these
		// checks based on the label.
		disk1 := ""
		disk2 := ""
		for _, deviceInfo := range devicesInfo.([]interface{}) {
			deviceMetadata := deviceInfo.(map[string]interface{})["DeviceInfo"]
			deviceLabel := deviceMetadata.(map[string]interface{})["Label"].(string)
			// Get the name of the hard disk and resize the disk to 20G
			if strings.EqualFold(deviceLabel, "Hard disk 1") {
				disk1 = deviceInfo.(map[string]interface{})["Name"].(string)
			} else if strings.EqualFold(deviceLabel, "Hard disk 2") {
				disk2 = deviceInfo.(map[string]interface{})["Name"].(string)
				break
			}
		}
		diskName := ""
		var diskSizeInGB int
		if disk2 != "" {
			logger.V(4).Info("Resizing disk 2 of template to 20G")
			diskName = disk2
			diskSizeInGB = 20
		} else if disk1 != "" {
			logger.V(4).Info("Resizing disk 1 of template to 22G")
			diskName = disk1
			diskSizeInGB = 22
		} else {
			return fmt.Errorf("template %v is not valid as there are no associated disks", templateName)
		}

		err = g.ResizeDisk(ctx, datacenter, templateName, diskName, diskSizeInGB)
		if err != nil {
			return fmt.Errorf("error resizing disk %v to 20G: %v", diskName, err)
		}
	}

	templateFullPath := filepath.Join(templateDir, templateName)

	logger.V(4).Info("Taking template snapshot", "templateName", templateFullPath)
	if err := g.createVMSnapshot(ctx, datacenter, templateFullPath); err != nil {
		return err
	}

	logger.V(4).Info("Marking vm as template", "templateName", templateFullPath)
	if err := g.markVMAsTemplate(ctx, datacenter, templateFullPath); err != nil {
		return err
	}

	return nil
}

func (g *Govc) ImportTemplate(ctx context.Context, library, ovaURL, name string) error {
	logger.V(4).Info("Importing template", "ova", ovaURL, "templateName", name)
	if _, err := g.exec(ctx, "library.import", "-k", "-pull", "-n", name, library, ovaURL); err != nil {
		return fmt.Errorf("error importing template: %v", err)
	}
	return nil
}

func (g *Govc) deployTemplate(ctx context.Context, library, templateName, deployFolder, datacenter, datastore, resourcePool string) error {
	envMap, err := g.validateAndSetupCreds()
	if err != nil {
		return fmt.Errorf("failed govc validations: %v", err)
	}

	templateInLibraryPath := filepath.Join(library, templateName)
	if !filepath.IsAbs(templateInLibraryPath) {
		templateInLibraryPath = fmt.Sprintf("/%s", templateInLibraryPath)
	}

	deployOptsPath, err := g.writer.Write(deployOptsFile, deployOpts, filewriter.PersistentFile)
	if err != nil {
		return fmt.Errorf("failed writing deploy options file to disk: %v", err)
	}

	bFolderNotFound := false
	params := []string{"folder.info", deployFolder}
	err = g.retrier.Retry(func() error {
		errBuffer, err := g.ExecuteWithEnv(ctx, envMap, params...)
		errString := strings.ToLower(errBuffer.String())
		if err != nil {
			if !strings.Contains(errString, "not found") {
				return fmt.Errorf("error obtaining folder information: %v", err)
			} else {
				bFolderNotFound = true
			}
		}
		return nil
	})
	if err != nil || bFolderNotFound {
		params = []string{"folder.create", deployFolder}
		err = g.retrier.Retry(func() error {
			errBuffer, err := g.ExecuteWithEnv(ctx, envMap, params...)
			errString := strings.ToLower(errBuffer.String())
			if err != nil && !strings.Contains(errString, "already exists") {
				return fmt.Errorf("error creating folder: %v", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("error creating folder: %v", err)
		}
	}

	params = []string{
		"library.deploy",
		"-dc", datacenter,
		"-ds", datastore,
		"-pool", resourcePool,
		"-folder", deployFolder,
		"-options", deployOptsPath,
		templateInLibraryPath, templateName,
	}
	if _, err := g.exec(ctx, params...); err != nil {
		return fmt.Errorf("error deploying template: %v", err)
	}

	return nil
}

func (g *Govc) DeleteTemplate(ctx context.Context, resourcePool, templatePath string) error {
	if err := g.markAsVM(ctx, resourcePool, templatePath); err != nil {
		return err
	}
	if err := g.removeSnapshotsFromVM(ctx, templatePath); err != nil {
		return err
	}
	if err := g.deleteVM(ctx, templatePath); err != nil {
		return err
	}

	return nil
}

func (g *Govc) markAsVM(ctx context.Context, resourcePool, path string) error {
	if _, err := g.exec(ctx, "vm.markasvm", "-pool", resourcePool, path); err != nil {
		return fmt.Errorf("failed marking as vm: %v", err)
	}
	return nil
}

func (g *Govc) removeSnapshotsFromVM(ctx context.Context, path string) error {
	if _, err := g.exec(ctx, "snapshot.remove", "-vm", path, "*"); err != nil {
		return fmt.Errorf("error removing snapshots from vm: %v", err)
	}
	return nil
}

func (g *Govc) deleteVM(ctx context.Context, path string) error {
	if _, err := g.exec(ctx, "vm.destroy", path); err != nil {
		return fmt.Errorf("error deleting vm: %v", err)
	}
	return nil
}

func (g *Govc) createVMSnapshot(ctx context.Context, datacenter, name string) error {
	if _, err := g.exec(ctx, "snapshot.create", "-dc", datacenter, "-m=false", "-vm", name, "root"); err != nil {
		return fmt.Errorf("govc failed taking vm snapshot: %v", err)
	}
	return nil
}

func (g *Govc) markVMAsTemplate(ctx context.Context, datacenter, vmName string) error {
	if _, err := g.exec(ctx, "vm.markastemplate", "-dc", datacenter, vmName); err != nil {
		return fmt.Errorf("error marking VM as template: %v", err)
	}
	return nil
}

func (g *Govc) getEnvMap() (map[string]string, error) {
	envMap := make(map[string]string)
	for key := range g.requiredEnvs.iterate() {
		if env, ok := os.LookupEnv(key); ok && len(env) > 0 {
			envMap[key] = env
		} else {
			if key != govcInsecure {
				return nil, fmt.Errorf("warning required env not set %s", key)
			}
			err := os.Setenv(govcInsecure, "false")
			if err != nil {
				logger.Info("Warning: Unable to set <%s>", govcInsecure)
			}
		}
	}

	return envMap, nil
}

func (g *Govc) validateAndSetupCreds() (map[string]string, error) {
	var vSphereUsername, vSpherePassword, vSphereURL string
	var ok bool
	var envMap map[string]string
	if vSphereUsername, ok = os.LookupEnv(vSphereUsernameKey); ok && len(vSphereUsername) > 0 {
		if err := os.Setenv(govcUsernameKey, vSphereUsername); err != nil {
			return nil, fmt.Errorf("unable to set %s: %v", govcUsernameKey, err)
		}
	} else if govcUsername, ok := os.LookupEnv(govcUsernameKey); !ok || len(govcUsername) <= 0 {
		return nil, fmt.Errorf("%s is not set or is empty: %t", govcUsernameKey, ok)
	}
	if vSpherePassword, ok = os.LookupEnv(vSpherePasswordKey); ok && len(vSpherePassword) > 0 {
		if err := os.Setenv(govcPasswordKey, vSpherePassword); err != nil {
			return nil, fmt.Errorf("unable to set %s: %v", govcPasswordKey, err)
		}
	} else if govcPassword, ok := os.LookupEnv(govcPasswordKey); !ok || len(govcPassword) <= 0 {
		return nil, fmt.Errorf("%s is not set or is empty: %t", govcPasswordKey, ok)
	}
	if vSphereURL, ok = os.LookupEnv(vSphereServerKey); ok && len(vSphereURL) > 0 {
		if err := os.Setenv(govcURLKey, vSphereURL); err != nil {
			return nil, fmt.Errorf("unable to set %s: %v", govcURLKey, err)
		}
	} else if govcURL, ok := os.LookupEnv(govcURLKey); !ok || len(govcURL) <= 0 {
		return nil, fmt.Errorf("%s is not set or is empty: %t", govcURLKey, ok)
	}
	envMap, err := g.getEnvMap()
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	return envMap, nil
}

func (g *Govc) CleanupVms(ctx context.Context, clusterName string, dryRun bool) error {
	envMap, err := g.validateAndSetupCreds()
	if err != nil {
		return fmt.Errorf("failed govc validations: %v", err)
	}

	var params []string
	var result bytes.Buffer

	params = strings.Fields("find -type VirtualMachine -name " + clusterName + "*")
	result, err = g.ExecuteWithEnv(ctx, envMap, params...)
	if err != nil {
		return fmt.Errorf("error getting vm list: %v", err)
	}
	scanner := bufio.NewScanner(strings.NewReader(result.String()))
	for scanner.Scan() {
		vmName := scanner.Text()
		if dryRun {
			logger.Info("Found ", "vm_name", vmName)
			continue
		}
		params = strings.Fields("vm.power -off -force " + vmName)
		result, _ = g.ExecuteWithEnv(ctx, envMap, params...)
		params = strings.Fields("object.destroy " + vmName)
		result, _ = g.ExecuteWithEnv(ctx, envMap, params...)
		logger.Info("Deleted ", "vm_name", vmName)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failure reading output of vm list")
	}
	return nil
}

func (g *Govc) ValidateVCenterConnection(ctx context.Context, server string) error {
	skipVerifyTransport := http.DefaultTransport.(*http.Transport).Clone()
	skipVerifyTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: skipVerifyTransport}

	if _, err := client.Get("https://" + server); err != nil {
		return fmt.Errorf("failed to reach server %s: %v", server, err)
	}

	return nil
}

func (g *Govc) ValidateVCenterAuthentication(ctx context.Context) error {
	err := g.retrier.Retry(func() error {
		_, err := g.exec(ctx, "about", "-k")
		return err
	})
	if err != nil {
		return fmt.Errorf("vSphere authentication failed: %v", err)
	}

	return nil
}

func (g *Govc) IsCertSelfSigned(ctx context.Context) bool {
	_, err := g.exec(ctx, "about")
	return err != nil
}

func (g *Govc) GetCertThumbprint(ctx context.Context) (string, error) {
	bufferResponse, err := g.exec(ctx, "about.cert", "-thumbprint", "-k")
	if err != nil {
		return "", fmt.Errorf("unable to retrieve thumbprint: %v", err)
	}

	data := strings.Split(strings.Trim(bufferResponse.String(), "\n"), " ")
	if len(data) != 2 {
		return "", fmt.Errorf("invalid thumbprint format")
	}

	return data[1], nil
}

func (g *Govc) ConfigureCertThumbprint(ctx context.Context, server, thumbprint string) error {
	path, err := g.writer.Write(filepath.Base(govcTlsHostsFile), []byte(fmt.Sprintf("%s %s", server, thumbprint)))
	if err != nil {
		return fmt.Errorf("error writing to file %s: %v", govcTlsHostsFile, err)
	}

	if err = os.Setenv(govcTlsKnownHostsKey, path); err != nil {
		return fmt.Errorf("unable to set %s: %v", govcTlsKnownHostsKey, err)
	}

	g.requiredEnvs.append(govcTlsKnownHostsKey)

	return nil
}

func (g *Govc) DatacenterExists(ctx context.Context, datacenter string) (bool, error) {
	exists := false
	err := g.retrier.Retry(func() error {
		result, err := g.exec(ctx, "datacenter.info", datacenter)
		if err == nil {
			exists = true
			return nil
		}

		if strings.HasSuffix(result.String(), "not found") {
			exists = false
			return nil
		}

		return err
	})
	if err != nil {
		return false, fmt.Errorf("failed to get datacenter: %v", err)
	}

	return exists, nil
}

func (g *Govc) NetworkExists(ctx context.Context, network string) (bool, error) {
	exists := false

	err := g.retrier.Retry(func() error {
		networkResponse, err := g.exec(ctx, "find", "-maxdepth=1", filepath.Dir(network), "-type", "n", "-name", filepath.Base(network))
		if err != nil {
			return err
		}

		if networkResponse.String() == "" {
			exists = false
			return nil
		}

		exists = true
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("failed checking '%s' network", filepath.Base(network))
	}

	return exists, nil
}

func (g *Govc) ValidateVCenterSetupMachineConfig(ctx context.Context, datacenterConfig *v1alpha1.VSphereDatacenterConfig, machineConfig *v1alpha1.VSphereMachineConfig, _ *bool) error {
	envMap, err := g.validateAndSetupCreds()
	if err != nil {
		return fmt.Errorf("failed govc validations: %v", err)
	}
	machineConfig.Spec.Datastore, err = prependPath(datastore, machineConfig.Spec.Datastore, datacenterConfig.Spec.Datacenter)
	if err != nil {
		return err
	}
	params := []string{"datastore.info", machineConfig.Spec.Datastore}
	err = g.retrier.Retry(func() error {
		_, err = g.ExecuteWithEnv(ctx, envMap, params...)
		if err != nil {
			datastorePath := filepath.Dir(machineConfig.Spec.Datastore)
			isValidDatastorePath := g.isValidPath(ctx, envMap, datastorePath)
			if isValidDatastorePath {
				leafDir := filepath.Base(machineConfig.Spec.Datastore)
				return fmt.Errorf("valid path, but '%s' is not a datastore", leafDir)
			} else {
				return fmt.Errorf("failed to get datastore: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get datastore: %v", err)
	}
	logger.MarkPass("Datastore validated")

	if len(machineConfig.Spec.Folder) > 0 {
		machineConfig.Spec.Folder, err = prependPath(vm, machineConfig.Spec.Folder, datacenterConfig.Spec.Datacenter)
		if err != nil {
			return err
		}
		params = []string{"folder.info", machineConfig.Spec.Folder}
		err = g.retrier.Retry(func() error {
			_, err := g.ExecuteWithEnv(ctx, envMap, params...)
			if err != nil {
				err = g.createFolder(ctx, envMap, machineConfig)
				if err != nil {
					currPath := "/" + datacenterConfig.Spec.Datacenter + "/"
					dirs := strings.Split(machineConfig.Spec.Folder, "/")
					for _, dir := range dirs[2:] {
						currPath += dir + "/"
						if !g.isValidPath(ctx, envMap, currPath) {
							return fmt.Errorf("%s is an invalid intermediate directory", currPath)
						}
					}
					return err
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to get folder: %v", err)
		}
		logger.MarkPass("Folder validated")
	}

	var poolInfoResponse bytes.Buffer
	params = []string{"find", "-json", "/" + datacenterConfig.Spec.Datacenter, "-type", "p", "-name", filepath.Base(machineConfig.Spec.ResourcePool)}
	err = g.retrier.Retry(func() error {
		poolInfoResponse, err = g.ExecuteWithEnv(ctx, envMap, params...)
		return err
	})
	if err != nil {
		return fmt.Errorf("error getting resource pool: %v", err)
	}

	poolInfoJson := poolInfoResponse.String()
	poolInfoJson = strings.TrimSuffix(poolInfoJson, "\n")
	if poolInfoJson == "null" || poolInfoJson == "" {
		return fmt.Errorf("resource pool '%s' not found", machineConfig.Spec.ResourcePool)
	}

	poolInfo := make([]string, 0)
	if err = json.Unmarshal([]byte(poolInfoJson), &poolInfo); err != nil {
		return fmt.Errorf("failed unmarshalling govc response: %v", err)
	}

	machineConfig.Spec.ResourcePool = strings.TrimPrefix(machineConfig.Spec.ResourcePool, "*/")
	bPoolFound := false
	var foundPool string
	for _, p := range poolInfo {
		if strings.HasSuffix(p, machineConfig.Spec.ResourcePool) {
			if bPoolFound {
				return fmt.Errorf("specified resource pool '%s' maps to multiple paths within the datacenter '%s'", machineConfig.Spec.ResourcePool, datacenterConfig.Spec.Datacenter)
			}
			bPoolFound = true
			foundPool = p
		}
	}
	if !bPoolFound {
		return fmt.Errorf("resource pool '%s' not found", machineConfig.Spec.ResourcePool)
	}
	machineConfig.Spec.ResourcePool = foundPool

	logger.MarkPass("Resource pool validated")
	return nil
}

func prependPath(folderType FolderType, folderPath string, datacenter string) (string, error) {
	prefix := fmt.Sprintf("/%s", datacenter)
	modPath := folderPath
	if !strings.HasPrefix(folderPath, prefix) {
		modPath = fmt.Sprintf("%s/%s/%s", prefix, folderType, folderPath)
		logger.V(4).Info(fmt.Sprintf("Relative %s path specified, using path %s", folderType, modPath))
		return modPath, nil
	}
	prefix += fmt.Sprintf("/%s", folderType)
	if !strings.HasPrefix(folderPath, prefix) {
		return folderPath, fmt.Errorf("invalid folder type, expected path under %s", prefix)
	}
	return modPath, nil
}

func (g *Govc) createFolder(ctx context.Context, envMap map[string]string, machineConfig *v1alpha1.VSphereMachineConfig) error {
	params := []string{"folder.create", machineConfig.Spec.Folder}
	err := g.retrier.Retry(func() error {
		_, err := g.ExecuteWithEnv(ctx, envMap, params...)
		if err != nil {
			return fmt.Errorf("error creating folder: %v", err)
		}
		return nil
	})
	return err
}

func (g *Govc) isValidPath(ctx context.Context, envMap map[string]string, path string) bool {
	params := []string{"folder.info", path}
	_, err := g.ExecuteWithEnv(ctx, envMap, params...)
	return err == nil
}

func (g *Govc) GetTags(ctx context.Context, path string) ([]string, error) {
	tagsResponse, err := g.exec(ctx, "tags.attached.ls", "-json", "-r", path)
	if err != nil {
		return nil, fmt.Errorf("govc returned error when listing tags for %s: %v", path, err)
	}

	tagsJson := tagsResponse.String()
	if tagsJson == "null" {
		return nil, nil
	}

	tags := make([]string, 0)
	if err = json.Unmarshal([]byte(tagsJson), &tags); err != nil {
		return nil, fmt.Errorf("failed unmarshalling govc response to get tags for %s: %v", path, err)
	}

	return tags, nil
}

type tag struct {
	Id         string
	Name       string
	CategoryId string `json:"category_id,omitempty"`
}

func (g *Govc) ListTags(ctx context.Context) ([]string, error) {
	tagsResponse, err := g.exec(ctx, "tags.ls", "-json")
	if err != nil {
		return nil, fmt.Errorf("govc returned error when listing tags: %v", err)
	}

	tagsJson := tagsResponse.String()
	if tagsJson == "null" {
		return nil, nil
	}

	tags := make([]tag, 0)
	if err = json.Unmarshal([]byte(tagsJson), &tags); err != nil {
		return nil, fmt.Errorf("failed unmarshalling govc response from list tags: %v", err)
	}

	tagNames := make([]string, 0, len(tags))
	for _, t := range tags {
		tagNames = append(tagNames, t.Name)
	}

	return tagNames, nil
}

func (g *Govc) AddTag(ctx context.Context, path, tag string) error {
	if _, err := g.exec(ctx, "tags.attach", tag, path); err != nil {
		return fmt.Errorf("govc returned error when attaching tag to %s: %v", path, err)
	}
	return nil
}

func (g *Govc) CreateTag(ctx context.Context, tag, category string) error {
	if _, err := g.exec(ctx, "tags.create", "-c", category, tag); err != nil {
		return fmt.Errorf("govc returned error when creating tag %s: %v", tag, err)
	}
	return nil
}

type category struct {
	Id              string
	Name            string
	Cardinality     string
	AssociableTypes []string `json:"associable_types,omitempty"`
}

func (g *Govc) ListCategories(ctx context.Context) ([]string, error) {
	categoriesResponse, err := g.exec(ctx, "tags.category.ls", "-json")
	if err != nil {
		return nil, fmt.Errorf("govc returned error when listing categories: %v", err)
	}

	categoriesJson := categoriesResponse.String()
	if categoriesJson == "null" {
		return nil, nil
	}

	categories := make([]category, 0)
	if err = json.Unmarshal([]byte(categoriesJson), &categories); err != nil {
		return nil, fmt.Errorf("failed unmarshalling govc response from list categories: %v", err)
	}

	categoryNames := make([]string, 0, len(categories))
	for _, c := range categories {
		categoryNames = append(categoryNames, c.Name)
	}

	return categoryNames, nil
}

type objectType string

const virtualMachine objectType = "VirtualMachine"

func (g *Govc) CreateCategoryForVM(ctx context.Context, name string) error {
	return g.createCategory(ctx, name, []objectType{virtualMachine})
}

func (g *Govc) createCategory(ctx context.Context, name string, objectTypes []objectType) error {
	params := []string{"tags.category.create"}
	for _, t := range objectTypes {
		params = append(params, "-t", string(t))
	}
	params = append(params, name)

	if _, err := g.exec(ctx, params...); err != nil {
		return fmt.Errorf("govc returned error when creating category %s: %v", name, err)
	}
	return nil
}
