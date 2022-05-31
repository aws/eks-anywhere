package hardware

/*
This code is commented as it may prove useful for identifying validations required on pre-generated
Hardware and BMC input in something like a validating webhook.

For now, we're not focusing on that functionality hence none of this is required as it's largely
focused on the old Tinkerbell stack where state was managed in the stack and Kubernetes.
*/

// func (c *Catalogue) ValidateHardware(skipPowerActions, force bool, tinkHardwareMap map[string]*tinkhardware.Hardware, tinkWorkflowMap map[string]*tinkworkflow.Workflow) error {
// 	for _, hw := range c.AllHardware() {
// 		if hw.Name == "" {
// 			return fmt.Errorf("hardware name is required")
// 		}

// 		if errs := apimachineryvalidation.IsDNS1123Subdomain(hw.Name); len(errs) > 0 {
// 			return fmt.Errorf("invalid hardware name: %v: %v", hw.Name, errs)
// 		}

// 		if hw.Spec.Metadata.Instance.ID == "" {
// 			return fmt.Errorf("hardware: %s ID is required", hw.Name)
// 		}

// 		allHardwareWithID, err := c.LookupHardware(HardwareIDIndex, hw.Spec.Metadata.Instance.ID)
// 		if err != nil {
// 			return err
// 		}

// 		if len(allHardwareWithID) > 1 {
// 			return fmt.Errorf("duplicate hardware id: %v", hw.Spec.Metadata.Instance.ID)
// 		}

// 		if _, ok := tinkHardwareMap[hw.Spec.Metadata.Instance.ID]; !ok {
// 			return fmt.Errorf("hardware id '%s' is not registered with tinkerbell stack", hw.Spec.Metadata.Instance.ID)
// 		}

// 		hardwareInterface := tinkHardwareMap[hw.Spec.Metadata.Instance.ID].GetNetwork().GetInterfaces()
// 		for _, interfaces := range hardwareInterface {
// 			mac := interfaces.GetDhcp()
// 			if _, ok := tinkWorkflowMap[mac.Mac]; ok {
// 				message := fmt.Sprintf("workflow %s already exixts for the hardware id %s", tinkWorkflowMap[mac.Mac].Id, hw.Spec.Metadata.Instance.ID)

// 				// If the --force-cleanup flag was set we have to warn. This is beacuse we haven't separated static
// 				// and interactive validations so there's no opportunity, after performing static yaml validation, to
// 				// delete any workflows before we check if workflows alredy exist. To avoid erroring out before getting
// 				// the chance to delete workflows this code assumes workflows will be deleted at a later stage,
// 				// therefore warns only.
// 				if !force {
// 					return fmt.Errorf(message)
// 				}
// 				logger.V(2).Info(fmt.Sprintf("Warn: %v", message))
// 			}
// 		}

// 		if !force {
// 			hardwareMetadata := make(map[string]interface{})
// 			tinkHardware := tinkHardwareMap[hw.Spec.Metadata.Instance.ID]

// 			if err := json.Unmarshal([]byte(tinkHardware.GetMetadata()), &hardwareMetadata); err != nil {
// 				return fmt.Errorf("unmarshaling hardware metadata: %v", err)
// 			}

// 			if hardwareMetadata["state"] != Provisioning {
// 				return fmt.Errorf("expecting hardware state to be '%s' but it is '%s'; use --force-cleanup flag to reset the state", "provisioning", hardwareMetadata["state"])
// 			}
// 		}

// 		if !skipPowerActions {
// 			if hw.Spec.BMCRef == nil {
// 				return fmt.Errorf("bmcRef not present in hardware %s", hw.Name)
// 			}

// 			bmcs, err := c.LookupBMC(BMCNameIndex, hw.Spec.BMCRef.String())
// 			if err != nil {
// 				return err
// 			}

// 			if len(bmcs) == 0 {
// 				return fmt.Errorf("bmcRef %s not found in hardware config", hw.Spec.BMCRef)
// 			}

// 			hardwareSharingBMCRef, err := c.LookupHardware(HardwareBMCRefIndex, hw.Spec.BMCRef.String())
// 			if err != nil {
// 				return err
// 			}

// 			if len(hardwareSharingBMCRef) > 1 {
// 				return fmt.Errorf("multiple hardware referencing bmc: bmc ref=%v", hw.Spec.BMCRef)
// 			}
// 		}
// 	}

// 	return nil
// }

// func (c *Catalogue) ValidateBMC() error {
// 	bmcIpMap := make(map[string]struct{}, c.TotalBMCs())
// 	for _, bmc := range c.AllBMCs() {
// 		if bmc.Name == "" {
// 			return fmt.Errorf("bmc name is required")
// 		}

// 		if bmc.Spec.AuthSecretRef.Name == "" {
// 			return fmt.Errorf("authSecretRef name required for bmc %s", bmc.Name)
// 		}

// 		if bmc.Spec.AuthSecretRef.Namespace != constants.EksaSystemNamespace {
// 			return fmt.Errorf("invalid authSecretRef namespace: %s for bmc %s", bmc.Spec.AuthSecretRef.Namespace, bmc.Name)
// 		}

// 		secret, err := c.LookupSecret(SecretNameIndex, bmc.Spec.AuthSecretRef.Name)
// 		if err != nil {
// 			return err
// 		}

// 		if len(secret) != 1 {
// 			return fmt.Errorf("bmc authSecretRef: %s not present in hardware config", bmc.Spec.AuthSecretRef.String())
// 		}

// 		if _, ok := bmcIpMap[bmc.Spec.Host]; ok {
// 			return fmt.Errorf("duplicate host IP: %s for bmc %s", bmc.Spec.Host, bmc.Name)
// 		} else {
// 			bmcIpMap[bmc.Spec.Host] = struct{}{}
// 		}

// 		if err := networkutils.ValidateIP(bmc.Spec.Host); err != nil {
// 			return fmt.Errorf("bmc host IP: %v", err)
// 		}

// 		if bmc.Spec.Vendor == "" {
// 			return fmt.Errorf("bmc: %s vendor is required", bmc.Name)
// 		}
// 	}

// 	return nil
// }

// func (c *Catalogue) ValidateBmcSecretRefs() error {
// 	for _, s := range c.AllSecrets() {
// 		if s.Name == "" {
// 			return fmt.Errorf("secret name is required")
// 		}
// 		if s.Namespace != constants.EksaSystemNamespace {
// 			return fmt.Errorf("invalid secret namespace: %s for secret: %s expected: %s", s.Namespace, s.Name, constants.EksaSystemNamespace)
// 		}
// 		dUsr, dOk := s.Data["username"]
// 		sdUsr, sdOk := s.StringData["username"]
// 		if !dOk && !sdOk {
// 			return fmt.Errorf("secret: %s must contain key username", s.Name)
// 		}
// 		if (dOk && len(dUsr) == 0) || (sdOk && sdUsr == "") {
// 			return fmt.Errorf("username can not be empty for secret: %s", s.Name)
// 		}

// 		dPwd, dOk := s.Data["password"]
// 		sdPwd, sdOk := s.StringData["password"]
// 		if !dOk && !sdOk {
// 			return fmt.Errorf("secret: %s must contain key password", s.Name)
// 		}
// 		if (dOk && len(dPwd) == 0) || (sdOk && sdPwd == "") {
// 			return fmt.Errorf("password can not be empty for secret: %s", s.Name)
// 		}
// 	}

// 	return nil
// }
