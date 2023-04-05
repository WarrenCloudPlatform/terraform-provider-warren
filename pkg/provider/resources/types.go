/*
Copyright 2023 OYE Network OÜ. All rights reserved.

This Source Code Form is subject to the terms of the Mozilla Public License,
v. 2.0. If a copy of the MPL was not distributed with this file, You can
obtain one at http://mozilla.org/MPL/2.0/.
*/

// Package resources contains all Terraform resources supported
package resources

import (
    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.com/warrenio/library/go-client/warren"
)

type ctxWrapDataKey string

// Disk defines the resource implementation.
type Disk struct {
	client *warren.Client
}

type diskCreateMethodData struct {
	DiskUUID string
}

// DiskModel describes the resource model for a disk
type DiskModel struct {
	BillingAccount        types.Int64  `tfsdk:"billing_account"`
	CreatedAt             types.String `tfsdk:"created_at"`
	ServerUUID            types.String `tfsdk:"server_uuid"`
	SizeInGB              types.Int64  `tfsdk:"size_in_gb"`
	Snapshots             types.List   `tfsdk:"snapshots"`
	SourceImageType       types.String `tfsdk:"source_image_type"`
	SourceImageUUID       types.String `tfsdk:"source_image_uuid"`
	Status                types.String `tfsdk:"status"`
	StatusComment         types.String `tfsdk:"status_comment"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
	UserID                types.Int64  `tfsdk:"user_id"`
	UUID                  types.String `tfsdk:"id"`
}

// FloatingIP defines the resource implementation.
type FloatingIP struct {
	client *warren.Client
}

type floatingIPCreateMethodData struct {
	FloatingIPID int
}

// FloatingIPModel describes the resource model for a floating IP
type FloatingIPModel struct {
	AssignedTo             types.String `tfsdk:"assigned_to"`
	AssignedToPrivateIP    types.String `tfsdk:"assigned_to_private_ip"`
	AssignedToResourceType types.String `tfsdk:"assigned_to_resource_type"`
	Address                types.String `tfsdk:"address"`
	BillingAccount         types.Int64  `tfsdk:"billing_account"`
	CreatedAt              types.String `tfsdk:"created_at"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	ID                     types.String `tfsdk:"id"`
	IsIPv6                 types.Bool   `tfsdk:"is_ipv6"`
	Name                   types.String `tfsdk:"name"`
	NetworkUUID            types.String `tfsdk:"network_uuid"`
	Type                   types.String `tfsdk:"type"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	UserID                 types.Int64  `tfsdk:"user_id"`
	UUID                   types.String `tfsdk:"uuid"`
}

// Network defines the resource implementation.
type Network struct {
	client *warren.Client
}

// VirtualMachine defines the resource implementation.
type VirtualMachine struct {
	client *warren.Client
}

// VirtualMachineModel describes the resource model for a virtual machine.
type VirtualMachineModel struct {
	Backup                types.Bool   `tfsdk:"backup"`
	BillingAccount        types.Int64  `tfsdk:"billing_account"`
	CreatedAt             types.String `tfsdk:"created_at"`
	CloudInit             types.String `tfsdk:"cloud_init"`
	Description           types.String `tfsdk:"description"`
	DiskSizeInGB          types.Int64  `tfsdk:"disk_size_in_gb"`
	Hostname              types.String `tfsdk:"hostname"`
	MAC                   types.String `tfsdk:"mac"`
	Memory                types.Int64  `tfsdk:"memory"`
	Name                  types.String `tfsdk:"name"`
	OSName                types.String `tfsdk:"os_name"`
	OSVersion             types.String `tfsdk:"os_version"`
	Password              types.String `tfsdk:"password"`
	SourceReplica         types.String `tfsdk:"source_replica"`
	SourceUUID            types.String `tfsdk:"source_uuid"`
	Status                types.String `tfsdk:"status"`
	Storage               types.List   `tfsdk:"storage"`
	PrivateIPv4           types.String `tfsdk:"private_ipv4"`
	PublicIPv6            types.String `tfsdk:"public_ipv6"`
	PublicKey             types.String `tfsdk:"public_key"`
	NetworkUUID           types.String `tfsdk:"network_uuid"`
	ReservePublicIP       types.Bool   `tfsdk:"reserve_public_ip"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
	UserID                types.Int64  `tfsdk:"user_id"`
	Username              types.String `tfsdk:"username"`
	UUID                  types.String `tfsdk:"id"`
	VCPU                  types.Int64  `tfsdk:"vcpu"`
}

// Constant warrenPasswordGeneratedLength is the length of the generated random password
const (
	warrenPasswordGeneratedLength = 32
	warrenDefaultVMUsername = "user"
	NetworkAssignedToResourceVM = "virtual_machine"
)

var (
	DiskSnapshotType = map[string]attr.Type{
		"created_at": types.StringType,
		"disk_uuid":  types.StringType,
		"size_in_gb": types.Int64Type,
		"uuid":       types.StringType,
	}
	VirtualMachineStorageType = map[string]attr.Type{
		"created_at": types.StringType,
		"name":       types.StringType,
		"primary":    types.BoolType,
		"replica":    types.ListType{ ElemType: types.ObjectType{ AttrTypes: VirtualMachineStorageReplicaType } },
		"size_in_gb": types.Int64Type,
		"user_id":    types.Int64Type,
		"uuid":       types.StringType,
	}
	VirtualMachineStorageReplicaType = map[string]attr.Type{
		"created_at":  types.StringType,
		"master_uuid": types.StringType,
		"size_in_gb":  types.Int64Type,
		"type":        types.StringType,
		"uuid":        types.StringType,
	}
)
