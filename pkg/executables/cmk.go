package executables

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/cmk.ini
var cmkConfigTemplate string

const (
	cmkPath                           = "cmk"
	cmkConfigFileName                 = "cmk_tmp.ini"
	Shared                            = "Shared"
	defaultCloudStackPreflightTimeout = "30"
)

// Cmk this struct wraps around the CloudMonkey executable CLI to perform operations against a CloudStack endpoint
type Cmk struct {
	writer     filewriter.FileWriter
	executable Executable
	config     decoder.CloudStackExecConfig
}

type cmkExecConfig struct {
	CloudStackApiKey        string
	CloudStackSecretKey     string
	CloudStackManagementUrl string
	CloudMonkeyVerifyCert   string
	CloudMonkeyTimeout      string
}

func (c *Cmk) Close(ctx context.Context) error {
	return nil
}

func (c *Cmk) ValidateTemplatePresent(ctx context.Context, domainId string, zoneId string, account string, template v1alpha1.CloudStackResourceIdentifier) error {
	command := newCmkCommand("list templates")
	applyCmkArgs(&command, appendArgs("templatefilter=all"), appendArgs("listall=true"))
	if len(template.Id) > 0 {
		applyCmkArgs(&command, withCloudStackId(template.Id))
	} else {
		applyCmkArgs(&command, withCloudStackName(template.Name))
	}

	applyCmkArgs(&command, withCloudStackZoneId(zoneId))
	if len(domainId) > 0 {
		applyCmkArgs(&command, withCloudStackDomainId(domainId))
		if len(account) > 0 {
			applyCmkArgs(&command, withCloudStackAccount(account))
		}
	}
	result, err := c.exec(ctx, command...)
	if err != nil {
		return fmt.Errorf("getting templates info - %s: %v", result.String(), err)
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

func (c *Cmk) ValidateServiceOfferingPresent(ctx context.Context, zoneId string, serviceOffering v1alpha1.CloudStackResourceIdentifier) error {
	command := newCmkCommand("list serviceofferings")
	if len(serviceOffering.Id) > 0 {
		applyCmkArgs(&command, withCloudStackId(serviceOffering.Id))
	} else {
		applyCmkArgs(&command, withCloudStackName(serviceOffering.Name))
	}
	applyCmkArgs(&command, withCloudStackZoneId(zoneId))
	result, err := c.exec(ctx, command...)
	if err != nil {
		return fmt.Errorf("getting service offerings info - %s: %v", result.String(), err)
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
	for _, affinityGroupId := range affinityGroupIds {
		command := newCmkCommand("list affinitygroups")
		applyCmkArgs(&command, withCloudStackId(affinityGroupId))
		// account must be specified with a domainId
		// domainId can be specified without account
		if len(domainId) > 0 {
			applyCmkArgs(&command, withCloudStackDomainId(domainId))
			if len(account) > 0 {
				applyCmkArgs(&command, withCloudStackAccount(account))
			}
		}

		result, err := c.exec(ctx, command...)
		if err != nil {
			return fmt.Errorf("getting affinity group info - %s: %v", result.String(), err)
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

func (c *Cmk) ValidateZonesPresent(ctx context.Context, zones []v1alpha1.CloudStackZone) ([]v1alpha1.CloudStackResourceIdentifier, error) {
	var zoneIdentifiers []v1alpha1.CloudStackResourceIdentifier
	for _, zone := range zones {
		command := newCmkCommand("list zones")
		if len(zone.Id) > 0 {
			applyCmkArgs(&command, withCloudStackId(zone.Id))
		} else {
			applyCmkArgs(&command, withCloudStackName(zone.Name))
		}
		result, err := c.exec(ctx, command...)
		if err != nil {
			return nil, fmt.Errorf("getting zones info - %s: %v", result.String(), err)
		}
		if result.Len() == 0 {
			return nil, fmt.Errorf("zone %s not found", zone)
		}

		response := struct {
			CmkZones []cmkZone `json:"zone"`
		}{}
		if err = json.Unmarshal(result.Bytes(), &response); err != nil {
			return nil, fmt.Errorf("failed to parse response into json: %v", err)
		}
		cmkZones := response.CmkZones
		if len(cmkZones) > 1 {
			return nil, fmt.Errorf("duplicate zone %s found", zone)
		} else if len(zones) == 0 {
			return nil, fmt.Errorf("zone %s not found", zone)
		} else {
			zoneIdentifiers = append(zoneIdentifiers, v1alpha1.CloudStackResourceIdentifier{Name: cmkZones[0].Name, Id: cmkZones[0].Id})
		}
	}
	return zoneIdentifiers, nil
}

func (c *Cmk) ValidateDomainPresent(ctx context.Context, domain string) (v1alpha1.CloudStackResourceIdentifier, error) {
	domainIdentifier := v1alpha1.CloudStackResourceIdentifier{Name: domain, Id: ""}
	command := newCmkCommand("list domains")
	applyCmkArgs(&command, withCloudStackName(domain))
	result, err := c.exec(ctx, command...)
	if err != nil {
		return domainIdentifier, fmt.Errorf("getting domain info - %s: %v", result.String(), err)
	}
	if result.Len() == 0 {
		return domainIdentifier, fmt.Errorf("domain %s not found", domain)
	}

	response := struct {
		CmkDomains []cmkDomain `json:"domain"`
	}{}
	if err = json.Unmarshal(result.Bytes(), &response); err != nil {
		return domainIdentifier, fmt.Errorf("failed to parse response into json: %v", err)
	}
	domains := response.CmkDomains
	if len(domains) > 1 {
		return domainIdentifier, fmt.Errorf("duplicate domain %s found", domain)
	} else if len(domains) == 0 {
		return domainIdentifier, fmt.Errorf("domain %s not found", domain)
	}

	domainIdentifier.Id = domains[0].Id
	domainIdentifier.Name = domains[0].Name

	return domainIdentifier, nil
}

func (c *Cmk) ValidateNetworkPresent(ctx context.Context, domainId string, zone v1alpha1.CloudStackZone, zones []v1alpha1.CloudStackResourceIdentifier, account string, multipleZone bool) error {
	command := newCmkCommand("list networks")
	if len(zone.Network.Id) > 0 {
		applyCmkArgs(&command, withCloudStackId(zone.Network.Id))
	}
	if multipleZone {
		applyCmkArgs(&command, withCloudStackNetworkType(Shared))
	}
	// account must be specified within a domainId
	// domainId can be specified without account
	if len(domainId) > 0 {
		applyCmkArgs(&command, withCloudStackDomainId(domainId))
		if len(account) > 0 {
			applyCmkArgs(&command, withCloudStackAccount(account))
		}
	}
	var zoneId string
	var err error
	if len(zone.Id) > 0 {
		zoneId = zone.Id
	} else {
		zoneId, err = getZoneIdByName(zones, zone.Name)
		if err != nil {
			return fmt.Errorf("getting zone id by name %s: %v", zone.Name, err)
		}
	}
	applyCmkArgs(&command, withCloudStackZoneId(zoneId))
	result, err := c.exec(ctx, command...)
	if err != nil {
		return fmt.Errorf("getting network info - %s: %v", result.String(), err)
	}
	if result.Len() == 0 {
		if multipleZone {
			return fmt.Errorf("%s network %s not found in zone %s", Shared, zone.Network, zone)
		} else {
			return fmt.Errorf("network %s not found in zone %s", zone.Network, zone)
		}
	}

	response := struct {
		CmkNetworks []cmkNetwork `json:"network"`
	}{}
	if err = json.Unmarshal(result.Bytes(), &response); err != nil {
		return fmt.Errorf("failed to parse response into json: %v", err)
	}
	networks := response.CmkNetworks

	// filter by network name -- cmk does not support name= filter
	// if network id and name are both provided, the following code is to confirm name matches return value retrieved by id.
	// if only name is provided, the following code is to only get networks with specified name.

	if len(zone.Network.Name) > 0 {
		networks = []cmkNetwork{}
		for _, net := range response.CmkNetworks {
			if net.Name == zone.Network.Name {
				networks = append(networks, net)
			}
		}
	}

	if len(networks) > 1 {
		return fmt.Errorf("duplicate network %s found", zone.Network)
	} else if len(networks) == 0 {
		if multipleZone {
			return fmt.Errorf("%s network %s not found in zoneRef %s", Shared, zone.Network, zone)
		} else {
			return fmt.Errorf("network %s not found in zoneRef %s", zone.Network, zone)
		}
	}
	return nil
}

func getZoneIdByName(zones []v1alpha1.CloudStackResourceIdentifier, zoneName string) (string, error) {
	for _, zoneIdentifier := range zones {
		if zoneName == zoneIdentifier.Name {
			return zoneIdentifier.Id, nil
		}
	}
	return "", fmt.Errorf("zoneId not found for zone %s", zoneName)
}

func (c *Cmk) ValidateAccountPresent(ctx context.Context, account string, domainId string) error {
	command := newCmkCommand("list accounts")
	applyCmkArgs(&command, withCloudStackName(account), withCloudStackDomainId(domainId))
	result, err := c.exec(ctx, command...)
	if err != nil {
		return fmt.Errorf("getting accounts info - %s: %v", result.String(), err)
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
	return nil
}

func NewCmk(executable Executable, writer filewriter.FileWriter, config decoder.CloudStackExecConfig) *Cmk {
	return &Cmk{
		writer:     writer,
		executable: executable,
		config:     config,
	}
}

// ValidateCloudStackConnection Calls `cmk sync` to ensure that the endpoint and credentials + domain are valid
func (c *Cmk) ValidateCloudStackConnection(ctx context.Context) error {
	command := newCmkCommand("sync")
	buffer, err := c.exec(ctx, command...)
	if err != nil {
		return fmt.Errorf("validating cloudstack connection for cmk config %s: %v", buffer.String(), err)
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

func (c *Cmk) buildCmkConfigFile() (configFile string, err error) {
	t := templater.New(c.writer)

	cloudstackPreflightTimeout := defaultCloudStackPreflightTimeout
	if timeout, isSet := os.LookupEnv("CLOUDSTACK_PREFLIGHT_TIMEOUT"); isSet {
		if _, err := strconv.ParseUint(timeout, 10, 16); err != nil {
			return "", fmt.Errorf("CLOUDSTACK_PREFLIGHT_TIMEOUT must be a number: %v", err)
		}
		cloudstackPreflightTimeout = timeout
	}

	cmkConfig := &cmkExecConfig{
		CloudStackApiKey:        c.config.ApiKey,
		CloudStackSecretKey:     c.config.SecretKey,
		CloudStackManagementUrl: c.config.ManagementUrl,
		CloudMonkeyVerifyCert:   c.config.VerifySsl,
		CloudMonkeyTimeout:      cloudstackPreflightTimeout,
	}
	writtenFileName, err := t.WriteToFile(cmkConfigTemplate, cmkConfig, cmkConfigFileName)
	if err != nil {
		return "", fmt.Errorf("creating file for cmk config: %v", err)
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
