package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HardwareState represents the hardware state.
type HardwareState string

const (
	// HardwareError represents hardware that is in an error state.
	HardwareError = HardwareState("Error")

	// HardwareReady represents hardware that is in a ready state.
	HardwareReady = HardwareState("Ready")
)

// +kubebuilder:object:root=true

// HardwareList contains a list of Hardware.
type HardwareList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hardware `json:"items"`
}

// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=hardware,scope=Namespaced,categories=tinkerbell,singular=hardware,shortName=hw
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:JSONPath=".status.state",name=State,type=string

// Hardware is the Schema for the Hardware API.
type Hardware struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HardwareSpec   `json:"spec,omitempty"`
	Status HardwareStatus `json:"status,omitempty"`
}

// HardwareSpec defines the desired state of Hardware.
type HardwareSpec struct {
	// BMCRef contains a relation to a BMC state management type in the same
	// namespace as the Hardware. This may be used for BMC management by
	// orchestrators.
	//+optional
	BMCRef *corev1.TypedLocalObjectReference `json:"bmcRef,omitempty"`

	//+optional
	Interfaces []Interface `json:"interfaces,omitempty"`

	//+optional
	// Metadata string `json:"metadata,omitempty"`

	//+optional
	Metadata *HardwareMetadata `json:"metadata,omitempty"`

	//+optional
	TinkVersion int64 `json:"tinkVersion,omitempty"`

	//+optional
	Disks []Disk `json:"disks,omitempty"`

	// Resources represents known resources that are available on a machine.
	// Resources may be used for scheduling by orchestrators.
	//+optional
	Resources map[string]resource.Quantity `json:"resources,omitempty"`

	// UserData is the user data to configure in the hardware's
	// metadata
	//+optional
	UserData *string `json:"userData,omitempty"`

	// VendorData is the vendor data to configure in the hardware's
	// metadata
	//+optional
	VendorData *string `json:"vendorData,omitempty"`
}

// Interface represents a network interface configuration for Hardware.
type Interface struct {
	//+optional
	Netboot *Netboot `json:"netboot,omitempty"`

	//+optional
	DHCP *DHCP `json:"dhcp,omitempty"`

	// DisableDHCP disables DHCP for this interface.
	// +kubebuilder:default=false
	// +optional
	DisableDHCP bool `json:"disableDhcp,omitempty"`
}

// Netboot configuration.
type Netboot struct {
	//+optional
	AllowPXE *bool `json:"allowPXE,omitempty"`

	//+optional
	AllowWorkflow *bool `json:"allowWorkflow,omitempty"`

	//+optional
	IPXE *IPXE `json:"ipxe,omitempty"`

	//+optional
	OSIE *OSIE `json:"osie,omitempty"`
}

// IPXE configuration.
type IPXE struct {
	URL      string `json:"url,omitempty"`
	Contents string `json:"contents,omitempty"`
}

// OSIE configuration.
type OSIE struct {
	BaseURL string `json:"baseURL,omitempty"`
	Kernel  string `json:"kernel,omitempty"`
	Initrd  string `json:"initrd,omitempty"`
}

