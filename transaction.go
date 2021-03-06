package onappgo

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/digitalocean/godo"
)

const (
	transactionsBasePath string = "transactions"

	// TransactionRunning is a running transaction status
	TransactionRunning = "running"

	// TransactionComplete is a completed transaction status
	TransactionComplete = "complete"

	// TransactionPending is a pending transaction status
	TransactionPending = "pending"

	// TransactionCancelled is a cancelled transaction status
	TransactionCancelled = "cancelled"

	// TransactionFailed is a failed transaction status
	TransactionFailed = "failed"
)

// TransactionsService handles communction with action related methods of the
// OnApp API: https://docs.onapp.com/apim/latest/transactions
type TransactionsService interface {
	List(context.Context, *ListOptions) ([]Transaction, *Response, error)
	Get(context.Context, int) (*Transaction, *Response, error)

	GetByFilter(context.Context, interface{}, *ListOptions) (*Transaction, *Response, error)
	ListByGroup(context.Context, interface{}, bool, *ListOptions) ([]Transaction, *Response, error)
}

// TransactionsServiceOp handles communition with the image action related methods of the
// OnApp API.
type TransactionsServiceOp struct {
	client *Client
}

var _ TransactionsService = &TransactionsServiceOp{}

// Transaction represents a OnApp Transaction
type Transaction struct {
	Action                 string                 `json:"action,omitempty"`
	Actor                  string                 `json:"actor,omitempty"`
	AllowedCancel          bool                   `json:"allowed_cancel,bool"`
	AssociatedObjectID     int                    `json:"associated_object_id,omitempty"`
	AssociatedObjectType   string                 `json:"associated_object_type,omitempty"`
	ChainID                int                    `json:"chain_id,omitempty"`
	CreatedAt              string                 `json:"created_at,omitempty"`
	DependentTransactionID int                    `json:"dependent_transaction_id,omitempty"`
	ID                     int                    `json:"id,omitempty"`
	Identifier             string                 `json:"identifier,omitempty"`
	LockVersion            int                    `json:"lock_version,omitempty"`
	ParentID               int                    `json:"parent_id,omitempty"`
	ParentType             string                 `json:"parent_type,omitempty"`
	Pid                    int                    `json:"pid,omitempty"`
	Priority               int                    `json:"priority,omitempty"`
	Scheduled              bool                   `json:"scheduled,bool"`
	StartAfter             string                 `json:"start_after,omitempty"`
	StartedAt              string                 `json:"started_at,omitempty"`
	Status                 string                 `json:"status,omitempty"`
	UpdatedAt              string                 `json:"updated_at,omitempty"`
	UserID                 int                    `json:"user_id,omitempty"`
	Params                 map[string]interface{} `json:"params,omitempty"`
}

type transactionRoot struct {
	Transaction *Transaction `json:"transaction"`
}

// List all transactions
func (s *TransactionsServiceOp) List(ctx context.Context, opt *ListOptions) ([]Transaction, *Response, error) {
	path := transactionsBasePath + apiFormat
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var out []map[string]Transaction
	resp, err := s.client.Do(ctx, req, &out)
	if err != nil {
		return nil, resp, err
	}

	trx := make([]Transaction, len(out))
	for i := range trx {
		trx[i] = out[i]["transaction"]
	}

	return trx, resp, err
}

