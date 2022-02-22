package executables

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/cmk.ini
var cmkConfigTemplate string

const (
	cmkConfigFileName = "cmk_tmp.ini"
)

// Cmk this type will be used once the CloudStack provider is added to the repository
type Cmk struct {
	writer     filewriter.FileWriter
	executable Executable
	config     CmkExecConfig
	zones []cmkZone
	domain cmkDomain
	account cmkAccount

}

func (c *Cmk) ValidateDatacenterConfig(ctx context.Context, datacenterConfig v1alpha1.CloudStackDatacenterConfig) error {
	err := c.ValidateDomainPresent(ctx, datacenterConfig.Spec.Domain)
	if err != nil {
		return fmt.Errorf("error while checking domain %v", err)
	}
	err = c.ValidateAccountPresent(ctx, datacenterConfig.Spec.Account, c.domain.Id)
	if err != nil {
		return fmt.Errorf("error while checking account %v", err)
	}
	err = c.ValidateZonesPresent(ctx, datacenterConfig.Spec.Zones)
	if err != nil {
		return fmt.Errorf("error while checking zones %v", err)
	}
	for _, zone := range datacenterConfig.Spec.Zones {
		err = c.ValidateNetworkPresent(ctx, c.domain.Id, zone, c.account.Name, len(datacenterConfig.Spec.Zones) > 1)
		if err != nil {
			return fmt.Errorf("error while checking network %v", err)
		}
	}
	return nil
}

func (c *Cmk) ValidateMachineConfig(ctx context.Context, machineConfig v1alpha1.CloudStackMachineConfig) error {
	domainId := c.domain.Id
	account := c.account.Name
	var err error

	if len(machineConfig.Spec.AffinityGroupIds) > 0 {
		if err = c.ValidateAffinityGroupsPresent(ctx, domainId, account, machineConfig.Spec.AffinityGroupIds); err != nil {
			return fmt.Errorf("error while checking affinity groupIds %v", err)
		}
	}

	for _, zone := range c.zones {
		zoneId := zone.Id
		if err = c.ValidateTemplatePresent(ctx, domainId, zoneId, account, machineConfig.Spec.Template); err != nil {
			return fmt.Errorf("error while checking template %v", err)
		}
		if err = c.ValidateServiceOfferingPresent(ctx, zoneId, machineConfig.Spec.ComputeOffering); err != nil {
			return fmt.Errorf("error while checking compute offering %v", err)
		}
	}
	return nil
}

