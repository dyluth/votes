package publicwhip

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

var (
	mpPolicyPositions     map[string]map[string]string
	mpPolicyPositionsLock = &sync.RWMutex{}
)

const (
	mpPolicyPositionsFilename = "./mpPolicyPositionsCache.json"
)

// converts the names to IDs then calls GetMPPolicyPosition
func GetVoteHistory(mpName, policyName string) (history string, err error) {
	mpName = strings.ToLower(strings.TrimSpace(mpName))
	mpID, ok := AllMPs[mpName]
	if !ok {
		return "", errors.New("could not find mp in list")
	}

	policyID, ok := Policies[policyName]
	if !ok {
		return "", errors.New("could not find policy in list: " + policyName)
	}
	return GetMPPolicyPositionCache(mpID, policyID)
}

func GetMPPolicyPositionCache(mpID, policyID string) (string, error) {
	// Lock the cache
	mpPolicyPositionsLock.Lock()
	defer mpPolicyPositionsLock.Unlock()

	if mpPolicyPositions == nil {
		// try to load from disk

		// try to load from cache
		b, err := os.ReadFile(mpPolicyPositionsFilename)
		if err != nil {
			// if not on disk, initialise the cache
			mpPolicyPositions = make(map[string]map[string]string)
		} else {
			// decode straight into mpPolicyPositions
			err = json.NewDecoder(bytes.NewBuffer(b)).Decode(&mpPolicyPositions)
			if err != nil {
				return "", err
			}
		}
	}
	votes, ok := mpPolicyPositions[mpID]
	if !ok {
		votes = make(map[string]string)
		mpPolicyPositions[mpID] = votes
	}

	vote, ok := votes[policyID]
	if !ok {
		result, err := GetMPPolicyPosition(mpID, policyID)
		if err != nil {
			return "", err
		}
		votes[policyID] = result
		// save to disk
		// TODO maybe move this to a periodic cache destage model - an async process
		// though it will happen less and less as the program runs, so it's probably ok
		file, _ := os.OpenFile(mpPolicyPositionsFilename, os.O_CREATE+os.O_WRONLY, os.ModePerm)
		defer file.Close()
		encoder := json.NewEncoder(file)
		err = encoder.Encode(mpPolicyPositions)
		if err != nil {
			return "", err
		}

	}
	return vote, nil
}

func GetMPPolicyPosition(mpID, policyID string) (string, error) {
	c := colly.NewCollector()

	voteSummary := ""

	c.OnHTML("p.voteexpl", func(e *colly.HTMLElement) {
		// printing all URLs associated with the a links in the page
		thing := e.Text
		voteSummary = thing
	})

	//https://www.publicwhip.org.uk/mp.php?mpid=42172&dmp=6703
	u := fmt.Sprintf("https://www.publicwhip.org.uk/mp.php?mpid=%v&dmp=%v", mpID, policyID)
	c.Visit(u)

	if voteSummary == "" {
		return "", errors.New("could not find summary " + u)
	}
	return voteSummary, nil
}

// we just need an example MP, doesnt matter who
// map is name to policyID
func loadAllPolicies(mpID string) map[string]string {
	fmt.Println("loading all policies")
	policies := make(map[string]string)

	c := colly.NewCollector()

	c.OnHTML("a", func(e *colly.HTMLElement) {
		suffix, found := strings.CutPrefix(e.Attr("href"), "/policy.php?id=")
		if found {
			// fmt.Printf("ID: %v, name: %v\n", suffix, e.Text)
			policies[e.Text] = suffix
		}
	})

	//https://www.publicwhip.org.uk/mp.php?mpid=42172
	u := fmt.Sprintf("https://www.publicwhip.org.uk/mp.php?mpid=%v", mpID)
	c.Visit(u)
	return policies
}

func GetMPID(name string) (string, bool) {

	return "", false
}

func GetPolicyID(name string) (string, bool) {

	return "", false
}
