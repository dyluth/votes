package publicwhip

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

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
		return "", errors.New("could not find summary")
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

func GetMPID(name string) {

}
