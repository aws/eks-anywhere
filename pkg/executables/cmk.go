package executables

import (
	"bytes"
	"context"
	_ "embed"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	cmkPath                       = "cmk"
	cmkConfigFileName             = "cmk_tmp.ini"
	cloudStackb64EncodedSecretKey = "CLOUDSTACK_B64ENCODED_SECRET"
	cloudmonkeyInsecureKey        = "CLOUDMONKEY_INSECURE"
)

var (
	//go:embed config/cmk.ini
	cmkConfigTemplate string
	requiredEnvsCmk   = []string{cloudStackb64EncodedSecretKey, cloudmonkeyInsecureKey}
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
	if diskOffering == "" {
		return diskOffering, nil
	}
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

func (c *Cmk) SearchAffinityGroups(ctx context.Context, domain string, zone string, account string, affinityGroupIds []string) error {
	for _, affinityGroupId := range affinityGroupIds {
		affinityGroup, err := c.ListAffinityGroupsById(ctx, affinityGroupId)
		if err != nil {
			return fmt.Errorf("affinity group %s not found. error: %v", affinityGroupId, err)
		} else if len(affinityGroup) > 1 {
			return fmt.Errorf("duplicate affinity group %s found", affinityGroupId)
		} else if len(affinityGroup) == 0 {
			return fmt.Errorf("affinity group %s not found", affinityGroupId)
		}
	}
	return nil
}

// TODO: Add support for network checking
func (c *Cmk) ValidateCloudStackSetup(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, selfSigned *bool) error {
	errConnection := c.ValidateCloudStackConnection(ctx)
	if errConnection != nil {
		return errConnection
	}

	zones, errZone := c.ListZones(ctx, deploymentConfig.Spec.Zone)
	if errZone != nil {
		return fmt.Errorf("zone %s not found. error: %v", deploymentConfig.Spec.Zone, errZone)
	} else if len(zones) > 1 {
		return fmt.Errorf("duplicate zone %s found", deploymentConfig.Spec.Zone)
	} else if len(zones) == 0 {
		return fmt.Errorf("zone %s not found", deploymentConfig.Spec.Zone)
	}

	return nil
}

func (c *Cmk) ValidateCloudStackSetupMachineConfig(ctx context.Context, deploymentConfig *v1alpha1.CloudStackDeploymentConfig, machineConfig *v1alpha1.CloudStackMachineConfig, selfSigned *bool) error {
	domain := deploymentConfig.Spec.Domain
	zone := deploymentConfig.Spec.Zone
	account := deploymentConfig.Spec.Account

	if template, err := c.SearchTemplate(ctx, domain, zone, account, machineConfig.Spec.Template); err != nil {
		return err
	} else {
		machineConfig.Spec.Template = template
	}

	if computeOffering, err := c.SearchComputeOffering(ctx, domain, zone, account, machineConfig.Spec.ComputeOffering); err != nil {
		return err
	} else {
		machineConfig.Spec.ComputeOffering = computeOffering
	}

	if diskOffering, err := c.SearchDiskOffering(ctx, domain, zone, account, machineConfig.Spec.DiskOffering); err != nil {
		return err
	} else {
		machineConfig.Spec.DiskOffering = diskOffering
	}

	if err := c.SearchAffinityGroups(ctx, domain, zone, account, machineConfig.Spec.AffinityGroupIds); err != nil {
		return err
	}

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
		return fmt.Errorf("error validating cloudstack connection for cmk config %s: %v", buffer.String(), err)
	}
	logger.MarkPass("Connected to CloudStack server")
	return nil
}