// DHCP configuration.
type DHCP struct {
	// +kubebuilder:validation:Pattern="([0-9a-f]{2}[:]){5}([0-9a-f]{2})"
	MAC         string   `json:"mac,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	LeaseTime   int64    `json:"lease_time,omitempty"`
	NameServers []string `json:"name_servers,omitempty"`
	TimeServers []string `json:"time_servers,omitempty"`
	Arch        string   `json:"arch,omitempty"`
	UEFI        bool     `json:"uefi,omitempty"`
	IfaceName   string   `json:"iface_name,omitempty"`
	IP          *IP      `json:"ip,omitempty"`
	// validation pattern for VLANDID is a string number between 0-4096
	// +kubebuilder:validation:Pattern="^(([0-9][0-9]{0,2}|[1-3][0-9][0-9][0-9]|40([0-8][0-9]|9[0-6]))(,[1-9][0-9]{0,2}|[1-3][0-9][0-9][0-9]|40([0-8][0-9]|9[0-6]))*)$"
	VLANID string `json:"vlan_id,omitempty"`
}

// IP configuration.
type IP struct {
	Address string `json:"address,omitempty"`
	Netmask string `json:"netmask,omitempty"`
	Gateway string `json:"gateway,omitempty"`
	Family  int64  `json:"family,omitempty"`
}

type HardwareMetadata struct {
	State        string                `json:"state,omitempty"`
	BondingMode  int64                 `json:"bonding_mode,omitempty"`
	Manufacturer *MetadataManufacturer `json:"manufacturer,omitempty"`
	Instance     *MetadataInstance     `json:"instance,omitempty"`
	Custom       *MetadataCustom       `json:"custom,omitempty"`
	Facility     *MetadataFacility     `json:"facility,omitempty"`
}

type MetadataManufacturer struct {
	ID   string `json:"id,omitempty"`
	Slug string `json:"slug,omitempty"`
}

type MetadataInstance struct {
	ID                  string                           `json:"id,omitempty"`
	State               string                           `json:"state,omitempty"`
	Hostname            string                           `json:"hostname,omitempty"`
	AllowPxe            bool                             `json:"allow_pxe,omitempty"`
	Rescue              bool                             `json:"rescue,omitempty"`
	OperatingSystem     *MetadataInstanceOperatingSystem `json:"operating_system,omitempty"`
	AlwaysPxe           bool                             `json:"always_pxe,omitempty"`
	IpxeScriptURL       string                           `json:"ipxe_script_url,omitempty"`
	Ips                 []*MetadataInstanceIP            `json:"ips,omitempty"`
	Userdata            string                           `json:"userdata,omitempty"`
	CryptedRootPassword string                           `json:"crypted_root_password,omitempty"`
	Tags                []string                         `json:"tags,omitempty"`
	Storage             *MetadataInstanceStorage         `json:"storage,omitempty"`
	SSHKeys             []string                         `json:"ssh_keys,omitempty"`
	NetworkReady        bool                             `json:"network_ready,omitempty"`
}

type MetadataInstanceOperatingSystem struct {
	Slug     string `json:"slug,omitempty"`
	Distro   string `json:"distro,omitempty"`
	Version  string `json:"version,omitempty"`
	ImageTag string `json:"image_tag,omitempty"`
	OsSlug   string `json:"os_slug,omitempty"`
}

type MetadataInstanceIP struct {
	Address    string `json:"address,omitempty"`
	Netmask    string `json:"netmask,omitempty"`
	Gateway    string `json:"gateway,omitempty"`
	Family     int64  `json:"family,omitempty"`
	Public     bool   `json:"public,omitempty"`
	Management bool   `json:"management,omitempty"`
}

type MetadataInstanceStorage struct {
	Disks       []*MetadataInstanceStorageDisk       `json:"disks,omitempty"`
	Raid        []*MetadataInstanceStorageRAID       `json:"raid,omitempty"`
	Filesystems []*MetadataInstanceStorageFilesystem `json:"filesystems,omitempty"`
}

type MetadataInstanceStorageDisk struct {
	Device     string                                  `json:"device,omitempty"`
	WipeTable  bool                                    `json:"wipe_table,omitempty"`
	Partitions []*MetadataInstanceStorageDiskPartition `json:"partitions,omitempty"`
}

type MetadataInstanceStorageDiskPartition struct {
	Label    string `json:"label,omitempty"`
	Number   int64  `json:"number,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Start    int64  `json:"start,omitempty"`
	TypeGUID string `json:"type_guid,omitempty"`
}

type MetadataInstanceStorageRAID struct {
	Name    string   `json:"name,omitempty"`
	Level   string   `json:"level,omitempty"`
	Devices []string `json:"devices,omitempty"`
	Spare   int64    `json:"spare,omitempty"`
}

type MetadataInstanceStorageFilesystem struct {
	Mount *MetadataInstanceStorageMount `json:"mount,omitempty"`
}

type MetadataInstanceStorageMount struct {
	Device string                                         `json:"device,omitempty"`
	Format string                                         `json:"format,omitempty"`
	Files  []*MetadataInstanceStorageFile                 `json:"files,omitempty"`
	Create *MetadataInstanceStorageMountFilesystemOptions `json:"create,omitempty"`
	Point  string                                         `json:"point,omitempty"`
}

type MetadataInstanceStorageFile struct {
	Path     string `json:"path,omitempty"`
	Contents string `json:"contents,omitempty"`
	Mode     int64  `json:"mode,omitempty"`
	UID      int64  `json:"uid,omitempty"`
	GID      int64  `json:"gid,omitempty"`
}

type MetadataInstanceStorageMountFilesystemOptions struct {
	Force   bool     `json:"force,omitempty"`
	Options []string `json:"options,omitempty"`
}

type MetadataCustom struct {
	PreinstalledOperatingSystemVersion *MetadataInstanceOperatingSystem `json:"preinstalled_operating_system_version,omitempty"`
	PrivateSubnets                     []string                         `json:"private_subnets,omitempty"`
}

type MetadataFacility struct {
	PlanSlug        string `json:"plan_slug,omitempty"`
	PlanVersionSlug string `json:"plan_version_slug,omitempty"`
	FacilityCode    string `json:"facility_code,omitempty"`
}

// Disk represents a disk device for Tinkerbell Hardware.
type Disk struct {
	//+optional
	Device string `json:"device,omitempty"`
}

// HardwareStatus defines the observed state of Hardware.
type HardwareStatus struct {
	//+optional
	State HardwareState `json:"state,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Hardware{}, &HardwareList{})
}
