package hardware

type Hardware struct {
	ID       string   `json:"id"`
	Metadata Metadata `json:"metadata"`
	Network  Network  `json:"network"`
}

type Metadata struct {
	Facility Facility `json:"facility"`
	Instance Instance `json:"instance"`
	State    string   `json:"state"`
}

type Facility struct {
	FacilityCode    string `json:"facility_code"`
	PlanSlug        string `json:"plan_slug"`
	PlanVersionSlug string `json:"plan_version_slug"`
}

type Instance struct {
	ID       string  `json:"id"`
	Hostname string  `json:"hostname"`
	Storage  Storage `json:"storage"`
}

type Storage struct {
	Disks []Disk `json:"disks"`
}

type Disk struct {
	Device string `json:"device"`
}

type Network struct {
	Interfaces []Interface `json:"interfaces"`
}

type Interface struct {
	DHCP    DHCP    `json:"dhcp"`
	Netboot Netboot `json:"netboot"`
}

type DHCP struct {
	Arch        string   `json:"arch"`
	Mac         string   `json:"mac"`
	Nameservers []string `json:"nameservers"`
	UEFI        bool     `json:"uefi"`
	IP          IP       `json:"ip"`
}

type IP struct {
	Address string `json:"address"`
	Gateway string `json:"gateway"`
	Netmask string `json:"netmask"`
}

type Netboot struct {
	AllowPXE      bool `json:"allow_pxe"`
	AllowWorkflow bool `json:"allow_workflow"`
}
