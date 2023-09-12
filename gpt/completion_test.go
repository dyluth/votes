package gpt

import "testing"

func Test_parseResponseMessage(t *testing.T) {

	tests := []struct {
		name      string
		msg       string
		wantFit   string
		wantTopic string
		wantErr   bool
	}{
		{
			name:      "2 lines good",
			msg:       "topic: Incentivise Low Carbon Electricity Generation\nfit: high",
			wantFit:   "high",
			wantTopic: "Incentivise Low Carbon Electricity Generation",
		},
		{
			name:      "2 lines low fit",
			msg:       "topic: Incentivise Low Carbon Electricity Generation\nfit: low",
			wantFit:   "low",
			wantTopic: "Incentivise Low Carbon Electricity Generation",
		},
		{
			name:      "2 lines medium fit",
			msg:       "topic: Incentivise Low Carbon Electricity Generation\nfit: medium",
			wantFit:   "medium",
			wantTopic: "Incentivise Low Carbon Electricity Generation",
		},
		{
			name:      "JSON good",
			msg:       `{"topic": "Encourage and incentivise saving","fit": "high"}`,
			wantFit:   "high",
			wantTopic: "Encourage and incentivise saving",
		},
		{
			name:      "bad",
			msg:       `gibberishsad,masdm`,
			wantFit:   "",
			wantTopic: "",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFit, gotTopic, err := parseResponseMessage(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResponseMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFit != tt.wantFit {
				t.Errorf("parseResponseMessage() gotFit = %v, want %v", gotFit, tt.wantFit)
			}
			if gotTopic != tt.wantTopic {
				t.Errorf("parseResponseMessage() gotTopic = %v, want %v", gotTopic, tt.wantTopic)
			}
		})
	}
}
