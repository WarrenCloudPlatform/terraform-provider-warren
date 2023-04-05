package warren

import (
	"fmt"
	"strconv"
)

/* Data types */

// LBForwardingRuleSettings Schema for load balancer forwarding rule settings
type LBForwardingRuleSettings struct {
	ConnectionLimit    int    `json:"connection_limit"`
	SessionPersistence string `json:"session_persistence"`
}

// LBForwardingRule Schema for load balancer forwarding rule definitions
type LBForwardingRule struct {
	Uuid       string                   `json:"uuid"`
	Protocol   string                   `json:"protocol"`
	CreatedAt  string                   `json:"created_at"`
	SourcePort int                      `json:"source_port"`
	TargetPort int                      `json:"target_port"`
	Settings   LBForwardingRuleSettings `json:"settings"`
}

// LBTarget Schema for load balancer target definitions
type LBTarget struct {
	CreatedAt       string `json:"created_at"`
	TargetUuid      string `json:"target_uuid"`
	TargetType      string `json:"target_type"`
	TargetIpAddress string `json:"target_ip_address"`
}

// LoadBalancer Schema for load balancer definitions
type LoadBalancer struct {
	DisplayName      *string            `json:"display_name,omitempty"`
	Uuid             string             `json:"uuid"`
	NetworkUuid      string             `json:"network_uuid"`
	UserId           int                `json:"user_id"`
	BillingAccountId int                `json:"billing_account_id"`
	CreatedAt        string             `json:"created_at"`
	UpdatedAt        string             `json:"updated_at"`
	IsDeleted        bool               `json:"is_deleted"`
	PrivateAddress   string             `json:"private_address"`
	ForwardingRules  []LBForwardingRule `json:"forwarding_rules"`
	Targets          []LBTarget         `json:"targets"`
}

/* Request input types */

// LBPortRulesRequest Schema for adding source and target port rules for load balancer
type LBPortRulesRequest struct {
	SourcePort *int `json:"source_port,omitempty"`
	TargetPort *int `json:"target_port,omitempty"`
}

// LBTargetRequest Schema for adding targets to load balancer
type LBTargetRequest struct {
	TargetUuid *string `json:"target_uuid,omitempty"`
	TargetType *string `json:"target_type,omitempty"`
}

// LBForwardingRuleRequest Schema for creating new forwarding rule for load balancer
type LBForwardingRuleRequest struct {
	SourcePort *int `json:"source_port,omitempty"`
	TargetPort *int `json:"target_port,omitempty"`
}

// LoadBalancerRequest Schema for creating new load balancer
type LoadBalancerRequest struct {
	DisplayName      *string               `json:"display_name,omitempty"`
	BillingAccountId *int                  `json:"billing_account_id,omitempty"`
	NetworkUuid      *string               `json:"network_uuid,omitempty"`
	ReservePublicIp  *bool                 `json:"reserve_public_ip,omitempty"`
	Rules            *[]LBPortRulesRequest `json:"rules,omitempty"`
	Targets          *[]LBTargetRequest    `json:"targets,omitempty"`
}

/* API methods */

// CreateLoadBalancer Create new load balancer
func (c *NetworkService) CreateLoadBalancer(loadBalancer *LoadBalancerRequest) (*LoadBalancer, error) {
	var resp LoadBalancer
	err := c.client.Call(ApiCall{
		method:       "POST",
		path:         "/network/load_balancers",
		jsonBody:     loadBalancer,
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// UpdateLoadBalancer Update load balancer name
func (c *NetworkService) UpdateLoadBalancer(lbUuid string, loadBalancer *LoadBalancerRequest) (*LoadBalancer, error) {
	var resp LoadBalancer
	err := c.client.Call(ApiCall{
		method:       "PATCH",
		path:         fmt.Sprintf("/network/load_balancers/%s", lbUuid),
		jsonBody:     loadBalancer,
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// ListLoadBalancers List user load balancers. When all=true, response includes also deleted load balancers.
func (c *NetworkService) ListLoadBalancers(showAlsoDeleted bool) (*[]LoadBalancer, error) {
	var resp []LoadBalancer
	err := c.client.Call(ApiCall{
		method:       "GET",
		path:         "/network/load_balancers",
		queryParams:  map[string]string{"all": strconv.FormatBool(showAlsoDeleted)},
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// DeleteLoadBalancer Delete user load balancer by uuid.
func (c *NetworkService) DeleteLoadBalancer(lbUuid string) error {
	err := c.client.Call(ApiCall{
		method: "DELETE",
		path:   fmt.Sprintf("/network/load_balancers/%s", lbUuid),
	})
	return err
}

// AddLoadBalancerTarget Add new target to user load balancer
func (c *NetworkService) AddLoadBalancerTarget(lbUuid string, lbTarget *LBTargetRequest) (*LBTarget, error) {
	var resp LBTarget
	err := c.client.Call(ApiCall{
		method:       "POST",
		path:         fmt.Sprintf("/network/load_balancers/%s/targets", lbUuid),
		jsonBody:     lbTarget,
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// UnlinkLoadBalancerTarget  Unlink target from load balancer
func (c *NetworkService) UnlinkLoadBalancerTarget(lbUuid string, targetUuid string) error {
	err := c.client.Call(ApiCall{
		method: "DELETE",
		path:   fmt.Sprintf("/network/load_balancers/%s/targets/%s", lbUuid, targetUuid),
	})
	return err
}

// AddLoadBalancerRule  Add new rule to user load balancer
func (c *NetworkService) AddLoadBalancerRule(lBUuid string, rule *LBForwardingRuleRequest) (*LBForwardingRule, error) {
	var resp LBForwardingRule
	err := c.client.Call(ApiCall{
		method:       "POST",
		path:         fmt.Sprintf("/network/load_balancers/%s/forwarding_rules", lBUuid),
		jsonBody:     rule,
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// DropLoadBalancerRule Delete rule by UUID from load balancer
func (c *NetworkService) DropLoadBalancerRule(lBUuid string, ruleUuid string) error {
	err := c.client.Call(ApiCall{
		method: "DELETE",
		path:   fmt.Sprintf("/network/load_balancers/%s/forwarding_rules/%s", lBUuid, ruleUuid),
	})
	return err
}

// ChangeLoadBalancerBillingAccount Change load balancer billing account
func (c *NetworkService) ChangeLoadBalancerBillingAccount(lBUuid string, newBaId int) (*LoadBalancer, error) {
	var resp LoadBalancer
	err := c.client.Call(ApiCall{
		method:       "PUT",
		path:         fmt.Sprintf("/network/load_balancers/%s/billing_account", lBUuid),
		queryParams:  map[string]string{"set_id": strconv.Itoa(newBaId)},
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}
