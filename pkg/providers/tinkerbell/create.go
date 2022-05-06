package tinkerbell

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"
	tinkhardware "github.com/tinkerbell/tink/protos/hardware"
	tinkworkflow "github.com/tinkerbell/tink/protos/workflow"
	corev1 "k8s.io/api/core/v1"
	errorutil "k8s.io/apimachinery/pkg/util/errors"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/pbnj"
	"github.com/aws/eks-anywhere/pkg/types"
)

func (p *Provider) BootstrapClusterOpts() ([]bootstrapper.BootstrapClusterOption, error) {
	env := map[string]string{}
	// Adding proxy environment vars to the bootstrap cluster
	if p.clusterConfig.Spec.ProxyConfiguration != nil {
		noProxy := fmt.Sprintf("%s,%s", p.clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host, p.datacenterConfig.Spec.TinkerbellIP)
		for _, s := range p.clusterConfig.Spec.ProxyConfiguration.NoProxy {
			if s != "" {
				noProxy += "," + s
			}
		}
		env["HTTP_PROXY"] = p.clusterConfig.Spec.ProxyConfiguration.HttpProxy
		env["HTTPS_PROXY"] = p.clusterConfig.Spec.ProxyConfiguration.HttpsProxy
		env["NO_PROXY"] = noProxy
	}

	return []bootstrapper.BootstrapClusterOption{bootstrapper.WithEnv(env)}, nil
}

func (p *Provider) PreCAPIInstallOnBootstrap(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return nil
}

func (p *Provider) PostBootstrapSetup(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	// TODO: figure out if we need something else here
	hardwareSpec, err := p.catalogue.HardwareSpecMarshallable()
	if err != nil {
		return fmt.Errorf("failed marshalling resources for hardware spec: %v", err)
	}
	err = p.providerKubectlClient.ApplyKubeSpecFromBytesForce(ctx, cluster, hardwareSpec)
	if err != nil {
		return fmt.Errorf("applying hardware yaml: %v", err)
	}
	return nil
}

func (p *Provider) SetupAndValidateCreateCluster(ctx context.Context, clusterSpec *cluster.Spec) error {
	logger.Info("Warning: The tinkerbell infrastructure provider is still in development and should not be used in production")

	tinkHardware, err := p.providerTinkClient.GetHardware(ctx)
	if err != nil {
		return fmt.Errorf("retrieving tinkerbell hardware: %v", err)
	}
	logger.MarkPass("Connected to tinkerbell stack")

	if err := setupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	tinkerbellClusterSpec := newSpec(clusterSpec, p.machineConfigs, p.datacenterConfig)

	if err := p.configureSshKeys(); err != nil {
		return err
	}

	if err := hardware.ParseYAMLCatalogueFromFile(&p.catalogue, p.hardwareManifestPath); err != nil {
		return err
	}

	// ValidateHardwareCatalogue performs a lazy load of hardware configuration. Given subsequent steps need the hardware
	// read into memory it needs to be done first. It also needs connection to
	// Tinkerbell steps to verify hardware availability on the stack
	if err := p.validator.ValidateHardwareCatalogue(ctx, p.catalogue, tinkHardware, p.skipPowerActions, p.force); err != nil {
		return err
	}

	if p.force {
		if err := p.setHardwareStateToProvisining(ctx); err != nil {
			return err
		}

		if !p.skipPowerActions {
			if err := p.setMachinesToPXEBoot(ctx); err != nil {
				return err
			}
		}

		if err := p.scrubWorkflowsFromTinkerbell(ctx, p.catalogue.Hardware, tinkHardware); err != nil {
			return err
		}
	} else if !p.skipPowerActions {
		if err := p.validator.ValidateMachinesPoweredOff(ctx, p.catalogue); err != nil {
			return fmt.Errorf("validating machines are powered off: %w", err)
		}
	}

	if err := p.validator.ValidateTinkerbellConfig(ctx, tinkerbellClusterSpec.datacenterConfig); err != nil {
		return err
	}

	if err := p.validator.ValidateClusterMachineConfigs(ctx, tinkerbellClusterSpec); err != nil {
		return err
	}

	if err := p.validator.ValidateAndPopulateTemplate(ctx, tinkerbellClusterSpec.datacenterConfig, tinkerbellClusterSpec.Spec.TinkerbellTemplateConfigs[tinkerbellClusterSpec.controlPlaneMachineConfig().Spec.TemplateRef.Name]); err != nil {
		return fmt.Errorf("failed validating control plane template config: %v", err)
	}

	if err := p.validator.ValidateMinHardwareAvailableForCreate(tinkerbellClusterSpec.Spec.Cluster.Spec, p.catalogue); err != nil {
		return fmt.Errorf("minimum hardware not available: %v", err)
	}

	if err := p.validator.ValidateAndPopulateTemplate(ctx, tinkerbellClusterSpec.datacenterConfig, tinkerbellClusterSpec.Spec.TinkerbellTemplateConfigs[tinkerbellClusterSpec.firstWorkerMachineConfig().Spec.TemplateRef.Name]); err != nil {
		return fmt.Errorf("failed validating worker node template config: %v", err)
	}

	if !p.skipIpCheck {
		if err := p.validator.validateControlPlaneIpUniqueness(tinkerbellClusterSpec); err != nil {
			return err
		}
	} else {
		logger.Info("Skipping check for whether control plane ip is in use")
	}

	return nil
}

