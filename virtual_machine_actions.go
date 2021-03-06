package onappgo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/digitalocean/godo"
)

// ActionRequest reprents OnApp Action Request
type ActionRequest map[string]interface{}

// VirtualMachineActionsService is an interface for interfacing with the VirtualMachine actions
// endpoints of the OnApp API
type VirtualMachineActionsService interface {
	Shutdown(context.Context, int) (*Transaction, *Response, error)
	Stop(context.Context, int) (*Transaction, *Response, error)
	Startup(context.Context, int) (*Transaction, *Response, error)
	Unlock(context.Context, int) (*Transaction, *Response, error)
	Reboot(context.Context, int) (*Transaction, *Response, error)
	Suspend(context.Context, int) (*Transaction, *Response, error)
	Unsuspend(context.Context, int) (*Transaction, *Response, error)

	ResetPassword(context.Context, int, string, string) (*Transaction, *Response, error)
	FQDN(context.Context, int, string, string) (*Transaction, *Response, error)

	RebuildNetwork(context.Context, int, interface{}) (*Transaction, *Response, error)

	AssignIPAddress(context.Context, int, interface{}) (*Transaction, *Response, error)
	UnAssignIPAddress(context.Context, int, int, interface{}) (*Transaction, *Response, error)
	ListIPAddresses(context.Context, int) (*Transaction, *Response, error)
}

// VirtualMachineActionsServiceOp handles communication with the VirtualMachine action related
// methods of the OnApp API.
type VirtualMachineActionsServiceOp struct {
	client *Client
}

var _ VirtualMachineActionsService = &VirtualMachineActionsServiceOp{}

// Shutdown a VirtualMachine gracefully
func (s *VirtualMachineActionsServiceOp) Shutdown(ctx context.Context, id int) (*Transaction, *Response, error) {
	request := &ActionRequest{"method": http.MethodPost, "type": "shutdown", "action": "stop_virtual_machine"}
	return s.doAction(ctx, id, request, nil, nil)
}

// Stop a VirtualMachine forcefully
func (s *VirtualMachineActionsServiceOp) Stop(ctx context.Context, id int) (*Transaction, *Response, error) {
	request := &ActionRequest{"method": http.MethodPost, "type": "stop", "action": "stop_virtual_machine"}
	return s.doAction(ctx, id, request, nil, nil)
}

// Startup a VirtualMachine
func (s *VirtualMachineActionsServiceOp) Startup(ctx context.Context, id int) (*Transaction, *Response, error) {
	request := &ActionRequest{"method": http.MethodPost, "type": "startup", "action": "startup_virtual_machine"}
	return s.doAction(ctx, id, request, nil, nil)
}

// Unlock a VirtualMachine
func (s *VirtualMachineActionsServiceOp) Unlock(ctx context.Context, id int) (*Transaction, *Response, error) {
	request := &ActionRequest{"method": http.MethodPost, "type": "unlock", "action": "startup_virtual_machine"}
	return s.doAction(ctx, id, request, nil, nil)
}

// Reboot a VirtualMachine
func (s *VirtualMachineActionsServiceOp) Reboot(ctx context.Context, id int) (*Transaction, *Response, error) {
	request := &ActionRequest{"method": http.MethodPost, "type": "reboot", "action": "reboot_virtual_machine"}
	return s.doAction(ctx, id, request, nil, nil)
}

// Suspend a VirtualMachine
func (s *VirtualMachineActionsServiceOp) Suspend(ctx context.Context, id int) (*Transaction, *Response, error) {
	request := &ActionRequest{"method": http.MethodPost, "type": "suspend", "action": "stop_virtual_machine"}
	return s.doAction(ctx, id, request, nil, nil)
}

// Unsuspend a VirtualMachine
func (s *VirtualMachineActionsServiceOp) Unsuspend(ctx context.Context, id int) (*Transaction, *Response, error) {
	request := &ActionRequest{
		"method": http.MethodPost,
		"type":   "unsuspend",
		"path":   "suspend",
		"action": "stop_virtual_machine",
	}
	return s.doAction(ctx, id, request, nil, nil)
}

type resetPassword struct {
	InitialRootPassword              string `json:"initial_root_password,omitempty"`
	InitialRootPasswordEncryptionKey string `json:"initial_root_password_encryption_key,omitempty"`
}

type rootResetPassword struct {
	ResetPassword *resetPassword `json:"virtual_machine"`
}

// ResetPassword a VirtualMachine
func (s *VirtualMachineActionsServiceOp) ResetPassword(ctx context.Context, id int, password string, key string) (*Transaction, *Response, error) {
	request := &ActionRequest{"method": http.MethodPost, "type": "reset_password", "action": "reset_root_password"}

	vmPassword := &resetPassword{
		InitialRootPassword:              password,
		InitialRootPasswordEncryptionKey: key,
	}

	root := &rootResetPassword{
		ResetPassword: vmPassword,
	}

	return s.doAction(ctx, id, request, root, nil)
}

