package executables

import (
	"bytes"
	"context"
	_ "embed"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/ini.v1"

	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
)

const (
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

// TODO: Add support for domain, account filtering
func (c *Cmk) ValidateTemplatePresent(ctx context.Context, domain, zone, account, template string) error {
	result, err := c.execWithNameAndIdFilters(ctx, template, "list", "templates", "templatefilter=all", "listall=true")
	if err != nil {
		return fmt.Errorf("error getting templates info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("template %s not found", template)
	}

	response := struct {
		CmkTemplates []CmkTemplate `json:"template"`
	}{}
	err = json.Unmarshal(result.Bytes(), &response)
	if err != nil {
		return fmt.Errorf("failed to parse response into json: %v", err)
	}
	templates := response.CmkTemplates
	if len(templates) > 1 {
		return fmt.Errorf("duplicate templates %s found", template)
	} else if len(templates) == 0 {
		return fmt.Errorf("template %s not found", template)
	}
	return nil
}

// TODO: Add support for domain, account filtering
func (c *Cmk) ValidateServiceOfferingPresent(ctx context.Context, domain, zone, account, serviceOffering string) error {
	result, err := c.execWithNameAndIdFilters(ctx, serviceOffering, "list", "serviceofferings")
	if err != nil {
		return fmt.Errorf("error getting service offerings info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("service offering %s not found", serviceOffering)
	}

	response := struct {
		CmkServiceOfferings []CmkServiceOffering `json:"serviceoffering"`
	}{}
	err = json.Unmarshal(result.Bytes(), &response)
	if err != nil {
		return fmt.Errorf("failed to parse response into json: %v", err)
	}
	offerings := response.CmkServiceOfferings
	if len(offerings) > 1 {
		return fmt.Errorf("duplicate service offering %s found", serviceOffering)
	} else if len(offerings) == 0 {
		return fmt.Errorf("service offering %s not found", serviceOffering)
	}

	return nil
}

// TODO: Add support for domain, account filtering
func (c *Cmk) ValidateDiskOfferingPresent(ctx context.Context, domain, zone, account, diskOffering string) error {
	if diskOffering == "" {
		return nil
	}
	result, err := c.execWithNameAndIdFilters(ctx, diskOffering, "list", "diskofferings")
	if err != nil {
		return fmt.Errorf("error getting disk offerings info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("disk offering %s not found", diskOffering)
	}

	response := struct {
		CmkDiskOfferings []CmkDiskOffering `json:"diskoffering"`
	}{}
	err = json.Unmarshal(result.Bytes(), &response)
	if err != nil {
		return fmt.Errorf("failed to parse response into json: %v", err)
	}
	offerings := response.CmkDiskOfferings
	if len(offerings) > 1 {
		return fmt.Errorf("duplicate disk offering %s found", diskOffering)
	} else if len(offerings) == 0 {
		return fmt.Errorf("disk offering %s not found", diskOffering)
	}
	return nil
}

// TODO: Add support for domain, account filtering
func (c *Cmk) ValidateAffinityGroupsPresent(ctx context.Context, domain, zone, account string, affinityGroupIds []string) error {
	for _, affinityGroupId := range affinityGroupIds {
		idFilterParam := fmt.Sprintf("id=\"%s\"", affinityGroupId)
		result, err := c.exec(ctx, "list", "affinitygroups", idFilterParam)
		if err != nil {
			return fmt.Errorf("error getting affinity group info: %v", err)
		}
		if result.Len() == 0 {
			return fmt.Errorf(fmt.Sprintf("affinity group %s not found", affinityGroupId))
		}

		response := struct {
			CmkAffinityGroups []CmkAffinityGroup `json:"affinitygroup"`
		}{}
		err = json.Unmarshal(result.Bytes(), &response)
		if err != nil {
			return fmt.Errorf("failed to parse response into json: %v", err)
		}
		affinityGroup := response.CmkAffinityGroups
		if len(affinityGroup) > 1 {
			return fmt.Errorf("duplicate affinity group %s found", affinityGroupId)
		} else if len(affinityGroup) == 0 {
			return fmt.Errorf("affinity group %s not found", affinityGroupId)
		}
	}
	return nil
}

func (c *Cmk) ValidateZonePresent(ctx context.Context, zone string) error {
	result, err := c.execWithNameAndIdFilters(ctx, zone, "list", "zones")
	if err != nil {
		return fmt.Errorf("error getting zones info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("zone %s not found", zone)
	}

	response := struct {
		CmkZones []CmkZone `json:"zone"`
	}{}
	err = json.Unmarshal(result.Bytes(), &response)
	if err != nil {
		return fmt.Errorf("failed to parse response into json: %v", err)
	}
	zones := response.CmkZones
	if len(zones) > 1 {
		return fmt.Errorf("duplicate zone %s found", zone)
	} else if len(zones) == 0 {
		return fmt.Errorf("zone %s not found", zone)
	}
	return nil
}


// TODO: Add support for domain filtering
func (c *Cmk) ValidateAccountPresent(ctx context.Context, account string) error {
	result, err := c.execWithNameAndIdFilters(ctx, account, "list", "accounts")
	if err != nil {
		return fmt.Errorf("error getting accounts info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("account %s not found", account)
	}

	response := struct {
		CmkAccounts []CmkAccount `json:"account"`
	}{}
	err = json.Unmarshal(result.Bytes(), &response)
	if err != nil {
		return fmt.Errorf("failed to parse response into json: %v", err)
	}
	accounts := response.CmkAccounts
	if len(accounts) > 1 {
		return fmt.Errorf("duplicate account %s found", account)
	} else if len(accounts) == 0 {
		return fmt.Errorf("account %s not found", account)
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

func (c *Cmk) execWithNameAndIdFilters(ctx context.Context, parameterValue string, genericArgs ...string) (stdout bytes.Buffer, err error) {
	argsWithNameFilterArg := append(genericArgs, fmt.Sprintf("name=\"%s\"", parameterValue))
	result, err := c.exec(ctx, argsWithNameFilterArg...)
	if err != nil {
		return result, fmt.Errorf("error getting resource info filtering by id %s: %v", parameterValue, err)
	}
	if result.Len() == 0 {
		msg := fmt.Sprintf("No resources found with name %s. Trying again filtering by id instead", parameterValue)
		argsWithIdFilterArg := append(genericArgs, fmt.Sprintf("id=\"%s\"", parameterValue))
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
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed get environment map: %v", err)
	}
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

type CmkTemplate struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Zonename string `json:"zonename"`
}

type CmkServiceOffering struct {
	CpuNumber int    `json:"cpunumber"`
	CpuSpeed  int    `json:"cpuspeed"`
	Memory    int    `json:"memory"`
	Id        string `json:"id"`
	Name      string `json:"name"`
}

type CmkDiskOffering struct {
	DiskSize int    `json:"disksize"`
	Id       string `json:"id"`
	Name     string `json:"name"`
}

type CmkAffinityGroup struct {
	Type string `json:"type"`
	Id   string `json:"id"`
	Name string `json:"name"`
}

type CmkZone struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type CmkAccount struct {
	RoleType string `json:"roletype"`
	Domain   string `json:"domain"`
	Id       string `json:"id"`
	Name     string `json:"name"`
}
