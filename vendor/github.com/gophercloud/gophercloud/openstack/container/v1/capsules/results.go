package capsules

import (
	"encoding/json"
	"time"

	"github.com/gophercloud/gophercloud"
)

type commonResult struct {
	gophercloud.Result
}

// Extract is a function that accepts a result and extracts a capsule resource.
func (r commonResult) Extract() (*Capsule, error) {
	var s *Capsule
	err := r.ExtractInto(&s)
	return s, err
}

// GetResult represents the result of a get operation.
type GetResult struct {
	commonResult
}

// CreateResult is the response from a Create operation. Call its Extract
// method to interpret it as a Server.
type CreateResult struct {
	gophercloud.ErrResult
}

// Represents a Container Orchestration Engine Bay, i.e. a cluster
type Capsule struct {
	// UUID for the capsule
	UUID string `json:"uuid"`

	// ID for the capsule
	ID int `json:"id"`

	// User ID for the capsule
	UserID string `json:"user_id"`

	// Project ID for the capsule
	ProjectID string `json:"project_id"`

	// cpu for the capsule
	CPU float64 `json:"cpu"`

	// Memory for the capsule
	Memory string `json:"memory"`

	// The name of the capsule
	MetaName string `json:"meta_name"`

	// Indicates whether capsule is currently operational.
	Status string `json:"status"`

	// Indicates whether capsule is currently operational.
	StatusReason string `json:"status_reason"`

	// The created time of the capsule.
	CreatedAt time.Time `json:"-"`

	// The updated time of the capsule.
	UpdatedAt time.Time `json:"-"`

	// Links includes HTTP references to the itself, useful for passing along to
	// other APIs that might want a server reference.
	Links []interface{} `json:"links"`

	// The capsule version
	CapsuleVersion string `json:"capsule_version"`

	// The capsule restart policy
	RestartPolicy string `json:"restart_policy"`

	// The capsule metadata labels
	MetaLabels map[string]string `json:"meta_labels"`

	// The list of containers uuids inside capsule.
	ContainersUUIDs []string `json:"containers_uuids"`

	// The capsule IP addresses
	Addresses map[string][]Address `json:"addresses"`

	// The capsule volume attached information
	VolumesInfo map[string][]string `json:"volumes_info"`

	// The container object inside capsule
	Containers []Container `json:"containers"`

	// The capsule host
	Host string `json:"host"`
}

type Container struct {
	// The Container IP addresses
	Addresses map[string][]Address `json:"addresses"`

	// UUID for the container
	UUID string `json:"uuid"`

	// ID for the container
	ID int `json:"id"`

	// User ID for the container
	UserID string `json:"user_id"`

	// Project ID for the container
	ProjectID string `json:"project_id"`

	// cpu for the container
	CPU float64 `json:"cpu"`

	// Memory for the container
	Memory string `json:"memory"`

	// Image for the container
	Image string `json:"image"`

	// The container container
	Labels map[string]string `json:"labels"`

	// The created time of the container
	CreatedAt time.Time `json:"-"`

	// The updated time of the container
	UpdatedAt time.Time `json:"-"`

	// Name for the container
	Name string `json:"name"`

	// Links includes HTTP references to the itself, useful for passing along to
	// other APIs that might want a server reference.
	Links []interface{} `json:"links"`

	// Container ID for the container
	ContainerID string `json:"container_id"`

	// Websocket url for the container
	WebsocketUrl string `json:"websocket_url"`

	// Websocket token for the container
	WebsocketToken string `json:"websocket_token"`

	// auto remove flag token for the container
	AutoRemove bool `json:"auto_remove"`

	// Host for the container
	Host string `json:"host"`

	// Work directory for the container
	WorkDir string `json:"workdir"`

	// Disk for the container
	Disk int `json:"disk"`

	// Image pull policy for the container
	ImagePullPolicy string `json:"image_pull_policy"`

	// Task state for the container
	TaskState string `json:"task_state"`

	// Host name for the container
	HostName string `json:"hostname"`

	// Environment for the container
	Environment map[string]string `json:"environment"`

	// Status for the container
	Status string `json:"status"`

	// Auto Heal flag for the container
	AutoHeal bool `json:"auto_heal"`

	// Status details for the container
	StatusDetail string `json:"status_detail"`

	// Status reason for the container
	StatusReason string `json:"status_reason"`

	// Image driver for the container
	ImageDriver string `json:"image_driver"`

	// Command for the container
	Command string `json:"command"`

	// Capsule ID for the container
	CapsuleID int `json:"capsule_id"`

	// Image for the container
	Runtime string `json:"runtime"`

	// Interactive flag for the container
	Interactive bool `json:"interactive"`

	// Restart Policy for the container
	RestartPolicy map[string]string `json:"restart_policy"`

	// Ports information for the container
	Ports []int `json:"ports"`

	// Meta for the container
	Meta map[string]string `json:"meta"`

	// Security groups for the container
	SecurityGroups []string `json:"security_groups"`
}

type Address struct {
	PreserveOnDelete bool    `json:"preserve_on_delete"`
	Addr             string  `json:"addr"`
	Port             string  `json:"port"`
	Version          float64 `json:"version"`
	SubnetID         string  `json:"subnet_id"`
}

func (r *Capsule) UnmarshalJSON(b []byte) error {
	type tmp Capsule
	var s struct {
		tmp
		CreatedAt gophercloud.JSONRFC3339ZNoT `json:"created_at"`
		UpdatedAt gophercloud.JSONRFC3339ZNoT `json:"updated_at"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = Capsule(s.tmp)

	r.CreatedAt = time.Time(s.CreatedAt)
	r.UpdatedAt = time.Time(s.UpdatedAt)

	return nil
}

func (r *Container) UnmarshalJSON(b []byte) error {
	type tmp Container
	var s struct {
		tmp
		CreatedAt gophercloud.JSONRFC3339ZNoT `json:"created_at"`
		UpdatedAt gophercloud.JSONRFC3339ZNoT `json:"updated_at"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = Container(s.tmp)

	r.CreatedAt = time.Time(s.CreatedAt)
	r.UpdatedAt = time.Time(s.UpdatedAt)

	return nil
}
