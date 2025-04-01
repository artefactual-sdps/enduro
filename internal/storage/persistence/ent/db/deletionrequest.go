// Code generated by ent, DO NOT EDIT.

package db

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/aip"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/deletionrequest"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence/ent/db/workflow"
	"github.com/google/uuid"
)

// DeletionRequest is the model entity for the DeletionRequest schema.
type DeletionRequest struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// UUID holds the value of the "uuid" field.
	UUID uuid.UUID `json:"uuid,omitempty"`
	// Requester holds the value of the "requester" field.
	Requester string `json:"requester,omitempty"`
	// RequesterIss holds the value of the "requester_iss" field.
	RequesterIss string `json:"requester_iss,omitempty"`
	// RequesterSub holds the value of the "requester_sub" field.
	RequesterSub string `json:"requester_sub,omitempty"`
	// Reviewer holds the value of the "reviewer" field.
	Reviewer string `json:"reviewer,omitempty"`
	// ReviewerIss holds the value of the "reviewer_iss" field.
	ReviewerIss string `json:"reviewer_iss,omitempty"`
	// ReviewerSub holds the value of the "reviewer_sub" field.
	ReviewerSub string `json:"reviewer_sub,omitempty"`
	// Reason holds the value of the "reason" field.
	Reason string `json:"reason,omitempty"`
	// Status holds the value of the "status" field.
	Status enums.DeletionRequestStatus `json:"status,omitempty"`
	// RequestedAt holds the value of the "requested_at" field.
	RequestedAt time.Time `json:"requested_at,omitempty"`
	// ReviewedAt holds the value of the "reviewed_at" field.
	ReviewedAt time.Time `json:"reviewed_at,omitempty"`
	// AipID holds the value of the "aip_id" field.
	AipID int `json:"aip_id,omitempty"`
	// WorkflowID holds the value of the "workflow_id" field.
	WorkflowID int `json:"workflow_id,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the DeletionRequestQuery when eager-loading is set.
	Edges        DeletionRequestEdges `json:"edges"`
	selectValues sql.SelectValues
}

// DeletionRequestEdges holds the relations/edges for other nodes in the graph.
type DeletionRequestEdges struct {
	// Aip holds the value of the aip edge.
	Aip *AIP `json:"aip,omitempty"`
	// Workflow holds the value of the workflow edge.
	Workflow *Workflow `json:"workflow,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// AipOrErr returns the Aip value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e DeletionRequestEdges) AipOrErr() (*AIP, error) {
	if e.Aip != nil {
		return e.Aip, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: aip.Label}
	}
	return nil, &NotLoadedError{edge: "aip"}
}

// WorkflowOrErr returns the Workflow value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e DeletionRequestEdges) WorkflowOrErr() (*Workflow, error) {
	if e.Workflow != nil {
		return e.Workflow, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: workflow.Label}
	}
	return nil, &NotLoadedError{edge: "workflow"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*DeletionRequest) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case deletionrequest.FieldID, deletionrequest.FieldAipID, deletionrequest.FieldWorkflowID:
			values[i] = new(sql.NullInt64)
		case deletionrequest.FieldRequester, deletionrequest.FieldRequesterIss, deletionrequest.FieldRequesterSub, deletionrequest.FieldReviewer, deletionrequest.FieldReviewerIss, deletionrequest.FieldReviewerSub, deletionrequest.FieldReason, deletionrequest.FieldStatus:
			values[i] = new(sql.NullString)
		case deletionrequest.FieldRequestedAt, deletionrequest.FieldReviewedAt:
			values[i] = new(sql.NullTime)
		case deletionrequest.FieldUUID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the DeletionRequest fields.
