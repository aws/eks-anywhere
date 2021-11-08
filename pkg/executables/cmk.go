package executables

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	"os"
	"time"

	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

const (
	cmkPath             	= "cmk"
	cmkConfigFileName   	= "cmk_tmp.ini"
	cloudStackUsernameKey   = "CLOUDSTACK_USERNAME"
	cloudStackPasswordKey 	= "CLOUDSTACK_PASSWORD"
	cloudStackURLKey      	= "CLOUDSTACK_URL"
)

//go:embed config/cmk.ini
var cmkConfigTemplate string
var requiredEnvsCmk = []string{cloudStackUsernameKey, cloudStackPasswordKey, cloudStackURLKey}

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

// cmkExecConfig contains transient information for the execution of cmk commands
// It must be cleaned after each execution to prevent side effects from past executions options
type cmkExecConfig struct {
	env                    map[string]string
	ConfigFile             string
}

func NewCmk(executable Executable, writer filewriter.FileWriter) *Cmk {
	return &Cmk{
		writer:     writer,
		executable: executable,
		retrier:    retrier.NewWithMaxRetries(maxRetriesCmk, backOffPeriodCmk),
	}
}

// ValidateCloudStackSetup No-op for initial commit. Implementation
// will be added as part of
func (c *Cmk) ValidateCloudStackSetup(ctx context.Context) error {
	buffer, err := c.exec(ctx, "list", "zones")
	logger.Info(buffer.String())
	if err != nil {
		return fmt.Errorf("error validating cloudstack setup for cmk config: %v", err)
	}
	return nil
}


type cmkTemplateResponse struct {
	CmkTemplates []types.CmkTemplate `json:"template"`
}

func (c *Cmk) ListTemplates(ctx context.Context, template string) ([]types.CmkTemplate, error) {
	templateNameFilterArg := fmt.Sprintf("name=\"%s\"", template)
	result, err := c.exec(ctx, "list", "templates", "templatefilter=all", "listall=true", templateNameFilterArg)
	if err != nil {
		return make([]types.CmkTemplate, 0), fmt.Errorf("error getting templates info: %v", err)
	}
	if result.Len() == 0 {
		return make([]types.CmkTemplate, 0), nil
	}

	response := &cmkTemplateResponse{}
	err = json.Unmarshal(result.Bytes(), response)
	if err != nil {
		return make([]types.CmkTemplate, 0), fmt.Errorf("failed to parse response into json: %v", err)
	}
	return response.CmkTemplates, nil
}

type cmkServiceOfferingResponse struct {
	CmkServiceOfferings []types.CmkServiceOffering `json:"serviceoffering"`
}

func (c *Cmk) ListServiceOfferings(ctx context.Context, offering string) ([]types.CmkServiceOffering, error) {
	templateNameFilterArg := fmt.Sprintf("name=\"%s\"", offering)
	result, err := c.exec(ctx, "list", "serviceofferings", templateNameFilterArg)
	if err != nil {
		return make([]types.CmkServiceOffering, 0), fmt.Errorf("error getting service offerings info: %v", err)
	}
	if result.Len() == 0 {
		return make([]types.CmkServiceOffering, 0), nil
	}

	response := &cmkServiceOfferingResponse{}
	err = json.Unmarshal(result.Bytes(), response)
	if err != nil {
		return make([]types.CmkServiceOffering, 0), fmt.Errorf("failed to parse response into json: %v", err)
	}
	return response.CmkServiceOfferings, nil
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

func (c *Cmk) buildCmkConfigFile(envMap map[string]string) (err error) {
	t := templater.New(c.writer)
	data := map[string]string{
		"CloudStackUsername":      	envMap[cloudStackUsernameKey],
		"CloudStackManagementUrl":  envMap[cloudStackURLKey],
		"CloudStackPassword":      	envMap[cloudStackPasswordKey],
	}
	writtenFileName, err := t.WriteToFile(cmkConfigTemplate, data, cmkConfigFileName)
	if err != nil {
		return fmt.Errorf("error creating file for cmk config: %v", err)
	}

	c.execConfig.ConfigFile = writtenFileName

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
	if cloudStackUsername, ok = os.LookupEnv(cloudStackUsernameKey); ok && len(cloudStackUsername) > 0 {
		if err := os.Setenv(cloudStackUsernameKey, cloudStackUsername); err != nil {
			return nil, fmt.Errorf("unable to set %s: %v", cloudStackUsernameKey, err)
		}
	}
	if cloudStackPassword, ok = os.LookupEnv(cloudStackPasswordKey); ok && len(cloudStackPassword) > 0 {
		if err := os.Setenv(cloudStackPasswordKey, cloudStackPassword); err != nil {
			return nil, fmt.Errorf("unable to set %s: %v", cloudStackPasswordKey, err)
		}
	}
	if cloudStackURL, ok = os.LookupEnv(cloudStackURLKey); ok && len(cloudStackURL) > 0 {
		if err := os.Setenv(cloudStackURLKey, cloudStackURL); err != nil {
			return nil, fmt.Errorf("unable to set %s: %v", cloudStackURLKey, err)
		}
	}
	envMap, err := c.getEnvMap()
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	return envMap, nil
}
