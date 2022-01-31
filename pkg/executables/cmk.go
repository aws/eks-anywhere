package executables

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/cmk.ini
var cmkConfigTemplate string

const (
	cmkConfigFileName = "cmk_tmp.ini"
	maxRetriesCmk     = 5
	backOffPeriodCmk  = 5 * time.Second
)

type Cmk struct {
	writer     filewriter.FileWriter
	executable Executable
	retrier    *retrier.Retrier
	config     CmkExecConfig
}

// TODO: Add support for domain, account filtering
func (c *Cmk) ValidateTemplatePresent(ctx context.Context, domain, zone, account string, template v1alpha1.CloudStackResourceRef) error {
	var filterArg string
	if template.Type == v1alpha1.Id {
		filterArg = fmt.Sprintf("id=\"%s\"", template.Value)
	} else {
		filterArg = fmt.Sprintf("name=\"%s\"", template.Value)
	}
	result, err := c.exec(ctx, "list", "templates", "templatefilter=all", "listall=true", filterArg)
	if err != nil {
		return fmt.Errorf("error getting templates info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("template %s not found", template)
	}

	response := struct {
		CmkTemplates []CmkTemplate `json:"template"`
	}{}
	if err = json.Unmarshal(result.Bytes(), &response); err != nil {
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
func (c *Cmk) ValidateServiceOfferingPresent(ctx context.Context, domain, zone, account string, serviceOffering v1alpha1.CloudStackResourceRef) error {
	var filterArg string
	if serviceOffering.Type == v1alpha1.Id {
		filterArg = fmt.Sprintf("id=\"%s\"", serviceOffering.Value)
	} else {
		filterArg = fmt.Sprintf("name=\"%s\"", serviceOffering.Value)
	}
	result, err := c.exec(ctx, "list", "serviceofferings", filterArg)
	if err != nil {
		return fmt.Errorf("error getting service offerings info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("service offering %s not found", serviceOffering)
	}

	response := struct {
		CmkServiceOfferings []CmkServiceOffering `json:"serviceoffering"`
	}{}
	if err = json.Unmarshal(result.Bytes(), &response); err != nil {
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
		if err = json.Unmarshal(result.Bytes(), &response); err != nil {
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

func (c *Cmk) ValidateZonePresent(ctx context.Context, zone v1alpha1.CloudStackResourceRef) error {
	var filterArg string
	if zone.Type == v1alpha1.Id {
		filterArg = fmt.Sprintf("id=\"%s\"", zone.Value)
	} else {
		filterArg = fmt.Sprintf("name=\"%s\"", zone.Value)
	}
	result, err := c.exec(ctx, "list", "zones", filterArg)
	if err != nil {
		return fmt.Errorf("error getting zones info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("zone %s not found", zone)
	}

	response := struct {
		CmkZones []CmkZone `json:"zone"`
	}{}
	if err = json.Unmarshal(result.Bytes(), &response); err != nil {
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
	result, err := c.exec(ctx, "list", "accounts", fmt.Sprintf("name=\"%s\"", account))
	if err != nil {
		return fmt.Errorf("error getting accounts info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("account %s not found", account)
	}

	response := struct {
		CmkAccounts []CmkAccount `json:"account"`
	}{}
	if err = json.Unmarshal(result.Bytes(), &response); err != nil {
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

func NewCmk(executable Executable, writer filewriter.FileWriter, config CmkExecConfig) *Cmk {
	return &Cmk{
		writer:     writer,
		executable: executable,
		retrier:    retrier.NewWithMaxRetries(maxRetriesCmk, backOffPeriodCmk),
		config:     config,
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

func (c *Cmk) exec(ctx context.Context, args ...string) (stdout bytes.Buffer, err error) {
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed get environment map: %v", err)
	}
	configFile, err := c.buildCmkConfigFile()
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed cmk validations: %v", err)
	}
	argsWithConfigFile := append([]string{"-c", configFile}, args...)

	return c.executable.Execute(ctx, argsWithConfigFile...)
}

// TODO: Add support for passing in domain from Deployment Config Spec
type CmkExecConfig struct {
	CloudStackApiKey        string // Api Key for CloudMonkey to access CloudStack Cluster
	CloudStackSecretKey     string // Secret Key for CloudMonkey to access CloudStack Cluster
	CloudStackManagementUrl string // Management Endpoint Url for CloudMonkey to access CloudStack Cluster
	CloudMonkeyVerifyCert   bool   // boolean indicating if CloudMonkey should verify the cert presented by the CloudStack Management Server
}

func (c *Cmk) buildCmkConfigFile() (configFile string, err error) {
	t := templater.New(c.writer)
	writtenFileName, err := t.WriteToFile(cmkConfigTemplate, c.config, cmkConfigFileName)
	if err != nil {
		return "", fmt.Errorf("error creating file for cmk config: %v", err)
	}
	configFile, err = filepath.Abs(writtenFileName)
	if err != nil {
		return "", fmt.Errorf("failed to generate absolute filepath for generated config file at %s", writtenFileName)
	}

	return configFile, nil
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
