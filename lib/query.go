package lib

import (
	"fmt"
	"strings"

	"github.com/hillside-labs/chet-client/models"
	"gorm.io/gorm"
)

type JSONQuery struct {
	Operation  string          `json:"operation"`
	Table      string          `json:"table"`
	Conditions []JSONCondition `json:"conditions"`
	OrderBy    []JSONOrderBy   `json:"orderBy"`
}

func (jq JSONQuery) String() string {
	var builder strings.Builder

	builder.WriteString("Operation: " + jq.Operation + "\n")
	builder.WriteString("Table: " + jq.Table + "\n")

	builder.WriteString("Conditions:\n")
	for _, condition := range jq.Conditions {
		builder.WriteString("  " + condition.String() + "\n")
	}

	builder.WriteString("OrderBy:\n")
	for _, orderBy := range jq.OrderBy {
		builder.WriteString("  " + orderBy.String() + "\n")
	}

	return builder.String()

}

type JSONCondition struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

func (jc JSONCondition) String() string {
	return fmt.Sprintf("JSONCondition:{Column: %s, Operator: %s, Value: %s}\n", jc.Column, jc.Operator, jc.Value)
}

type JSONOrderBy struct {
	Column string `json:"column"`
	Order  string `json:"order"`
}

func (job JSONOrderBy) String() string {
	return fmt.Sprintf("JSONOrderBy:{Column: %s, Order: %s}\n", job.Column, job.Order)
}

func (jq *JSONQuery) Query(db *gorm.DB) ([]models.Record, error) {
	gormDB := db.Table(jq.Table)

	for _, condition := range jq.Conditions {
		gormDB = gormDB.Where(condition.Column+" "+condition.Operator+" ?", condition.Value)
	}

	for _, orderBy := range jq.OrderBy {
		gormDB = gormDB.Order(orderBy.Column + " " + orderBy.Order)
	}

	var results []models.Record
	err := gormDB.Find(&results).Error

	return results, err
}

func NewJSONQueryFromQuery(query string) (*JSONQuery, error) {
	parts := strings.Fields(query)
	jsonQuery := &JSONQuery{
		Conditions: make([]JSONCondition, 0),
		OrderBy:    make([]JSONOrderBy, 0),
	}

	for _, part := range parts {
		if strings.HasPrefix(part, "ORDER BY") {
			orderParts := strings.Fields(part[8:])
			if len(orderParts) == 2 {
				ob := JSONOrderBy{
					Column: orderParts[0],
					Order:  orderParts[1],
				}
				jsonQuery.OrderBy = append(jsonQuery.OrderBy, ob)
			}
		} else {
			conditionParts := strings.SplitN(part, ":", 2)
			if len(conditionParts) == 2 {
				col := conditionParts[0]
				switch col {
				case "team":
					col = "team_name"
				case "user":
					col = "user_email"
				}
				cond := JSONCondition{
					Column:   col,
					Operator: "=",
					Value:    conditionParts[1],
				}
				jsonQuery.Conditions = append(jsonQuery.Conditions, cond)
			}
		}
	}

	if len(jsonQuery.Conditions) == 0 {
		return nil, fmt.Errorf("at least one condition is required")
	}

	// Define the table name (change this as needed)
	jsonQuery.Table = "records"

	return jsonQuery, nil
}
