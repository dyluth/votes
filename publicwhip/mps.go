package publicwhip

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

var (
	AllMPs   map[string]string // name : id
	Policies map[string]string // name : id
)

func SetupMPs() {
	loadMPs()
	for _, v := range AllMPs {
		Policies = loadAllPolicies(v) // not cached locally
		break
	}
}

func GetAllPolicies() []string {
	p := []string{}
	for name := range Policies {
		p = append(p, name)
	}
	return p
}

func GetReducedPolicies() []string {
	fmt.Println("WARNING: Using GetReducedPolicies - hardcoded list by Cam")
	return []string{
		"Bankers' Bonus Tax",
		"Higher Pay for Public Sector Workers",
		"HS2 - In Favour",
		"Incentivise Low Carbon Electricity Generation",
		"Minimum Wage",
		"Right to strike",
		"Public Ownership of Railways",
		"Require voters to show photo ID before voting",
		"Termination of pregnancy - against",
		"More Emergency Service Workers",
		"Right for EU Citizens in the UK to Stay",
		"Higher taxes on banks",
		"Stop climate change",
		"Prevent abuse of zero hours contracts",
		"Tougher on illegal immigration",
		"Homosexuality - Equal rights",
		"None",
	}
}

func loadMPs() {
	// try to read the file, if its there, use that
	dat, err := os.ReadFile("./mpData")
	if err == nil {
		buff := bytes.NewBuffer(dat)
		d := gob.NewDecoder(buff)

		err = d.Decode(&AllMPs)
		if err != nil {
			panic(err)
		}
		fmt.Println("loaded MPs file.. first 10:")
		count := 0
		for name := range AllMPs {
			fmt.Printf("MP: %v\n", name)
			count++
			if count > 10 {
				break
			}
		}
		return
	}
	// else calculate the data and create the file
	fmt.Println("couldnt find MPs file, scraping from www.publicwhip.org.uk")
	downloadMPData()

	// now write the MPs to a file
	fmt.Println("now creating MPs file")
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	// Encoding the map
	err = e.Encode(AllMPs)
	if err != nil {
		panic(err)
	}
	f, err := os.Create("./mpData")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.Write(b.Bytes())
	if err != nil {
		panic(err)
	}
}

func downloadMPData() {
	AllMPs = map[string]string{}
	// load the all MPs page
	//view-source:https://www.publicwhip.org.uk/mps.php?sort=party

	//<tr class="">
	//  <td>
	//    <a href="/mp.php?mpn=Paul_Holmes&mpc=Eastleigh&house=commons"> Paul Holmes</a>  <-- this URL
	//  </td>
	//  <td>
	//    <a href="/mp.php?mpc=Eastleigh&house=commons">Eastleigh</a>
	//  </td>
	//  <td>Con</td>
	//  <td class="percent">0.4%</td>
	//</tr>

	// find "a" where it matches this pattern: /mp.php?mpn=.*&mpc=.*&house=.*
	// USE the url parser to make work
	// eg `/mp.php?mpn=Paul_Holmes&mpc=Eastleigh&house=commons`

	c := colly.NewCollector()

	c.OnHTML("a", func(e *colly.HTMLElement) {

		u, err := url.Parse(e.Attr("href"))
		if err != nil {
			panic(err)
		}
		if u.Path == "/mp.php" {
			query := u.Query()
			if query.Get("mpn") != "" && query.Get("mpc") != "" && query.Get("house") != "" {
				//then we need to load that URL and get the mpID from it
				id, err := getIDfromMPURL(e.Attr("href"))
				if err == nil {
					name := strings.ToLower(strings.TrimSpace(e.Text))
					AllMPs[name] = id
					fmt.Printf("%v,%v\n", name, id)
				}
			}
		}
	})
	c.Visit("https://www.publicwhip.org.uk/mps.php?sort=party")

}

// used in setup
func getIDfromMPURL(url string) (string, error) {
	url = "https://www.publicwhip.org.uk" + url
	//fmt.Printf("getIDfromMPURL: %v\n", url)

	//searching for a URL that looks like: http://www.theyworkforyou.com/mp/?m=42172
	c := colly.NewCollector()
	mpID := ""

	c.OnResponse(func(resp *colly.Response) {
		if resp == nil {
			return
		}
		re := regexp.MustCompile(`http:\/\/www\.theyworkforyou\.com\/mp\/\?m=(\d+)`)
		matches := re.FindStringSubmatch(string(resp.Body))
		if len(matches) == 2 {
			mpID = matches[1]
		}
	})

	c.OnHTML("ul.li", func(e *colly.HTMLElement) {
		suffix, found := strings.CutPrefix(e.Attr("href"), "http://www.theyworkforyou.com/mp/?m=")
		if found {
			mpID = suffix
		}
	})

	c.Visit(url)
	if mpID == "" {
		return mpID, errors.New("couldnt find id for MP at " + url)
	}
	return mpID, nil
}
