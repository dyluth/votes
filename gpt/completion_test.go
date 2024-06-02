package gpt

import (
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func makeAPICallDummy(req *http.Request) ([]byte, error) {
	body := "{\n  \"id\": \"chatcmpl-9OMePTENm35zHRn73ReFGUs2Yqwtk\",\n  \"object\": \"chat.completion\",\n  \"created\": 1715593969,\n  \"model\": \"gpt-4-0613\",\n  \"choices\": [\n    {\n      \"index\": 0,\n      \"message\": {\n        \"role\": \"assistant\",\n        \"content\": null,\n        \"function_call\": {\n          \"name\": \"classify\",\n          \"arguments\": \"{\\n  \\\"prediction\\\": \\\"Incentivise Low Carbon Electricity Generation\\\"\\n}\"\n        }\n      },\n      \"logprobs\": null,\n      \"finish_reason\": \"function_call\"\n    }\n  ],\n  \"usage\": {\n    \"prompt_tokens\": 260,\n    \"completion_tokens\": 22,\n    \"total_tokens\": 282\n  },\n  \"system_fingerprint\": null\n}\n"

	return []byte(body), nil
}

func TestGetTopicOfMessage(t *testing.T) {

	tweet := `Great 
	@Conservatives
	 progress on electric vehicle charge points.
	
	People won’t make the switch to electric if they can’t find a charge point.
	
	This is very good progress! 
	
	My constituency hosts Britain’s first petrol station converted to EV charging, on Fulham Road, SW6.`

	makeAPICall = makeAPICallDummy
	got, err := GetTopicOfMessage("FAKE-KEY", tweet, logrus.New())
	require.NoError(t, err)
	require.Equal(t, "Incentivise Low Carbon Electricity Generation", got)
}

func Test_parseResponseMessage(t *testing.T) {

	gotTopic, err := parseResponseMessage("The topic of the message is likely \"More Emergency Service Workers\".")
	require.NoError(t, err)
	require.Equal(t, "More Emergency Service Workers", gotTopic)

	gotTopic, err = parseResponseMessage("The likely topic of this message is:\n\n- \"Make it easier to trigger a new election for an MP\"")
	require.NoError(t, err)
	require.Equal(t, "Make it easier to trigger a new election for an MP", gotTopic)

}