// TODO: Add support for domain, account filtering
func (c *Cmk) ListTemplates(ctx context.Context, template string) ([]types.CmkTemplate, error) {
	result, err := c.execWithNameAndIdFilters(ctx, template, "list", "templates", "templatefilter=all", "listall=true")
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
	result, err := c.execWithNameAndIdFilters(ctx, offering, "list", "serviceofferings")
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
	result, err := c.execWithNameAndIdFilters(ctx, offering, "list", "diskofferings")
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
	result, err := c.execWithNameAndIdFilters(ctx, offering, "list", "zones")
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
func (c *Cmk) ListAccounts(ctx context.Context, account string) ([]types.CmkAccount, error) {
	result, err := c.execWithNameAndIdFilters(ctx, account, "list", "accounts")
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

func (c *Cmk) ListAffinityGroupsById(ctx context.Context, affinityGroupId string) ([]types.CmkAffinityGroup, error) {
	idFilterParam := fmt.Sprintf("id=\"%s\"", affinityGroupId)
	result, err := c.exec(ctx, "list", "affinitygroups", idFilterParam)
	if err != nil {
		return make([]types.CmkAffinityGroup, 0), fmt.Errorf("error getting affinity group info: %v", err)
	}
	if result.Len() == 0 {
		return make([]types.CmkAffinityGroup, 0), nil
	}

	response := struct {
		CmkAffinityGroups []types.CmkAffinityGroup `json:"affinitygroup"`
	}{}
	err = json.Unmarshal(result.Bytes(), &response)
	if err != nil {
		return make([]types.CmkAffinityGroup, 0), fmt.Errorf("failed to parse response into json: %v", err)
	}
	return response.CmkAffinityGroups, nil
}

func (c *Cmk) execWithNameAndIdFilters(ctx context.Context, parameterValue string, genericArgs ...string) (stdout bytes.Buffer, err error) {
	argsWithNameFilterArg := append(genericArgs, fmt.Sprintf("name=\"%s\"", parameterValue))
	argsWithIdFilterArg := append(genericArgs, fmt.Sprintf("id=\"%s\"", parameterValue))
	result, err := c.exec(ctx, argsWithNameFilterArg...)
	if err != nil {
		return result, fmt.Errorf("error getting resource info filtering by id %s: %v", parameterValue, err)
	}
	if result.Len() == 0 {
		msg := fmt.Sprintf("No resources found with name %s. Trying again filtering by id instead", parameterValue)
		logger.V(6).Info(msg)
		result, err = c.exec(ctx, argsWithIdFilterArg...)
		if err != nil {
			return result, fmt.Errorf("error getting resource info filtering by id %s: %v", parameterValue, err)
		}
	}
	return result, nil
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
	decodedString, err := b64.StdEncoding.DecodeString(envMap[cloudStackb64EncodedSecretKey])
	if err != nil {
		return fmt.Errorf("failed to decode value for %s with base64: %v", cloudStackb64EncodedSecretKey, err)
	}
	cfg, err := ini.Load(decodedString)
	if err != nil {
		return fmt.Errorf("failed to extract values from %s with ini: %v", cloudStackb64EncodedSecretKey, err)
	}
	section, err := cfg.GetSection("Global")
	if err != nil {
		return fmt.Errorf("failed to extract section 'Global' from %s: %v", cloudStackb64EncodedSecretKey, err)
	}
	apiKey, err := section.GetKey("api-key")
	if err != nil {
		return fmt.Errorf("failed to extract value of 'api-key' from %s: %v", cloudStackb64EncodedSecretKey, err)
	}
	secretKey, err := section.GetKey("secret-key")
	if err != nil {
		return fmt.Errorf("failed to extract value of 'secret-key' from %s: %v", cloudStackb64EncodedSecretKey, err)
	}
	apiUrl, err := section.GetKey("api-url")
	if err != nil {
		return fmt.Errorf("failed to extract value of 'api-url' from %s: %v", cloudStackb64EncodedSecretKey, err)
	}
	cmkInsecure, err := strconv.ParseBool(envMap[cloudmonkeyInsecureKey])
	if err != nil {
		return fmt.Errorf("failed to parse boolean value from %s: %v", cloudmonkeyInsecureKey, err)
	}
	cmkVerifyCert := strconv.FormatBool(!cmkInsecure)
	t := templater.New(c.writer)
	data := map[string]string{
		"CloudStackApiKey":        apiKey.Value(),
		"CloudStackSecretKey":     secretKey.Value(),
		"CloudStackManagementUrl": apiUrl.Value(),
		"CloudMonkeyVerifyCert":   cmkVerifyCert,
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
	var cloudStackb64EncodedSecret string
	var ok bool
	var envMap map[string]string
	if cloudStackb64EncodedSecret, ok = os.LookupEnv(cloudStackb64EncodedSecretKey); !ok || len(cloudStackb64EncodedSecret) <= 0 {
		return nil, fmt.Errorf("%s is not set or is empty: %t", cloudStackb64EncodedSecretKey, ok)
	}
	envMap, err := c.getEnvMap()
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	return envMap, nil
}
