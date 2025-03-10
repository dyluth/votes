package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/dyluth/votes/parliament"
)

func main() {

	goodMembers, err := parliament.GetCurrentByName("Paul Holmes")
	if err != nil {
		panic(err)
	}
	for _, m := range goodMembers {
		fmt.Printf("%v - %v is a %v MP for %v\n", m.ID, m.NameDisplayAs, m.LatestParty.Name, m.LatestHouseMembership.MembershipFrom)
	}

	if len(goodMembers) != 1 {
		panic(fmt.Sprintf("we did not find exaclty 1 MP %v", len(goodMembers)))
	}

	mp := goodMembers[0]
	GetAllVotesForMP(mp.ID)

}

func GetAllVotesForMP(mpID int) ([]string, error) {

	//house=1&page=1
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://members-api.parliament.uk/api/Members/%v/Voting", mpID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	q.Add("house", "1") // house of commons
	q.Add("page", "1")  // pagination??!?
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
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)

	// results := GetForNameBody{}

	// err = json.Unmarshal(bodyBytes, &results)
	// if err != nil {
	// 	return nil, err
	// }
	return nil, nil
}

type AutoGenerated struct {
	Items []struct {
		Value struct {
			House              int    `json:"house"`
			ID                 int    `json:"id"`
			InAffirmativeLobby bool   `json:"inAffirmativeLobby"`
			ActedAsTeller      bool   `json:"actedAsTeller"`
			Title              string `json:"title"`
			Date               string `json:"date"`
			DivisionNumber     int    `json:"divisionNumber"`
			NumberInFavour     int    `json:"numberInFavour"`
			NumberAgainst      int    `json:"numberAgainst"`
		} `json:"value"`
		Links []struct {
			Rel    string `json:"rel"`
			Href   string `json:"href"`
			Method string `json:"method"`
		} `json:"links"`
	} `json:"items"`
}