func (c *Cmk) ValidateTemplatePresent(ctx context.Context, domainId string, zoneId string, account string, template v1alpha1.CloudStackResourceRef) error {
	filterArgs := []string{"list", "templates", "templatefilter=all", "listall=true"}
	if template.Type == v1alpha1.Id {
		filterArgs = append(filterArgs, fmt.Sprintf("id=\"%s\"", template.Value))
	} else {
		filterArgs = append(filterArgs, fmt.Sprintf("name=\"%s\"", template.Value))
	}

	filterArgs = append(filterArgs, fmt.Sprintf("zoneid=\"%s\"", zoneId))
	if len(domainId) > 0 || len(account) > 0 {
		filterArgs = append(filterArgs, fmt.Sprintf("domainid=\"%s\"", domainId))
		filterArgs = append(filterArgs, fmt.Sprintf("account=\"%s\"", account))
	}
	result, err := c.exec(ctx, filterArgs...)
	if err != nil {
		return fmt.Errorf("error getting templates info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("template %s not found", template)
	}

	response := struct {
		CmkTemplates []cmkTemplate `json:"template"`
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

func (c *Cmk) ValidateServiceOfferingPresent(ctx context.Context, zoneId string, serviceOffering v1alpha1.CloudStackResourceRef) error {
	filterArgs := []string{"list", "serviceofferings"}
	if serviceOffering.Type == v1alpha1.Id {
		filterArgs = append(filterArgs, fmt.Sprintf("id=\"%s\"", serviceOffering.Value))
		filterArgs = append(filterArgs, fmt.Sprintf("zoneid=\"%s\"", zoneId))
	} else {
		filterArgs = append(filterArgs, fmt.Sprintf("name=\"%s\"", serviceOffering.Value))
		filterArgs = append(filterArgs, fmt.Sprintf("zoneid=\"%s\"", zoneId))
	}
	result, err := c.exec(ctx, filterArgs...)
	if err != nil {
		return fmt.Errorf("error getting service offerings info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("service offering %s not found", serviceOffering)
	}

	response := struct {
		CmkServiceOfferings []cmkServiceOffering `json:"serviceoffering"`
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

func (c *Cmk) ValidateAffinityGroupsPresent(ctx context.Context, domainId string, account string, affinityGroupIds []string) error {
	var filterArgs []string
	for _, affinityGroupId := range affinityGroupIds {
		filterArgs = []string{"list", "affinitygroups"}
		filterArgs = append(filterArgs, fmt.Sprintf("id=\"%s\"", affinityGroupId))
		// account must be specified within a domainId
		// domainId can be specified without account
		if len(domainId) > 0 {
			filterArgs = append(filterArgs,  fmt.Sprintf("domainid=\"%s\"", domainId))
			if len(account) > 0 {
				filterArgs = append(filterArgs,  fmt.Sprintf("account=\"%s\"", account))
			}
		}
		result, err := c.exec(ctx, filterArgs...)
		if err != nil {
			return fmt.Errorf("error getting affinity group info: %v", err)
		}
		if result.Len() == 0 {
			return fmt.Errorf(fmt.Sprintf("affinity group %s not found", affinityGroupId))
		}

		response := struct {
			CmkAffinityGroups []cmkAffinityGroup `json:"affinitygroup"`
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

func (c *Cmk) ValidateZonesPresent(ctx context.Context, zones []v1alpha1.Zone) error {
	var filterArg string

	for _, z := range zones {
		zone := z.ZoneRef
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
			return fmt.Errorf("zone %s not found", filterArg)
		}

		response := struct {
			CmkZones []cmkZone `json:"zone"`
		}{}
		if err = json.Unmarshal(result.Bytes(), &response); err != nil {
			return fmt.Errorf("failed to parse response into json: %v", err)
		}
		cmkZones := response.CmkZones
		if len(cmkZones) > 1 {
			return fmt.Errorf("duplicate zone %s found", cmkZones)
		} else if len(zones) == 0 {
			return fmt.Errorf("zone %s not found", filterArg)
		} else {
			c.zones = append(c.zones, cmkZones[0])
		}
	}

	return nil
}

func (c *Cmk) ValidateNetworkPresent(ctx context.Context, domainId string, zone v1alpha1.Zone, account string, multipleZone bool) error {
	filterArgs := []string{"list", "networks"}
	shared := ""
	if multipleZone {
		shared = "Shared"
	}
	if zone.Network.Type == v1alpha1.Id {
		filterArgs = append(filterArgs, fmt.Sprintf("id=\"%s\"", zone.Network.Value))
	}
	if len(shared) > 0 {
		filterArgs = append(filterArgs, fmt.Sprintf("type=\"%s\"", shared))
	}
	// account must be specified within a domainId
	// domainId can be specified without account
	if len(domainId) > 0 {
		filterArgs = append(filterArgs, fmt.Sprintf("domainid=\"%s\"", domainId))
		if len(account) > 0 {
			filterArgs = append(filterArgs, fmt.Sprintf("account=\"%s\"", account))
		}
	}
	zoneId := ""
	if zone.ZoneRef.Type == v1alpha1.Id {
		zoneId = zone.ZoneRef.Value
	} else {
		for _, zoneRef := range c.zones {
			if zoneRef.Name == zone.ZoneRef.Value {
				zoneId = zoneRef.Id
				break
			}
		}
	}
	filterArgs = append(filterArgs, fmt.Sprintf("zoneid=\"%s\"", zoneId))
	result, err := c.exec(ctx, filterArgs...)
	if err != nil {
		return fmt.Errorf("error getting network info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("%s network %s not found in zone %s", shared, zone.Network.Value, zone.ZoneRef.Value)
	}

	response := struct {
		CmkNetworks []cmkNetwork `json:"network"`
	}{}
	if err = json.Unmarshal(result.Bytes(), &response); err != nil {
		return fmt.Errorf("failed to parse response into json: %v", err)
	}
	networks := response.CmkNetworks

	// filter by network name -- cmk does not support name= filter
	if zone.Network.Type == v1alpha1.Name {
		networks = []cmkNetwork{}
		for _, net := range response.CmkNetworks {
			if net.Name == zone.Network.Value {
				networks = append(networks, net)
			}
		}
	}

	if len(networks) > 1 {
		return fmt.Errorf("duplicate network %s found", zone.Network.Value)
	} else if len(networks) == 0 {
		return fmt.Errorf("%s network %s not found in zone %s", shared, zone.Network.Value, zone.ZoneRef.Value)
	}
	return nil
}

func (c *Cmk) ValidateDomainPresent(ctx context.Context, domain string) error {
	result, err := c.exec(ctx, "list", "domains", fmt.Sprintf("name=\"%s\"", domain))
	if err != nil {
		return fmt.Errorf("error getting domain info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("domain %s not found", domain)
	}

	response := struct {
		CmkDomains []cmkDomain `json:"domain"`
	}{}
	if err = json.Unmarshal(result.Bytes(), &response); err != nil {
		return fmt.Errorf("failed to parse response into json: %v", err)
	}
	domains := response.CmkDomains
	if len(domains) > 1 {
		return fmt.Errorf("duplicate domain %s found", domain)
	} else if len(domains) == 0 {
		return fmt.Errorf("domain %s not found", domain)
	}
	c.domain = cmkDomain{Id: domains[0].Id, Name: domains[0].Name}
	return nil
}

func (c *Cmk) ValidateAccountPresent(ctx context.Context, account string, domainId string) error {
	result, err := c.exec(ctx, "list", "accounts", fmt.Sprintf("name=\"%s\"", account), fmt.Sprintf("domainid=\"%s\"", domainId))
	if err != nil {
		return fmt.Errorf("error getting accounts info: %v", err)
	}
	if result.Len() == 0 {
		return fmt.Errorf("account %s not found", account)
	}

	response := struct {
		CmkAccounts []cmkAccount `json:"account"`
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
	c.account = accounts[0]
	return nil
}

func NewCmk(executable Executable, writer filewriter.FileWriter, config CmkExecConfig) *Cmk {
	return &Cmk{
		writer:     writer,
		executable: executable,
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
	argsWithConfigFile := []string{"-c", configFile}
	for _, arg := range args {
		if strings.TrimSpace(arg) != "" {
			argsWithConfigFile = append(argsWithConfigFile, arg)
		}
	}

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

type cmkTemplate struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Zonename string `json:"zonename"`
}

type cmkServiceOffering struct {
	CpuNumber int    `json:"cpunumber"`
	CpuSpeed  int    `json:"cpuspeed"`
	Memory    int    `json:"memory"`
	Id        string `json:"id"`
	Name      string `json:"name"`
}

type cmkAffinityGroup struct {
	Type string `json:"type"`
	Id   string `json:"id"`
	Name string `json:"name"`
}

type cmkZone struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type cmkNetwork struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type cmkDomain struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type cmkAccount struct {
	RoleType string `json:"roletype"`
	Domain   string `json:"domain"`
	Id       string `json:"id"`
	Name     string `json:"name"`
}
