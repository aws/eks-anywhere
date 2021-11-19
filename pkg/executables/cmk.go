package executables

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	cmkPath               = "cmk"
	cmkConfigFileName     = "cmk_tmp.ini"
	cloudStackUsernameKey = "EKSA_CLOUDSTACK_USERNAME"
	cloudStackPasswordKey = "EKSA_CLOUDSTACK_PASSWORD"
	cloudStackURLKey      = "CLOUDSTACK_URL"
)

var (
	//go:embed config/cmk.ini
	cmkConfigTemplate string
	requiredEnvsCmk   = []string{cloudStackUsernameKey, cloudStackPasswordKey, cloudStackURLKey}
)

const (
	maxRetriesCmk    = 5
	backOffPeriodCmk = 5 * time.Second
)

type Cmk struct {
	writer     filewriter.FileWriter
	executable Executable
	retrier    *retrier.Retrier
	execConfig *cmkExecConfig
}

func (c *Cmk) SearchTemplate(ctx context.Context, domain string, zone string, account string, template string) (string, error) {
	templates, err := c.ListTemplates(ctx, template)
	if err != nil {
		return "", fmt.Errorf("search template %s error: %v", template, err)
	} else if len(templates) > 1 {
		return "", fmt.Errorf("duplicate templates %s found", template)
	} else if len(templates) == 0 {
		return "", fmt.Errorf("template %s not found", template)
	}
	return templates[0].Name, nil
}

func (c *Cmk) SearchComputeOffering(ctx context.Context, domain string, zone string, account string, computeOffering string) (string, error) {
	offerings, err := c.ListServiceOfferings(ctx, computeOffering)
	if err != nil {
		return "", fmt.Errorf("compute offering %s not found. error: %v", computeOffering, err)
	} else if len(offerings) > 1 {
		return "", fmt.Errorf("duplicate compute offering %s found", computeOffering)
	} else if len(offerings) == 0 {
		return "", fmt.Errorf("compute offering %s not found", computeOffering)
	}

	return offerings[0].Name, nil
}

func (c *Cmk) SearchDiskOffering(ctx context.Context, domain string, zone string, account string, diskOffering string) (string, error) {
	offerings, err := c.ListDiskOfferings(ctx, diskOffering)
	if err != nil {
		return "", fmt.Errorf("disk offering %s not found. error: %v", diskOffering, err)
	} else if len(offerings) > 1 {
		return "", fmt.Errorf("duplicate disk offering %s found", diskOffering)
	} else if len(offerings) == 0 {
		return "", fmt.Errorf("disk offering %s not found", diskOffering)
	}
	return offerings[0].Name, nil
}

func (c *Cmk) ValidateCloudStackSetup(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, selfSigned *bool) error {
	return nil
	// return c.ValidateCloudStackConnection(ctx)
}

func (c *Cmk) ValidateCloudStackSetupMachineConfig(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, machineConfig *v1alpha1.CloudStackMachineConfig, selfSigned *bool) error {
	//_, err := c.ListZones(ctx, deploymentConfig.Spec.Zone)
	//if err != nil {
	//	return fmt.Errorf("zone: %s not found, error: %v", deploymentConfig.Spec.Zone, err)
	//}
	return nil
}

// cmkExecConfig contains transient information for the execution of cmk commands
// It must be cleaned after each execution to prevent side effects from past executions options
type cmkExecConfig struct {
	env        map[string]string
	ConfigFile string
}

func NewCmk(executable Executable, writer filewriter.FileWriter) *Cmk {
	return &Cmk{
		writer:     writer,
		executable: executable,
		retrier:    retrier.NewWithMaxRetries(maxRetriesCmk, backOffPeriodCmk),
	}
}

// ValidateCloudStackConnection Calls `cmk sync` to ensure that the endpoint and credentials + domain are valid
func (c *Cmk) ValidateCloudStackConnection(ctx context.Context) error {
	buffer, err := c.exec(ctx, "sync")
	if err != nil {
		return fmt.Errorf("error validating cloudstack connection for cmk config: %s", buffer.String())
	}
	logger.MarkPass("Connected to CloudStack server")
	return nil
}

// TODO: Add support for domain, account filtering
func (c *Cmk) ListTemplates(ctx context.Context, template string) ([]types.CmkTemplate, error) {
	templateNameFilterArg := fmt.Sprintf("name=\"%s\"", template)
	result, err := c.exec(ctx, "list", "templates", "templatefilter=all", "listall=true", templateNameFilterArg)
	if err != nil {
		return make([]types.CmkTemplate, 0), fmt.Errorf("error getting templates info: %v", err)
	}
	if result.Len() == 0 {
		return make([]types.CmkTemplate, 0), nil
	}

	response := struct {
		CmkTemplates []types.CmkTemplate `json:"template"`
	}{}
	err = json.Unmarshal(result.Bytes(), &response)
	if err != nil {
		return make([]types.CmkTemplate, 0), fmt.Errorf("failed to parse response into json: %v", err)
	}
	return response.CmkTemplates, nil
}

// TODO: Add support for domain, account filtering
func (c *Cmk) ListServiceOfferings(ctx context.Context, offering string) ([]types.CmkServiceOffering, error) {
	templateNameFilterArg := fmt.Sprintf("name=\"%s\"", offering)
	result, err := c.exec(ctx, "list", "serviceofferings", templateNameFilterArg)
	if err != nil {
		return make([]types.CmkServiceOffering, 0), fmt.Errorf("error getting service offerings info: %v", err)
	}
	if result.Len() == 0 {
		return make([]types.CmkServiceOffering, 0), nil
	}

	response := struct {
		CmkServiceOfferings []types.CmkServiceOffering `json:"serviceoffering"`
	}{}
	err = json.Unmarshal(result.Bytes(), &response)
	if err != nil {
		return make([]types.CmkServiceOffering, 0), fmt.Errorf("failed to parse response into json: %v", err)
	}
	return response.CmkServiceOfferings, nil
}