func (dr *DeletionRequest) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case deletionrequest.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			dr.ID = int(value.Int64)
		case deletionrequest.FieldUUID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field uuid", values[i])
			} else if value != nil {
				dr.UUID = *value
			}
		case deletionrequest.FieldRequester:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field requester", values[i])
			} else if value.Valid {
				dr.Requester = value.String
			}
		case deletionrequest.FieldRequesterIss:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field requester_iss", values[i])
			} else if value.Valid {
				dr.RequesterIss = value.String
			}
		case deletionrequest.FieldRequesterSub:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field requester_sub", values[i])
			} else if value.Valid {
				dr.RequesterSub = value.String
			}
		case deletionrequest.FieldReviewer:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field reviewer", values[i])
			} else if value.Valid {
				dr.Reviewer = value.String
			}
		case deletionrequest.FieldReviewerIss:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field reviewer_iss", values[i])
			} else if value.Valid {
				dr.ReviewerIss = value.String
			}
		case deletionrequest.FieldReviewerSub:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field reviewer_sub", values[i])
			} else if value.Valid {
				dr.ReviewerSub = value.String
			}
		case deletionrequest.FieldReason:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field reason", values[i])
			} else if value.Valid {
				dr.Reason = value.String
			}
		case deletionrequest.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				dr.Status = enums.DeletionRequestStatus(value.String)
			}
		case deletionrequest.FieldRequestedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field requested_at", values[i])
			} else if value.Valid {
				dr.RequestedAt = value.Time
			}
		case deletionrequest.FieldReviewedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field reviewed_at", values[i])
			} else if value.Valid {
				dr.ReviewedAt = value.Time
			}
		case deletionrequest.FieldAipID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field aip_id", values[i])
			} else if value.Valid {
				dr.AipID = int(value.Int64)
			}
		case deletionrequest.FieldWorkflowID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field workflow_id", values[i])
			} else if value.Valid {
				dr.WorkflowID = int(value.Int64)
			}
		default:
			dr.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the DeletionRequest.
// This includes values selected through modifiers, order, etc.
func (dr *DeletionRequest) Value(name string) (ent.Value, error) {
	return dr.selectValues.Get(name)
}

// QueryAip queries the "aip" edge of the DeletionRequest entity.
func (dr *DeletionRequest) QueryAip() *AIPQuery {
	return NewDeletionRequestClient(dr.config).QueryAip(dr)
}

// QueryWorkflow queries the "workflow" edge of the DeletionRequest entity.
func (dr *DeletionRequest) QueryWorkflow() *WorkflowQuery {
	return NewDeletionRequestClient(dr.config).QueryWorkflow(dr)
}

// Update returns a builder for updating this DeletionRequest.
// Note that you need to call DeletionRequest.Unwrap() before calling this method if this DeletionRequest
// was returned from a transaction, and the transaction was committed or rolled back.
func (dr *DeletionRequest) Update() *DeletionRequestUpdateOne {
	return NewDeletionRequestClient(dr.config).UpdateOne(dr)
}

// Unwrap unwraps the DeletionRequest entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (dr *DeletionRequest) Unwrap() *DeletionRequest {
	_tx, ok := dr.config.driver.(*txDriver)
	if !ok {
		panic("db: DeletionRequest is not a transactional entity")
	}
	dr.config.driver = _tx.drv
	return dr
}

// String implements the fmt.Stringer.
func (dr *DeletionRequest) String() string {
	var builder strings.Builder
	builder.WriteString("DeletionRequest(")
	builder.WriteString(fmt.Sprintf("id=%v, ", dr.ID))
	builder.WriteString("uuid=")
	builder.WriteString(fmt.Sprintf("%v", dr.UUID))
	builder.WriteString(", ")
	builder.WriteString("requester=")
	builder.WriteString(dr.Requester)
	builder.WriteString(", ")
	builder.WriteString("requester_iss=")
	builder.WriteString(dr.RequesterIss)
	builder.WriteString(", ")
	builder.WriteString("requester_sub=")
	builder.WriteString(dr.RequesterSub)
	builder.WriteString(", ")
	builder.WriteString("reviewer=")
	builder.WriteString(dr.Reviewer)
	builder.WriteString(", ")
	builder.WriteString("reviewer_iss=")
	builder.WriteString(dr.ReviewerIss)
	builder.WriteString(", ")
	builder.WriteString("reviewer_sub=")
	builder.WriteString(dr.ReviewerSub)
	builder.WriteString(", ")
	builder.WriteString("reason=")
	builder.WriteString(dr.Reason)
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", dr.Status))
	builder.WriteString(", ")
	builder.WriteString("requested_at=")
	builder.WriteString(dr.RequestedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("reviewed_at=")
	builder.WriteString(dr.ReviewedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("aip_id=")
	builder.WriteString(fmt.Sprintf("%v", dr.AipID))
	builder.WriteString(", ")
	builder.WriteString("workflow_id=")
	builder.WriteString(fmt.Sprintf("%v", dr.WorkflowID))
	builder.WriteByte(')')
	return builder.String()
}

// DeletionRequests is a parsable slice of DeletionRequest.
type DeletionRequests []*DeletionRequest
