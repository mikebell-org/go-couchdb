package couchdb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func responseToCouchError(r *http.Response) error {
	e := new(CouchError)
	e.ReturnCode = r.StatusCode
	e.URL = r.Request.URL.String()
	j := json.NewDecoder(r.Body)
	err := j.Decode(e)
	if err != nil {
		e.Err = err.Error()
	}
	return e
}

type CouchError struct {
	ReturnCode int
	URL        string
	Err        string `json:"error"`
	Reason     string `json:"reason"`
}

func (c *CouchError) Error() string {
	if c.ReturnCode == 0 {
		if c.Err == "" {
			panic("Internal error in couchdb library, c.ReturnCode == 0 and c.Err == \"\"")
		}
		return c.Err
	}
	return fmt.Sprintf("URL: %s, HTTP response: %d, Error: %s, Reason: %s", c.URL, c.ReturnCode, c.Err, c.Reason)
}

type CouchSuccess struct {
	// {"ok":true,"id":"bob","rev":"1-967a00dff5e02add41819138abb3284d"}
	OK  bool   `json:"ok"`
	ID  string `json:"id"`
	Rev string `json:"rev"`
}

type CouchInfo struct {
	Name               string      `json:"db_name"`
	DocCount           int         `json:"doc_count"`
	DocDelCount        int         `json:"doc_del_count"`
	UpdateSeq          interface{} `json:"update_seq"`
	PurgeSeq           int         `json:"purge_seq"`
	CompactRunning     bool        `json:"compact_running"`
	DiskSize           int         `json:"disk_size"`
	InstanceStartTime  string      `json:"instance_start_time"`
	DiskFormatVersion  int         `json:"disk_format_version"`
	CommittedUpdateSeq int         `json:"committed_update_seq"`
}

/* {
"seq":"15906810-g1AAAAXReJyV1DlIA0EUBuBFrWJhId5oJFoEi2C80Eob7_t4tWSywRBCApqIjYittdp4xPuY3lZbsbcTxFYU8Yi3jv9Mt83K2-Zv9oP5581M3LKs0miubflsEU5ORzps0RqIJlPRSMiurw8GwvFk2g4lUoFEJBXHvzkhS3ikPI4Jy1N254RNrlAzKYqIxrWtPnfaBncLNiG8Sv1oWzPrtEF3C_Yr6qQ81Nabwymr2ZFoJhrVttTP6gs2JtqV-tK2opDVF-xb9Em5r60_n9UX7EAQ0bBZs9dp29wt2IiYVOrDrHmZ1RfsU8Sk3DVnY5XVF2xPpIkGta2cYvUFGxILSr0Z28LqC_aeyLO2pVzSuirIaQy4A4joJ1rRvnie0xpwABCRVSpjdu2K0xzwFRCRwdXSvvyUc7oBtwARvUQn2tcWMPv3ASKelTozt_rG6Rv_8y-AiA0pL8z-LzL7bwIiuokutS954EwfsAcQ8ajUtXkdfMz-T4CINSlvzY2ZY85_HRDRSZQ1879n9u8CRCh8Zv9mYn9OHvC1",
"id":"1cab9146e8abf70e3387d22016294bae:68793e573e9404163cfeff8e3d5a98dd",
"changes":[{"rev":"2-490b51dbd75b248d4b519c01742bf237"}],
"deleted":true
}, */

type DocRev struct {
	ID      string          `json:"id"`
	Seq     interface{}     `json:"seq"`
	Doc     json.RawMessage `json:"doc"`
	Deleted bool            `json:"deleted"`
	//	Changes []something
}

type ViewResults struct {
	TotalRows int `json:"total_rows"`
	Offset    int `json:"offset"`
	UpdateSeq int `json:"update_seq"`
	Rows      []ViewRow
}

type ViewRow struct {
	ID    string          `json:"id"`
	Key   interface{}     `json:"key"`
	Value interface{}     `json:"value"`
	Doc   json.RawMessage `json:"doc"`
}

type BulkCommitResponseRow struct {
	ID     string `json:"id"`
	Rev    string `json:"rev"`
	Error  string `json:"error"`
	Reason string `json:"reason"`
}

type BulkCommitResponse []BulkCommitResponseRow

type BulkCommit struct {
	AllOrNothing bool          `json:"all_or_nothing,omitempty"`
	Docs         []interface{} `json:"docs"`
}

//type CouchDocument map[string]interface{}