// FQDN a VirtualMachine
func (s *VirtualMachineActionsServiceOp) FQDN(ctx context.Context, id int, hostname string, domain string) (*Transaction, *Response, error) {
	request := &ActionRequest{"method": http.MethodPatch, "type": "fqdn", "action": "update_fqdn"}

	vmFQDN := &VirtualMachine{
		Domain:   domain,
		Hostname: hostname,
	}

	root := &virtualMachineRoot{
		VirtualMachine: vmFQDN,
	}

	return s.doAction(ctx, id, request, root, nil)
}

// VirtualMachineRestartRequest -
type VirtualMachineRestartRequest struct {
	Force int `url:"force"`

	// "hard", "graceful" or "soft"
	ShutdownType    string `url:"shutdown_type"`
	RequiredStartup int    `url:"required_startup"`
}

// RebuildNetwork a VirtualMachine
func (s *VirtualMachineActionsServiceOp) RebuildNetwork(ctx context.Context, id int, opts interface{}) (*Transaction, *Response, error) {
	request := &ActionRequest{"method": http.MethodPost, "type": "rebuild_network", "action": "rebuild_network"}
	return s.doAction(ctx, id, request, nil, opts)
}

type rootIPAddress struct {
	AssignIPAddress *AssignIPAddress `json:"ip_address"`
}

// AssignIPAddress - Assign IPAddress to the VirtualMachine
func (s *VirtualMachineActionsServiceOp) AssignIPAddress(ctx context.Context, id int, params interface{}) (*Transaction, *Response, error) {
	request := &ActionRequest{"method": http.MethodPost, "type": "assign_ip_address", "action": "ip_addresses"}

	// params - must containe required parameters in AssignIPAddress structure
	return s.doAction(ctx, id, request, params, nil)
}

// UnAssignIPAddressRequest -
type UnAssignIPAddressRequest struct {
	RebuildNetwork int `url:"rebuild_network"`
}

// UnAssignIPAddress - UnAssign IPAddress from the VirtualMachine
func (s *VirtualMachineActionsServiceOp) UnAssignIPAddress(ctx context.Context, id int, ipID int, opts interface{}) (*Transaction, *Response, error) {
	request := &ActionRequest{"method": http.MethodDelete, "type": "unassign_ip_address", "action": "ip_addresses", "ip_address_id": ipID}

	// opts - must containe '?rebuild_network=1' url parameter if needed by UnAssignIPAddressRequest structure
	return s.doAction(ctx, id, request, nil, nil)
}

// ListIPAddresses - List IPAddresses from the VirtualMachine
func (s *VirtualMachineActionsServiceOp) ListIPAddresses(ctx context.Context, id int) (*Transaction, *Response, error) {
	request := &ActionRequest{"method": http.MethodGet, "type": "list_ip_address", "action": "ip_addresses"}
	return s.doAction(ctx, id, request, nil, nil)
}

func (s *VirtualMachineActionsServiceOp) doAction(ctx context.Context, id int,
	request *ActionRequest, jsonParams interface{}, urlParams interface{}) (*Transaction, *Response, error) {
	if id < 1 {
		return nil, nil, godo.NewArgError("id", "cannot be less than 1")
	}

	if request == nil {
		return nil, nil, godo.NewArgError("request", "request can't be nil")
	}

	path, err := virtualMachineActionPath(id, request)
	if err != nil {
		return nil, nil, err
	}

	path, err = addOptions(path, urlParams)
	if err != nil {
		return nil, nil, err
	}

	if (*request)["method"] == nil {
		return nil, nil, godo.NewArgError("method", "must be specified")
	}
	httpMethod := (*request)["method"].(string)

	req, err := s.client.NewRequest(ctx, httpMethod, path, jsonParams)
	if err != nil {
		return nil, nil, err
	}

	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return nil, resp, err
	}

	filter := struct {
		Action               string
		AssociatedObjectID   int
		AssociatedObjectType string
	}{
		Action:               (*request)["action"].(string),
		AssociatedObjectID:   id,
		AssociatedObjectType: "VirtualMachine",
	}

	return lastTransaction(ctx, s.client, filter)
}

func virtualMachineActionPath(id int, request *ActionRequest) (string, error) {
	if (*request)["type"] == nil {
		return "", godo.NewArgError("type", "must be specified")
	}

	rtype := (*request)["type"].(string)
	path := rtype

	if (*request)["path"] != nil {
		path = (*request)["path"].(string)
	}

	if rtype == "unassign_ip_address" {
		if (*request)["ip_address_id"] == nil {
			return "", godo.NewArgError("ip_address_id", "must be specified")
		}

		// url - /virtual_machines/:virtual_machine_id/ip_addresses/:id.json
		ipAddressID := (*request)["ip_address_id"].(int)
		res := fmt.Sprintf("%s/%d/%s/%d%s", virtualMachineBasePath, id, path, ipAddressID, apiFormat)
		fmt.Printf("%s >>> url: %s\n", rtype, res)

		return res, nil
	}

	return fmt.Sprintf("%s/%d/%s%s", virtualMachineBasePath, id, path, apiFormat), nil
}
