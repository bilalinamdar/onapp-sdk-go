package onappgo

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/digitalocean/godo"
)

const dataStoresBasePath string = "settings/data_stores"
const dataStoresIoLimitsBasePath string = "settings/data_stores/%d/io_limits"

// DataStoresService is an interface for interfacing with the DataStore
// endpoints of the OnApp API
// https://docs.onapp.com/apim/latest/data-stores
type DataStoresService interface {
	List(context.Context, *ListOptions) ([]DataStore, *Response, error)
	Get(context.Context, int) (*DataStore, *Response, error)
	Create(context.Context, *DataStoreCreateRequest) (*DataStore, *Response, error)
	Delete(context.Context, int, interface{}) (*Response, error)
	Edit(context.Context, int, *DataStoreEditRequest) (*Response, error)

	IoLimits(context.Context, int, *IoLimits) (*Response, error)
}

// DataStoresServiceOp handles communication with the Data Store related methods of the
// OnApp API.
type DataStoresServiceOp struct {
	client *Client
}

var _ DataStoresService = &DataStoresServiceOp{}

type AdminAttributes struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// AccountAttributes - SolidFire account username
type AccountAttributes struct {
	Username        string `json:"username,omitempty"`
	InitiatorSecret string `json:"initiator_secret,omitempty"`
	TargetSecret    string `json:"target_secret,omitempty"`
}

// DataStore represents a DataStore
type DataStore struct {
	ID                             int         `json:"id,omitempty"`
	Label                          string      `json:"label,omitempty"`
	Identifier                     string      `json:"identifier,omitempty"`
	CreatedAt                      string      `json:"created_at,omitempty"`
	UpdatedAt                      string      `json:"updated_at,omitempty"`
	LocalHypervisorID              int         `json:"local_hypervisor_id,omitempty"`
	DataStoreSize                  int         `json:"data_store_size,omitempty"`
	ZombieDisksSize                int         `json:"zombie_disks_size,omitempty"`
	IP                             string      `json:"ip,omitempty"`
	DataStoreGroupID               int         `json:"data_store_group_id,omitempty"`
	Enabled                        bool        `json:"enabled,bool"`
	DataStoreType                  string      `json:"data_store_type,omitempty"`
	IscsiIP                        string      `json:"iscsi_ip,omitempty"`
	HypervisorGroupID              int         `json:"hypervisor_group_id,omitempty"`
	VdcID                          int         `json:"vdc_id,omitempty"`
	IntegratedStorageCacheEnabled  bool        `json:"integrated_storage_cache_enabled,bool"`
	IntegratedStorageCacheSettings interface{} `json:"integrated_storage_cache_settings,omitempty"`
	AutoHealing                    bool        `json:"auto_healing,bool"`
	IoLimits                       IoLimits    `json:"io_limits,omitempty"`
	Epoch                          bool        `json:"epoch,bool"`
	Default                        bool        `json:"default,bool"`
	Usage                          int         `json:"usage,omitempty"`
	Trim                           bool        `json:"trim,bool"`
}

// DataStoreCreateRequest represents a request to create a DataStore
type DataStoreCreateRequest struct {
	Label             string `json:"label,omitempty"`
	DataStoreGroupID  int    `json:"data_store_group_id,omitempty"`
	LocalHypervisorID int    `json:"local_hypervisor_id,omitempty"`
	IP                string `json:"ip,omitempty"`
	Enabled           bool   `json:"enabled,bool"`
	DataStoreSize     int    `json:"data_store_size,omitempty"`
	DataStoreType     string `json:"data_store_type,omitempty"`
	IscsiIP           string `json:"iscsi_ip,omitempty"`
}

// DataStoreSolidFireCreateRequest represents a request to create a SolidFire DataStore
type DataStoreSolidFireCreateRequest struct {
	Label             string `json:"label,omitempty"`
	DataStoreGroupID  int    `json:"data_store_group_id,omitempty"`
	LocalHypervisorID int    `json:"local_hypervisor_id,omitempty"`
	IP                string `json:"ip,omitempty"`
	Enabled           bool   `json:"enabled,bool"`
	DataStoreSize     int    `json:"data_store_size,omitempty"`
	DataStoreType     string `json:"data_store_type,omitempty"`
	IscsiIP           string `json:"iscsi_ip,omitempty"`

	AdminAttributes   AdminAttributes   `json:"admin_attributes,omitempty"`
	AccountAttributes AccountAttributes `json:"account_attributes,omitempty"`
}