// Get an transaction by ID.
func (s *TransactionsServiceOp) Get(ctx context.Context, id int) (*Transaction, *Response, error) {
	if id < 1 {
		return nil, nil, godo.NewArgError("id", "cannot be less than 1")
	}

	path := fmt.Sprintf("%s/%d%s", transactionsBasePath, id, apiFormat)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(transactionRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Transaction, resp, err
}

// ListByGroup return group of transactions depended by action
func (s *TransactionsServiceOp) ListByGroup(ctx context.Context, meta interface{}, revers bool, opt *ListOptions) ([]Transaction, *Response, error) {
	var associatedObjectID, parentID int
	var associatedObjectType, parentType string

	val := reflect.ValueOf(meta)

	v1 := val.FieldByName("AssociatedObjectID")
	if v1.IsValid() {
		associatedObjectID = v1.Interface().(int)
		// fmt.Printf("associatedObjectID: <%d>\n", associatedObjectID)

		if associatedObjectID < 1 {
			return nil, nil, godo.NewArgError("associatedObjectID", "cannot be less than 1")
		}
	}

	v2 := val.FieldByName("ParentID")
	if v2.IsValid() {
		parentID = v2.Interface().(int)
		// fmt.Printf("parentID: <%d>\n", parentID)

		if parentID < 1 {
			return nil, nil, godo.NewArgError("parentID", "cannot be less than 1")
		}
	}

	v1 = val.FieldByName("AssociatedObjectType")
	if v1.IsValid() {
		associatedObjectType = v1.String()
	}

	v2 = val.FieldByName("ParentType")
	if v2.IsValid() {
		parentType = v2.String()
	}

	lst, resp, err := s.client.Transactions.List(ctx, opt)
	if err != nil {
		return nil, resp, fmt.Errorf("ListByGroup.lst: %s", err)
	}

	var next *Transaction

	len := len(lst)
	var groupList []Transaction

	for i, cur := range lst {
		if associatedObjectType != "" {
			// fmt.Printf("cur.AssociatedObjectID: <%d> -> associatedObjectID: <%d>\n", cur.AssociatedObjectID, associatedObjectID)
			// fmt.Printf("cur.AssociatedObjectType: <%s> -> associatedObjectType: <%s>\n", cur.AssociatedObjectType, associatedObjectType)
			if cur.AssociatedObjectID != associatedObjectID || cur.AssociatedObjectType != associatedObjectType {
				continue
			}
		}

		if parentType != "" {
			// fmt.Printf("cur.ParentID: <%d> -> parentID: <%d>\n", cur.ParentID, parentID)
			// fmt.Printf("cur.ParentType: <%s> -> parentType: <%s>\n", cur.ParentType, parentType)
			if cur.ParentID != parentID || cur.ParentType != parentType {
				continue
			}
		}

		if cur.DependentTransactionID == 0 {
			if revers == false {
				groupList = append(groupList, cur)
			} else {
				groupList = append([]Transaction{cur}, groupList...) // prepend
			}
			break
		}

		if i+1 < len {
			next = &lst[i+1]
		}

		if next != nil {
			if associatedObjectType != "" {
				if cur.AssociatedObjectID == next.AssociatedObjectID &&
					cur.AssociatedObjectType == next.AssociatedObjectType &&
					cur.ChainID == next.ChainID {
					if revers == false {
						groupList = append(groupList, cur)
					} else {
						groupList = append([]Transaction{cur}, groupList...) // prepend
					}
				}
			}

			if parentType != "" {
				if cur.ParentID == next.ParentID &&
					cur.ParentType == next.ParentType &&
					cur.ChainID == next.ChainID {
					if revers == false {
						groupList = append(groupList, cur)
					} else {
						groupList = append([]Transaction{cur}, groupList...) // prepend
					}
				}
			}
		}
	}

	return groupList, resp, err
}

// GetByFilter find transaction with specified fields.
func (s *TransactionsServiceOp) GetByFilter(ctx context.Context, filter interface{}, opts *ListOptions) (*Transaction, *Response, error) {
	lst, resp, err := s.client.Transactions.List(ctx, opts)
	if err != nil {
		return nil, resp, fmt.Errorf("GetByFilter.lst: %s", err)
	}

	for _, v := range lst {
		if v.equal(filter) {
			return &v, resp, err
		}
	}

	return nil, nil, fmt.Errorf("Transaction not found or wrong filter %+v", filter)
}

// EqualFilter -
func (trx *Transaction) EqualFilter(filter interface{}) bool {
	return trx.equal(filter)
}

func (trx *Transaction) equal(filter interface{}) bool {
	val := reflect.ValueOf(filter)
	filterFields := reflect.Indirect(reflect.ValueOf(trx))

	// fmt.Printf("        equal.filterFields: %#v\n", filterFields)
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		value := val.Field(i)
		filterValue := filterFields.FieldByName(typeField.Name)

		// fmt.Printf("%s: %s[%#v] -> %s[%#v]\n", typeField.Name, value.Type(), value, filterValue.Type(), filterValue)

		if value.Interface() != filterValue.Interface() {
			// fmt.Printf("[false] return on filed [%s]\n\n", typeField.Name)
			return false
		}
	}

	// fmt.Printf("[true] transaction found with id[%d]\n\n", trx.ID)
	return true
}

func lastTransaction(ctx context.Context, client *Client, filter interface{}) (*Transaction, *Response, error) {
	opt := &ListOptions{
		PerPage: searchTransactions,
	}

	lst, resp, err := client.Transactions.ListByGroup(ctx, filter, false, opt)
	if lst == nil || err != nil {
		return nil, nil, err
	}

	return &lst[0], resp, err
}

func (trx Transaction) String() string {
	return godo.Stringify(trx)
}

// Running check if transaction state is 'runing'
func (trx Transaction) Running() bool {
	return trx.Status == TransactionRunning
}

// Pending check if transaction state is 'pending'
func (trx Transaction) Pending() bool {
	return trx.Status == TransactionPending
}

// Incomplete check if transaction state is 'running' or 'pending'
func (trx Transaction) Incomplete() bool {
	return trx.Running() || trx.Pending()
}

// Complete check if transaction state is 'complete'
func (trx Transaction) Complete() bool {
	return trx.Status == TransactionComplete
}

// Failed check if transaction state is 'failed'
func (trx Transaction) Failed() bool {
	return trx.Status == TransactionFailed
}

// Cancelled check if transaction state is 'cancelled'
func (trx Transaction) Cancelled() bool {
	return trx.Status == TransactionCancelled
}

// Unlucky check if transaction state is 'failed' or 'cancelled'
func (trx Transaction) Unlucky() bool {
	return trx.Failed() || trx.Cancelled()
}

// Finished check if transaction state is
// 'complete' or 'failed' or 'cancelled'
func (trx Transaction) Finished() bool {
	return trx.Complete() || trx.Unlucky()
}
