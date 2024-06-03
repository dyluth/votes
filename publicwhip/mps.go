package publicwhip

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/sirupsen/logrus"
)

var (
	AllMPs   map[string]string // name : id
	Policies map[string]string // name : id
	log      *logrus.Logger
)

func SetupMPs(logger *logrus.Logger) {
	log = logger
	loadMPs()

	for _, v := range AllMPs {
		err := loadAllPoliciesFromCache(v)
		if err != nil {
			Policies = loadAllPolicies(v)
			err := saveAllPoliciesToCache()
			if err != nil {
				log.Warnf("WARNING: failed to save cache of all Policies: %v\n", err)
			}
		}
		log.Info("loaded policies from cache")
		break
	}
}
func saveAllPoliciesToCache() error {
	file, _ := os.OpenFile("./allPoliciesCache.json", os.O_CREATE+os.O_WRONLY, os.ModePerm)
	defer file.Close()
	encoder := json.NewEncoder(file)
	return encoder.Encode(Policies)
}

func loadAllPoliciesFromCache(MPID string) error {
	// try to load from cache
	b, err := os.ReadFile("allPoliciesCache.json")
	if err != nil {
		return err
	}
	// decode straight into Policies
	return json.NewDecoder(bytes.NewBuffer(b)).Decode(&Policies)
}

func GetAllPolicies() []string {
	p := []string{}
	for name := range Policies {
		p = append(p, name)
	}
	return p
}

func GetReducedPolicies() []string {
	log.Warn("WARNING: Using GetReducedPolicies - hardcoded list by Cam")
	return []string{
		"Human Rights and Equality",
		"Bankers' Bonus Tax",
		"Higher Pay for Public Sector Workers",
		"HS2 - In Favour",
		"Incentivise Low Carbon Electricity Generation",
		"Minimum Wage",
		"Right to strike",
		"Public Ownership of Railways",
		"Termination of pregnancy - against",
		"More Emergency Service Workers",
		"Right for EU Citizens in the UK to Stay",
		"Higher taxes on banks",
		"Stop climate change",
		"Prevent abuse of zero hours contracts",
		"Tougher on illegal immigration",
		"Homosexuality - Equal rights",
		"Welfare benefits ought rise in line with prices",
		"Energy Prices - More Affordable",
		"In Favour of Mass Surveillance",
		"Rail Fares - Lower",
		"State control of bus services",
		"Reduce the rate of Corporation Tax",
		"Do more to help refugees inclding children",
		"Asylum System - More strict",
		"Imported Goods Must Equal UK Standards",
		"Higher Benefits for Ill and Disabled",
		"More funds for social care",
		"Reduce Spending on Welfare Benefits",
		"Voting age - Reduce to 16",
		"Increase the income tax - tax free allowance",
		"Use of UK Military Forces Overseas",
		"Deployment of UK armed forces in Afghanistan",
		"Trident replacement - In favour",
		"Support current and former armed service members ",
		"Reduce central funding for local government",
		"Trade Unions - Restrict",
		"Require voters to show photo ID before voting",
		"Openness and Transparency - In Favour",
	}
}

func loadMPs() {
	// try to read the file, if its there, use that
	dat, err := os.ReadFile("./mpData")
	if err != nil { // use hardcoded local path for easier testing
		dat, err = os.ReadFile("/Users/cam/go/src/github.com/dyluth/votes/classifier/publicwhip/mpData")
	}
	if err == nil {
		buff := bytes.NewBuffer(dat)
		d := gob.NewDecoder(buff)

		err = d.Decode(&AllMPs)
		if err != nil {
			panic(err)
		}
		log.Debug("loaded MPs file.. first 10:")
		count := 0
		for name, ID := range AllMPs {
			log.Debugf("MP: %v - %v\n", name, ID)
			count++
			if count > 10 {
				break
			}
		}
		return
	}
	// else calculate the data and create the file
	log.Warn("couldnt find MPs file, scraping from www.publicwhip.org.uk")
	downloadMPData()

	// now write the MPs to a file
	log.Info("now creating MPs file")
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
					log.Debugf("%v,%v\n", name, id)
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
