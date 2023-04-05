package warren

import (
	"strconv"
)

/* Data types */

// BaseImage Schema for base image with UI specific definitions
type BaseImage struct {
	OsName       string             `json:"os_name"`
	DisplayName  string             `json:"display_name"`
	UiPosition   int                `json:"ui_position"`
	IsDefault    bool               `json:"is_default"`
	IsAppCatalog bool               `json:"is_app_catalog"`
	Icon         string             `json:"icon"`
	Versions     []BaseImageVersion `json:"versions"`
}

// BaseImageVersion Schema for OS base image for VM instance creation
type BaseImageVersion struct {
	OsVersion   string `json:"os_version"`
	DisplayName string `json:"display_name"`
	Published   bool   `json:"published"`
}

// VirtualMachine Schema for active VM instance
type VirtualMachine struct {
	Backup         bool        `json:"backup"`
	BillingAccount uint64      `json:"billing_account"`
	CreatedAt      string      `json:"created_at"`
	Description    string      `json:"description"`
	Hostname       string      `json:"hostname"`
	Mac            string      `json:"mac"`
	Memory         int         `json:"memory"`
	Name           string      `json:"name"`
	OsName         string      `json:"os_name"`
	OsVersion      string      `json:"os_version"`
	PrivateIPv4    string      `json:"private_ipv4"`
	PublicIPv6     string      `json:"public_ipv6"`
	Status         string      `json:"status"`
	Storage        []VMStorage `json:"storage"`
	UpdatedAt      string      `json:"updated_at"`
	UserId         uint64      `json:"user_id"`
	Username       string      `json:"username"`
	Uuid           string      `json:"uuid"`
	VCpu           int         `json:"vcpu"`
}

// VMStorage Schema for VM storage
type VMStorage struct {
	CreatedAt string           `json:"created_at"`
	Name      string           `json:"name"`
	Primary   bool             `json:"primary"`
	Replica   []StorageReplica `json:"replica"`
	Size      int              `json:"size"`
	UserId    int              `json:"user_id"`
	Uuid      string           `json:"uuid"`
}

// StorageReplica Schema for block storage replica related to VM
type StorageReplica struct {
	CreatedAt  string `json:"created_at"`
	MasterUuid string `json:"master_uuid"`
	Size       int    `json:"size"`
	Type       string `json:"type"`
	Uuid       string `json:"uuid"`
}

// CreateVirtualMachineRequest Schema for creating new VM instance
type CreateVirtualMachineRequest struct {
	Name      *string `json:"name,omitempty"`
	OsName    *string `json:"os_name,omitempty"`
	OsVersion *string `json:"os_version,omitempty"`
	// Disks is boot disk size in gigabytes
	Disks *int `json:"disks,omitempty"`
	// VCpu is the number of virtual CPUs
	VCpu *int `json:"vcpu,omitempty"`
	// Ram is the amount of RAM in megabytes
	Ram              *int    `json:"ram,omitempty"`
	Username         *string `json:"username,omitempty"`
	Password         *string `json:"password,omitempty"`
	BillingAccountId *int    `json:"billing_account_id,omitempty"`
	Backup           *bool   `json:"backup,omitempty"`
	PublicKey        *string `json:"public_key,omitempty"`
	NetworkUuid      *string `json:"network_uuid,omitempty"`
	SourceReplica    *string `json:"source_replica,omitempty"`
	SourceUuid       *string `json:"source_uuid,omitempty"`
	ReservePublicIp  *bool   `json:"reserve_public_ip,omitempty"`
	CloudInit        *string `json:"cloud_init,omitempty"`
}

/* API methods */

// VirtualMachineService Repo for Warren VM related services
type VirtualMachineService struct {
	client *Client
}

// GetByUuid Get user VM instance by ID
func (c *VirtualMachineService) GetByUuid(uuid string) (*VirtualMachine, error) {
	var resp VirtualMachine
	err := c.client.Call(ApiCall{
		method:       "GET",
		path:         "/user-resource/vm",
		queryParams:  map[string]string{"uuid": uuid},
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// ListVms List all user VM instances
func (c *VirtualMachineService) ListVms() (*[]VirtualMachine, error) {
	var resp []VirtualMachine
	err := c.client.Call(ApiCall{
		method:       "GET",
		path:         "/user-resource/vm/list",
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// CreateVirtualMachine Create new VM instance by providing schema
func (c *VirtualMachineService) CreateVirtualMachine(createVm *CreateVirtualMachineRequest) (*VirtualMachine, error) {
	var resp VirtualMachine
	err := c.client.Call(ApiCall{
		method:       "POST",
		path:         "/user-resource/vm",
		jsonBody:     createVm,
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// StartVm Start VM instance by ID
func (c *VirtualMachineService) StartVm(uuid string) (*VirtualMachine, error) {
	var resp VirtualMachine
	err := c.client.Call(ApiCall{
		method:       "POST",
		path:         "/user-resource/vm/start",
		formParams:   map[string]string{"uuid": uuid},
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// StopVm Stop VM instance by ID, vm stopping can be forced
func (c *VirtualMachineService) StopVm(uuid string, force bool) (*VirtualMachine, error) {
	var resp VirtualMachine
	err := c.client.Call(ApiCall{
		method:       "POST",
		path:         "/user-resource/vm/stop",
		formParams:   map[string]string{"uuid": uuid, "force": strconv.FormatBool(force)},
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// DeleteVm Delete VM instance by ID
func (c *VirtualMachineService) DeleteVm(uuid string) error {
	err := c.client.Call(ApiCall{
		method:     "DELETE",
		path:       "/user-resource/vm",
		formParams: map[string]string{"uuid": uuid},
	})
	return err
}

// AttachDisk Attach specific disk to specified VM instance
func (c *VirtualMachineService) AttachDisk(vmUuid string, diskUuid string) (*VMStorage, error) {
	var resp VMStorage
	err := c.client.Call(ApiCall{
		method:       "POST",
		path:         "/user-resource/vm/storage/attach",
		formParams:   map[string]string{"uuid": vmUuid, "storage_uuid": diskUuid},
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// DetachDisk Detach specific disk from specified VM instance
func (c *VirtualMachineService) DetachDisk(vmUuid string, diskUuid string) error {
	err := c.client.Call(ApiCall{
		method:     "POST",
		path:       "/user-resource/vm/storage/detach",
		formParams: map[string]string{"uuid": vmUuid, "storage_uuid": diskUuid},
	})
	return err
}

// ListBaseImages List all available VM base images
func (c *VirtualMachineService) ListBaseImages() (*[]BaseImage, error) {
	var resp []BaseImage
	err := c.client.Call(ApiCall{
		method:       "GET",
		path:         "/config/vm_images",
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}
