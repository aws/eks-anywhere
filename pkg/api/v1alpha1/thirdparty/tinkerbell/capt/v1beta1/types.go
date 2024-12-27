/*
Copyright 2022 The Tinkerbell Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

// TinkerbellResourceStatus describes the status of a Tinkerbell resource.
type TinkerbellResourceStatus int

//nolint:gomnd,gochecknoglobals
var (
	TinkerbellResourceStatusPending = TinkerbellResourceStatus(0)
	TinkerbellResourceStatusRunning = TinkerbellResourceStatus(1)
	TinkerbellResourceStatusFailed  = TinkerbellResourceStatus(2)
	TinkerbellResourceStatusTimeout = TinkerbellResourceStatus(3)
	TinkerbellResourceStatusSuccess = TinkerbellResourceStatus(4)
)

// TinkerbellMachineTemplateResource describes the data needed to create am TinkerbellMachine from a template.
type TinkerbellMachineTemplateResource struct {
	// Spec is the specification of the desired behavior of the machine.
	Spec TinkerbellMachineSpec `json:"spec"`
}
