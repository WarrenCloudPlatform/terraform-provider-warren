package warren

import (
	"fmt"
	"net"
)

/* Data types */

// Network Schema for VPC network definition
type Network struct {
	VlanId         int      `json:"vlan_id"`
	Subnet         string   `json:"subnet"`
	SubnetIPv6     string   `json:"subnet_ipv6"`
	Name           string   `json:"name"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	Uuid           string   `json:"uuid"`
	Type           string   `json:"type"`
	IsDefault      bool     `json:"is_default"`
	VmUuids        []string `json:"vm_uuids"`
	ResourcesCount int      `json:"resources_count"`
}

// FloatingIp Schema for floating IP definitions
type FloatingIp struct {
	Id                     int    `json:"id"`
	IsIPv6                 bool   `json:"is_ipv6"`
	Address                string `json:"address"`
	UserId                 int    `json:"user_id"`
	BillingAccountId       int    `json:"billing_account_id"`
	Type                   string `json:"type"`
	Name                   string `json:"name"`
	Enabled                bool   `json:"enabled"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
	Uuid                   string `json:"uuid"`
	AssignedTo             string `json:"assigned_to"`
	AssignedToResourceType string `json:"assigned_to_resource_type"`
	AssignedToPrivateIp    string `json:"assigned_to_private_ip"`
}

/* Request input types */

// CreateFloatingIpRequest Schema for creating new floating IP instance
type CreateFloatingIpRequest struct {
	Name             *string `json:"name,omitempty"`
	BillingAccountId *int    `json:"billing_account_id,omitempty"`
}

/* API methods */

// NetworkService Repo for Warren network services
type NetworkService struct {
	client *Client
}

// GetNetworkByUUID Get user VPC network data by ID
func (c *NetworkService) GetNetworkByUUID(networkUUID string) (*Network, error) {
	var network Network
	err := c.client.Call(ApiCall{
		method:       "GET",
		path:         fmt.Sprintf("/network/network/%s/", networkUUID),
		responseData: &network,
	})
  if err != nil {
    return nil, err
  }
	return &network, err
}

// ListNetworks List user VPC networks with resources
func (c *NetworkService) ListNetworks() (*[]Network, error) {
	var networks []Network
	err := c.client.Call(ApiCall{
		method:       "GET",
		path:         "/network/networks",
		responseData: &networks,
	})
  if err != nil {
    return nil, err
  }
	return &networks, err
}

// CreateNetwork Create new network with specified name
func (c *NetworkService) CreateNetwork(networkName string) (*Network, error) {
	var network Network
	err := c.client.Call(ApiCall{
		method:       "POST",
		path:         "/network/network",
		queryParams:  map[string]string{"name": networkName},
		responseData: &network,
	})
  if err != nil {
    return nil, err
  }
	return &network, err
}

// GetDefaultNetwork Get current default VPC network
func (c *NetworkService) GetDefaultNetwork() (*Network, error) {
	var network Network
	err := c.client.Call(ApiCall{
		method:       "POST",
		path:         "/network/network",
		responseData: &network,
	})
  if err != nil {
    return nil, err
  }
	return &network, err
}

// DeleteNetworkByUUID Delete VPC network by ID. The network can be deleted if it
// does not contain any resources, and it is not default.
func (c *NetworkService) DeleteNetworkByUUID(networkUUID string) error {
	err := c.client.Call(ApiCall{
		method: "DELETE",
		path:   fmt.Sprintf("/network/network/%s/", networkUUID),
	})
	return err
}

// ChangeNetworkToDefault Change VPC network as default by ID
func (c *NetworkService) ChangeNetworkToDefault(networkUUID string) (*Network, error) {
	var network Network
	err := c.client.Call(ApiCall{
		method:       "PUT",
		path:         fmt.Sprintf("/network/network/%s/default", networkUUID),
		responseData: &network,
	})
  if err != nil {
    return nil, err
  }
	return &network, err
}

// ChangeNetworkName Change network name by ID
func (c *NetworkService) ChangeNetworkName(uuid string, newName *string) (*Network, error) {
	var network Network
	err := c.client.Call(ApiCall{
		method: "PATCH",
		path:   fmt.Sprintf("/network/network/%s", uuid),
		jsonBody: struct {
			Name *string `json:"name,omitempty"`
		}{
			newName,
		},
		responseData: &network,
	})
  if err != nil {
    return nil, err
  }
	return &network, err
}

// ListFloatingIps List user floating IPs
func (c *NetworkService) ListFloatingIps() (*[]FloatingIp, error) {
	var floatingIps []FloatingIp
	err := c.client.Call(ApiCall{
		method:       "GET",
		path:         "/network/ip_addresses",
		responseData: &floatingIps,
	})
  if err != nil {
    return nil, err
  }
	return &floatingIps, err
}

// CreateFloatingIp Create new unassigned floating IP
func (c *NetworkService) CreateFloatingIp(createIp *CreateFloatingIpRequest) (*FloatingIp, error) {
	var floatingIp FloatingIp
	err := c.client.Call(ApiCall{
		method:       "POST",
		path:         "/network/ip_addresses",
		jsonBody:     createIp,
		responseData: &floatingIp,
	})
  if err != nil {
    return nil, err
  }
	return &floatingIp, err
}

// AssignFloatingIp Assign floating IP to specific VM instance
func (c *NetworkService) AssignFloatingIp(ipAddress net.IP, vmUuid string) (*FloatingIp, error) {
	var floatingIp FloatingIp
	err := c.client.Call(ApiCall{
		method: "POST",
		path:   fmt.Sprintf("/network/ip_addresses/%s/assign", ipAddress.String()),
		jsonBody: map[string]any{
			"vm_uuid": vmUuid,
		},
		responseData: &floatingIp,
	})
  if err != nil {
    return nil, err
  }
	return &floatingIp, err
}

// UnAssignFloatingIp Un-assign floating IP from VM instance
func (c *NetworkService) UnAssignFloatingIp(ipAddress net.IP) (*FloatingIp, error) {
	var floatingIp FloatingIp
	err := c.client.Call(ApiCall{
		method:       "POST",
		path:         fmt.Sprintf("/network/ip_addresses/%s/unassign", ipAddress.String()),
		jsonBody:     map[string]any{},
		responseData: &floatingIp,
	})
  if err != nil {
    return nil, err
  }
	return &floatingIp, err
}

// DeleteFloatingIp Delete floating IP instance from user
func (c *NetworkService) DeleteFloatingIp(ipAddress net.IP) error {
	err := c.client.Call(ApiCall{
		method: "DELETE",
		path:   fmt.Sprintf("/network/ip_addresses/%s/", ipAddress.String()),
	})
	return err
}
