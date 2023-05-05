package warren

import (
	"strconv"
)

/* Data types */

// Disk Schema for disk definitions
type Disk struct {
	Uuid             string     `json:"uuid"`
	Status           string     `json:"status"`
	UserId           int        `json:"user_id"`
	BillingAccountId int        `json:"billing_account_id"`
	SizeGb           int        `json:"size_gb"`
	SourceImageType  string     `json:"source_image_type"`
	SourceImage      string     `json:"source_image"`
	CreatedAt        string     `json:"created_at"`
	UpdatedAt        string     `json:"updated_at"`
	Snapshots        []Snapshot `json:"snapshots"`
	StatusComment    string     `json:"status_comment"`
}

// Snapshot Schema for disk snapshots
type Snapshot struct {
	Uuid      string `json:"uuid"`
	SizeGb    int    `json:"sizeGb"`
	CreatedAt string `json:"created_at"`
	DiskUuid  string `json:"disk_uuid"`
}

// SourceImageType Enum for disk source image type
type SourceImageType string

const (
	OS_BASE  SourceImageType = "OS_BASE"
	DISK     SourceImageType = "DISK"
	SNAPSHOT SourceImageType = "SNAPSHOT"
	EXTERNAL SourceImageType = "EXTERNAL"
	EMPTY    SourceImageType = "EMPTY"
)

/* Request input types */

// CreateDiskRequest Schema for creating new disk instance
type CreateDiskRequest struct {
	SizeGb           *int             `json:"size_gb,omitempty"`
	BillingAccountId *int             `json:"billing_account_id,omitempty"`
	SourceImageType  *SourceImageType `json:"source_image_type,omitempty"`
	SourceImage      *string          `json:"source_image,omitempty"`
}

/* API methods */

// BlockStorageService Repo for Warren block storage related services
type BlockStorageService struct {
	client *Client
}

// CreateDisk Create new disk with specified options or from snapshot
func (s *BlockStorageService) CreateDisk(req *CreateDiskRequest) (*Disk, error) {
	var resp Disk
	params := map[string]string{}
	if req.SizeGb != nil {
		params["size_gb"] = strconv.Itoa(*req.SizeGb)
	}
	if req.BillingAccountId != nil {
		params["billing_account_id"] = strconv.Itoa(*req.BillingAccountId)
	}
	if req.SourceImageType != nil {
		params["source_image_type"] = string(*req.SourceImageType)
	}
	if req.SourceImage != nil {
		params["source_image"] = *req.SourceImage
	}

	err := s.client.Call(ApiCall{
		method:       "POST",
		path:         "/storage/disks",
		formParams:   params,
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// GetDiskById Get disk by ID
func (s *BlockStorageService) GetDiskById(diskUuid string) (*Disk, error) {
	var resp Disk
	err := s.client.Call(ApiCall{
		method:       "GET",
		path:         "/storage/disk/" + diskUuid,
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// ListUserDisks List user disks
func (s *BlockStorageService) ListUserDisks() (*[]Disk, error) {
	var resp []Disk
	err := s.client.Call(ApiCall{
		method:       "GET",
		path:         "/storage/disks",
		responseData: &resp,
	})
  if err != nil {
    return nil, err
  }
	return &resp, err
}

// DeleteDiskById Delete disk by ID
func (s *BlockStorageService) DeleteDiskById(diskUuid string) error {
	err := s.client.Call(ApiCall{

		method: "DELETE",
		path:   "/storage/disk/" + diskUuid,
	})
	return err
}
