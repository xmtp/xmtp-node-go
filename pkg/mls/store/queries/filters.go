package queries

import "encoding/json"

type InboxLogFilter struct {
	InboxId    string `json:"inbox_id"`
	SequenceId int64  `json:"sequence_id"`
}

type InboxLogFilterList []InboxLogFilter

func (f *InboxLogFilterList) ToSql() (json.RawMessage, error) {
	jsonBytes, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}