func (p *Provider) setHardwareStateToProvisining(ctx context.Context) error {
	for _, hardware := range p.catalogue.Hardware {
		tinkHardware, err := p.providerTinkClient.GetHardwareByUuid(ctx, hardware.Spec.ID)
		if err != nil {
			return fmt.Errorf("getting hardware with UUID '%s': %v", hardware.Spec.ID, err)
		}

		if tinkHardware.Metadata.State != Provisioning {
			tinkHardware.Metadata.State = Provisioning

			patchedHardware, err := json.Marshal(tinkHardware)
			if err != nil {
				return fmt.Errorf("marshaling hardware %s: %v", tinkHardware.Id, err)
			}

			logger.Info(fmt.Sprintf("Attempting to set state of hardware '%s' to '%s'", tinkHardware.Id, Provisioning))

			if err := p.providerTinkClient.PushHardware(ctx, patchedHardware); err != nil {
				return fmt.Errorf("patching hardware state: %v", err)
			}
		}
	}
	return nil
}

// setMachinesToPXEBoot iterates over all catalogue.BMCs and instructs them to turn off, one-time
// PXE boot, then turn on.
func (p *Provider) setMachinesToPXEBoot(ctx context.Context) error {
	secrets := make(map[string]corev1.Secret, len(p.catalogue.Secrets))
	for _, secret := range p.catalogue.Secrets {
		secrets[secret.Name] = secret
	}

	var errs []error
	for _, bmc := range p.catalogue.BMCs {
		secret, found := secrets[bmc.Spec.AuthSecretRef.Name]
		if !found {
			errs = append(errs, fmt.Errorf("could not find bmc secret for '%v'", bmc.Name))
		}

		conf := pbnj.BmcSecretConfig{
			Host:     bmc.Spec.Host,
			Username: string(secret.Data["username"]),
			Password: string(secret.Data["password"]),
			Vendor:   bmc.Spec.Vendor,
		}

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)

		if err := p.pbnj.PowerOff(ctx, conf); err != nil {
			errs = append(errs, err)
		}

		if err := p.pbnj.SetBootDevice(ctx, conf, pbnj.BootDevicePXE); err != nil {
			errs = append(errs, err)
		}

		cancel()
	}

	return errorutil.NewAggregate(errs)
}

// scrubWorkflowsFromTinkerbell removes all workflows in the Tinkerbell stack that feature in hardware by retrieving
// hardware MAC addresses using tinkerbellHardware. tinkerbellHardware is necessary because MAC addresses aren't
// available on the Hardware object type.
func (p *Provider) scrubWorkflowsFromTinkerbell(ctx context.Context, hardware []tinkv1alpha1.Hardware, tinkerbellHardware []*tinkhardware.Hardware) error {
	workflows, err := p.providerTinkClient.GetWorkflow(ctx)
	if err != nil {
		return fmt.Errorf("retrieving workflows: %w", err)
	}

	hardwareMACLookup, err := createHardwareIDToMACMapping(tinkerbellHardware)
	if err != nil {
		return err
	}

	manifestHardwareMACs, err := createMACSetFromHardwareManifests(hardwareMACLookup, hardware)
	if err != nil {
		return err
	}

	workflowIDs, err := getWorkflowsIDsFromMACs(manifestHardwareMACs, workflows)
	if err != nil {
		return err
	}

	if err := p.providerTinkClient.DeleteWorkflow(ctx, workflowIDs...); err != nil {
		return fmt.Errorf("could not delete tinkerbell workflow: %v", err)
	}

	return nil
}

func createHardwareIDToMACMapping(hardware []*tinkhardware.Hardware) (map[string]string, error) {
	hardwareMACLookup := make(map[string]string)
	for _, h := range hardware {
		if len(h.Network.Interfaces) == 0 {
			return nil, fmt.Errorf("hardware manifest without interface: hardware ID = '%v'", h.Id)
		}
		hardwareMACLookup[h.Id] = h.Network.Interfaces[0].Dhcp.Mac
	}

	return hardwareMACLookup, nil
}

func createMACSetFromHardwareManifests(hardwareMACLookup map[string]string, hardware []tinkv1alpha1.Hardware) (macAddressSet, error) {
	manifestHardwareMACs := make(macAddressSet)
	for _, h := range hardware {
		mac, found := hardwareMACLookup[h.Spec.ID]
		if !found {
			return nil, fmt.Errorf("couldn't find mac address for hardware manifest: manifest hardware ID = '%v'", h.Spec.ID)
		}

		manifestHardwareMACs.Insert(mac)
	}

	return manifestHardwareMACs, nil
}

func getWorkflowsIDsFromMACs(hardwareMACs macAddressSet, workflows []*tinkworkflow.Workflow) ([]string, error) {
	var workflowIDs []string
	for _, w := range workflows {
		mac, err := macFromWorkflow(w)
		if err != nil {
			return nil, err
		}

		if hardwareMACs.Contains(mac) {
			workflowIDs = append(workflowIDs, w.Id)
		}
	}

	return workflowIDs, nil
}

func macFromWorkflow(workflow *tinkworkflow.Workflow) (string, error) {
	var data struct {
		Mac string `json:"device_1"` // Assume the hardware device data uses device_1 as the key.
	}

	if err := json.Unmarshal([]byte(workflow.Hardware), &data); err != nil {
		return "", err
	}

	return data.Mac, nil
}

type macAddressSet map[string]struct{}

func (m *macAddressSet) Contains(mac string) bool {
	_, found := (*m)[strings.ToLower(mac)]
	return found
}

func (m *macAddressSet) Insert(mac string) {
	(*m)[strings.ToLower(mac)] = struct{}{}
}