// TODO: Add support for domain, account filtering
func (c *Cmk) ListDiskOfferings(ctx context.Context, offering string) ([]types.CmkDiskOffering, error) {
	templateNameFilterArg := fmt.Sprintf("name=\"%s\"", offering)
	result, err := c.exec(ctx, "list", "diskofferings", templateNameFilterArg)
	if err != nil {
		return make([]types.CmkDiskOffering, 0), fmt.Errorf("error getting disk offerings info: %v", err)
	}
	if result.Len() == 0 {
		return make([]types.CmkDiskOffering, 0), nil
	}

	response := struct {
		CmkDiskOfferings []types.CmkDiskOffering `json:"diskoffering"`
	}{}
	err = json.Unmarshal(result.Bytes(), &response)
	if err != nil {
		return make([]types.CmkDiskOffering, 0), fmt.Errorf("failed to parse response into json: %v", err)
	}
	return response.CmkDiskOfferings, nil
}

// TODO: Add support for domain, account filtering
func (c *Cmk) ListZones(ctx context.Context, offering string) ([]types.CmkZone, error) {
	nameFilterArg := fmt.Sprintf("name=\"%s\"", offering)
	result, err := c.exec(ctx, "list", "zones", nameFilterArg)
	if err != nil {
		return make([]types.CmkZone, 0), fmt.Errorf("error getting zones info: %v", err)
	}
	if result.Len() == 0 {
		return make([]types.CmkZone, 0), nil
	}

	response := struct {
		CmkZones []types.CmkZone `json:"zone"`
	}{}
	err = json.Unmarshal(result.Bytes(), &response)
	if err != nil {
		return make([]types.CmkZone, 0), fmt.Errorf("failed to parse response into json: %v", err)
	}
	return response.CmkZones, nil
}

// TODO: Add support for domain filtering
func (c *Cmk) ListAccounts(ctx context.Context, accountName string) ([]types.CmkAccount, error) {
	nameFilterArg := fmt.Sprintf("name=\"%s\"", accountName)
	result, err := c.exec(ctx, "list", "accounts", nameFilterArg)
	if err != nil {
		return make([]types.CmkAccount, 0), fmt.Errorf("error getting accounts info: %v", err)
	}
	if result.Len() == 0 {
		return make([]types.CmkAccount, 0), nil
	}

	response := struct {
		CmkAccounts []types.CmkAccount `json:"account"`
	}{}
	err = json.Unmarshal(result.Bytes(), &response)
	if err != nil {
		return make([]types.CmkAccount, 0), fmt.Errorf("failed to parse response into json: %v", err)
	}
	return response.CmkAccounts, nil
}

func (c *Cmk) exec(ctx context.Context, args ...string) (stdout bytes.Buffer, err error) {
	c.setupExecConfig()
	envMap, err := c.getEnvMap()
	err = c.buildCmkConfigFile(envMap)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed cmk validations: %v", err)
	}
	argsWithConfigFile := append([]string{"-c", c.execConfig.ConfigFile}, args...)

	return c.executable.Execute(ctx, argsWithConfigFile...)
}

// TODO: Add support for passing in domain from Deployment Config Spec
func (c *Cmk) buildCmkConfigFile(envMap map[string]string) (err error) {
	t := templater.New(c.writer)
	data := map[string]string{
		"CloudStackUsername":      envMap[cloudStackUsernameKey],
		"CloudStackManagementUrl": envMap[cloudStackURLKey],
		"CloudStackPassword":      envMap[cloudStackPasswordKey],
	}
	writtenFileName, err := t.WriteToFile(cmkConfigTemplate, data, cmkConfigFileName)
	if err != nil {
		return fmt.Errorf("error creating file for cmk config: %v", err)
	}
	c.execConfig.ConfigFile, err = filepath.Abs(writtenFileName)
	if err != nil {
		return fmt.Errorf("failed to generate absolute filepath for generated config file at %s", writtenFileName)
	}

	return nil
}

func (c *Cmk) getEnvMap() (map[string]string, error) {
	envMap := make(map[string]string)
	for _, key := range requiredEnvsCmk {
		if env, ok := os.LookupEnv(key); ok && len(env) > 0 {
			envMap[key] = env
		} else {
			return envMap, fmt.Errorf("warning required env not set %s", key)
		}
	}
	return envMap, nil
}

func (c *Cmk) setupExecConfig() {
	c.execConfig = &cmkExecConfig{
		env: make(map[string]string),
	}
}

func (c *Cmk) validateAndSetupCreds() (map[string]string, error) {
	var cloudStackUsername, cloudStackPassword, cloudStackURL string
	var ok bool
	var envMap map[string]string
	if cloudStackUsername, ok = os.LookupEnv(cloudStackUsernameKey); !ok || len(cloudStackUsername) <= 0 {
		return nil, fmt.Errorf("%s is not set or is empty: %t", cloudStackUsernameKey, ok)
	}
	if cloudStackPassword, ok = os.LookupEnv(cloudStackPasswordKey); !ok || len(cloudStackPassword) <= 0 {
		return nil, fmt.Errorf("%s is not set or is empty: %t", cloudStackPasswordKey, ok)
	}
	if cloudStackURL, ok = os.LookupEnv(cloudStackURLKey); !ok || len(cloudStackURL) <= 0 {
		return nil, fmt.Errorf("%s is not set or is empty: %t", cloudStackURLKey, ok)
	}
	envMap, err := c.getEnvMap()
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	return envMap, nil
}
