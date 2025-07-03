package postflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	ONOSUsername = "onos"
	ONOSPassword = "rocks"
)


// 流表结构体定义
type Flow struct {
	Priority      int       `json:"priority"`
	Timeout       int       `json:"timeout"`
	IsPermanent   bool      `json:"isPermanent"`
	TableID       string    `json:"tableId"`
	DeviceID      string    `json:"deviceId"`
	Treatment     Treatment `json:"treatment"`
	ClearDeferred bool      `json:"clearDeferred"`
	Selector      Selector  `json:"selector"`
}

type Treatment struct {
	Instructions []Instruction `json:"instructions"`
}

type Instruction struct {
	Type         string                 `json:"type"`
	Subtype      string                 `json:"subtype"`
	ActionID     string                 `json:"actionId"`
	ActionParams map[string]interface{} `json:"actionParams"`
}

type Selector struct {
	Criteria []Criteria `json:"criteria"`
}

type Criteria struct {
	Type    string  `json:"type"`
	Matches []Match `json:"matches"`
}

type Match struct {
	Field string `json:"field"`
	Match string `json:"match"`
	Value string `json:"value"`
}

func PostToOnos(url string, flows interface{}) (string, error) {
	jsonData, err := json.Marshal(flows)
	if err != nil {
		return "marshal error", fmt.Errorf("failed to marshal flows: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "request error", fmt.Errorf("failed to create request: %v", err)
	}

	req.SetBasicAuth("onos", "rocks")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "send request error", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "send request error", fmt.Errorf("ONOS returned status: %s", resp.Status)
	}

	// 可根据实际需要解析返回体
	return "success", nil
}
