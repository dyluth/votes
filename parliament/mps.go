package parliament

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// finds currently serving MPs with the same name
func GetCurrentByName(name string) ([]Member, error) {
	found, err := GetByName(name)
	if err != nil {
		panic(err)
	}
	// filter them down
	fmt.Printf("found %v matches\n", len(found.Members))

	goodMembers := []Member{}
	for i, m := range found.Members {
		// if they are not an active MP, skip them
		if m.Value.LatestHouseMembership.MembershipEndDate != nil {
			continue
		}
		goodMembers = append(goodMembers, found.Members[i].Value)
	}
	return goodMembers, nil

}

func GetByName(name string) (*GetForNameBody, error) {
	// https://members-api.parliament.uk/api/Members/Search?Name=Paul%20Holmes&skip=0&take=20

	req, err := http.NewRequest(http.MethodGet, "https://members-api.parliament.uk/api/Members/Search", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	q.Add("Name", name)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{} // TODO better client
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("expecting 2xx response code, received %v", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	results := GetForNameBody{}

	err = json.Unmarshal(bodyBytes, &results)
	if err != nil {
		return nil, err
	}
	return &results, nil
}

type GetForNameBody struct {
	Members []struct {
		Value Member            `json:"value"`
		Links []PaginationLinks `json:"links"` // valid links are: "self", "overview", "synopsis" & "contactInformation"
	} `json:"items"`
}

// "json: cannot unmarshal object into Go struct field .items.value.latestHouseMembership.membershipStatus of type string"
type Member struct {
	ID            int         `json:"id"`
	NameListAs    string      `json:"nameListAs"`
	NameDisplayAs string      `json:"nameDisplayAs"`
	NameFullTitle string      `json:"nameFullTitle"`
	NameAddressAs interface{} `json:"nameAddressAs"`
	LatestParty   struct {
		ID                    int         `json:"id"`
		Name                  string      `json:"name"`
		Abbreviation          string      `json:"abbreviation"`
		BackgroundColour      string      `json:"backgroundColour"`
		ForegroundColour      string      `json:"foregroundColour"`
		IsLordsMainParty      bool        `json:"isLordsMainParty"`
		IsLordsSpiritualParty bool        `json:"isLordsSpiritualParty"`
		GovernmentType        interface{} `json:"governmentType"`
		IsIndependentParty    bool        `json:"isIndependentParty"`
	} `json:"latestParty"`
	Gender                string `json:"gender"`
	LatestHouseMembership struct {
		MembershipFrom           string      `json:"membershipFrom"`
		MembershipFromID         int         `json:"membershipFromId"`
		House                    int         `json:"house"`
		MembershipStartDate      string      `json:"membershipStartDate"`
		MembershipEndDate        interface{} `json:"membershipEndDate"`
		MembershipEndReason      interface{} `json:"membershipEndReason"`
		MembershipEndReasonNotes interface{} `json:"membershipEndReasonNotes"`
		MembershipEndReasonID    interface{} `json:"membershipEndReasonId"`
		MembershipStatus         interface{} `json:"membershipStatus"`
	} `json:"latestHouseMembership"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

// parse the top level items into this struct to get the pagination results
type Pagination struct {
	TotalResults  int               `json:"totalResults"`
	ResultContext string            `json:"resultContext"`
	Skip          int               `json:"skip"`
	Take          int               `json:"take"`
	Links         []PaginationLinks `json:"links"`
}

type PaginationLinks struct {
	Rel    string `json:"rel"`
	Href   string `json:"href"`
	Method string `json:"method"`
}