// DataStoreEditRequest represents a request to edit a DataStore
type DataStoreEditRequest struct {
	Label             string `json:"label,omitempty"`
	DataStoreGroupID  int    `json:"data_store_group_id,omitempty"`
	LocalHypervisorID int    `json:"local_hypervisor_id,omitempty"`
	IP                string `json:"ip,omitempty"`
	Enabled           bool   `json:"enabled,bool"`
	DataStoreSize     int    `json:"data_store_size,omitempty"`
	DataStoreType     string `json:"data_store_type,omitempty"`
	IscsiIP           string `json:"iscsi_ip,omitempty"`
	Trim              bool   `json:"trim,bool"`
}

type rootIoLimits struct {
	IoLimits *IoLimits `json:"io_limits"`
}

type dataStoreCreateRequestRoot struct {
	DataStoreCreateRequest *DataStoreCreateRequest `json:"data_store"`
}

type dataStoreRoot struct {
	DataStore *DataStore `json:"data_store"`
}

func (d DataStoreCreateRequest) String() string {
	return godo.Stringify(d)
}

// List all DataStores.
func (s *DataStoresServiceOp) List(ctx context.Context, opt *ListOptions) ([]DataStore, *Response, error) {
	path := dataStoresBasePath + apiFormat
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var out []map[string]DataStore
	resp, err := s.client.Do(ctx, req, &out)
	if err != nil {
		return nil, resp, err
	}

	arr := make([]DataStore, len(out))
	for i := range arr {
		arr[i] = out[i]["data_store"]
	}

	return arr, resp, err
}

// Get individual DataStore.
func (s *DataStoresServiceOp) Get(ctx context.Context, id int) (*DataStore, *Response, error) {
	if id < 1 {
		return nil, nil, godo.NewArgError("id", "cannot be less than 1")
	}

	path := fmt.Sprintf("%s/%d%s", dataStoresBasePath, id, apiFormat)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(dataStoreRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.DataStore, resp, err
}

// Create DataStore.
func (s *DataStoresServiceOp) Create(ctx context.Context, createRequest *DataStoreCreateRequest) (*DataStore, *Response, error) {
	if createRequest == nil {
		return nil, nil, godo.NewArgError("DataStore createRequest", "cannot be nil")
	}

	path := dataStoresBasePath + apiFormat
	rootRequest := &dataStoreCreateRequestRoot{
		DataStoreCreateRequest: createRequest,
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, rootRequest)
	if err != nil {
		return nil, nil, err
	}
	log.Println("DataStore [Create] req: ", req)

	root := new(dataStoreRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.DataStore, resp, err
}

// Delete DataStore.
func (s *DataStoresServiceOp) Delete(ctx context.Context, id int, meta interface{}) (*Response, error) {
	if id < 1 {
		return nil, godo.NewArgError("id", "cannot be less than 1")
	}

	path := fmt.Sprintf("%s/%d%s", dataStoresBasePath, id, apiFormat)
	path, err := addOptions(path, meta)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	log.Println("DataStore [Delete] req: ", req)

	return s.client.Do(ctx, req, nil)
}

// Edit DataStore.
func (s *DataStoresServiceOp) Edit(ctx context.Context, id int, editRequest *DataStoreEditRequest) (*Response, error) {
	if id < 1 {
		return nil, godo.NewArgError("id", "cannot be less than 1")
	}

	if editRequest == nil {
		return nil, godo.NewArgError("DataStore [Edit] editRequest", "cannot be nil")
	}

	path := fmt.Sprintf("%s/%d%s", dataStoresBasePath, id, apiFormat)

	req, err := s.client.NewRequest(ctx, http.MethodPut, path, editRequest)
	if err != nil {
		return nil, err
	}
	log.Println("DataStore [Edit]  req: ", req)

	return s.client.Do(ctx, req, nil)
}

// IoLimits edit io limits for DataStore.
func (s *DataStoresServiceOp) IoLimits(ctx context.Context, id int, editRequest *IoLimits) (*Response, error) {
	if id < 1 {
		return nil, godo.NewArgError("id", "cannot be less than 1")
	}

	if editRequest == nil {
		return nil, godo.NewArgError("DataStore [IoLimits] editRequest", "cannot be nil")
	}

	path := fmt.Sprintf(dataStoresIoLimitsBasePath, id) + apiFormat

	rootRequest := &rootIoLimits{
		IoLimits: editRequest,
	}

	req, err := s.client.NewRequest(ctx, http.MethodPut, path, rootRequest)
	if err != nil {
		return nil, err
	}
	log.Println("DataStore [IoLimits]  req: ", req)

	return s.client.Do(ctx, req, nil)
}
