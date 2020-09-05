package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gocarina/gocsv"
	"github.com/peterbourgon/ff/v3"
)

func main() {
	fs := flag.NewFlagSet("ge", flag.ExitOnError)
	var (
		cookie    = fs.String("cookie", "develop", "cookie value from browser")
		csrfToken = fs.String("csrf_token", "", "the CSRF token from the browser")
	)

	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarNoPrefix(),
	)

	w := log.NewSyncWriter(os.Stderr)
	l := log.NewLogfmtLogger(w)
	l = log.With(l, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	if *cookie == "" && *csrfToken == "" {
		l.Log("msg", "missing parameters", "err", errors.New("cookies or csrfToken env variables can't be empty"))
		return
	}

	client := &http.Client{
		Transport: &AuthenticatedTransport{
			cookie:    *cookie,
			csrfToken: *csrfToken,
		},
	}
	if err := collect(client, l); err != nil {
		l.Log("msg", "there was an error collecting analytics data", "err", err)
		return
	}
}

func collect(c *http.Client, l log.Logger) error {
	req, err := http.NewRequest("GET", "https://secure.gaug.es/gauges/embedded", nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	var p profile
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return err
	}

	for _, g := range p.Gauges {
		l.Log("msg", "exporting traffic for site", "site", g.Title)
		if err := exportMonth(c, g.Urls.Traffic, g.Title); err != nil {
			l.Log("msg", "error while exporting monthly data", "err", err)
			continue
		}
	}
	return nil
}

func exportMonth(c *http.Client, url string, gaugeName string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	var pd profileData
	if err := json.NewDecoder(resp.Body).Decode(&pd); err != nil {
		return err
	}

	var events []dayTraffic
	for _, d := range pd.Data.Month.Traffic {
		events = append(events, dayTraffic{
			Date:   d.Date,
			Views:  d.Views,
			People: d.People,
		})
	}

	// Get month for file name
	var dateFilename string
	if len(events) > 0 {
		t, err := time.Parse("2006-01-02", events[0].Date)
		if err != nil {
			return err
		}
		dateFilename = t.Format("2006-01")
	}
	outputDirectory := fmt.Sprintf("output/%s", gaugeName)
	if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
		if err := os.Mkdir(outputDirectory, 0700); err != nil {
			return err
		}
	}

	file, err := os.Create(path.Join(outputDirectory, fmt.Sprintf("%s-%s.csv", gaugeName, dateFilename)))
	if err != nil {
		return err
	}
	defer file.Close()

	err = gocsv.MarshalFile(events, file)
	if err != nil {
		return err
	}

	if pd.Urls.Month.Older != "" {
		if err := exportMonth(c, pd.Urls.Older, gaugeName); err != nil {
			return err
		}
	}
	return nil
}

type AuthenticatedTransport struct {
	cookie    string
	csrfToken string
	T         http.RoundTripper
}

func (t *AuthenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	originalRoundtripper := t.T
	if originalRoundtripper == nil {
		originalRoundtripper = http.DefaultTransport
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Cookie", t.cookie)
	req.Header.Add("Accept-Language", "en-us")
	req.Header.Add("Host", "secure.gaug.es")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.2 Safari/605.1.15")
	req.Header.Add("Referer", "https://secure.gaug.es/dashboard")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("X_csrf_token", t.csrfToken)
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	return originalRoundtripper.RoundTrip(req)
}

// dayTraffic contains the aggregated traffic information for one day
type dayTraffic struct {
	Date   string `csv:"date"`
	Views  int    `csv:"views"`
	People int    `csv:"people"`
}

// profile is a user profile on gaug.es
type profile struct {
	Gauges []struct {
		ID           string    `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		CreatorID    string    `json:"creator_id"`
		AllowedHosts string    `json:"allowed_hosts"`
		Title        string    `json:"title"`
		Tz           string    `json:"tz"`
		Enabled      bool      `json:"enabled"`
		NowInZone    time.Time `json:"now_in_zone"`
		Urls         struct {
			Self        string `json:"self"`
			Referrers   string `json:"referrers"`
			Content     string `json:"content"`
			Traffic     string `json:"traffic"`
			Resolutions string `json:"resolutions"`
			Technology  string `json:"technology"`
			Terms       string `json:"terms"`
			Engines     string `json:"engines"`
			Locations   string `json:"locations"`
			Shares      string `json:"shares"`
		} `json:"urls"`
		AllTime struct {
			Views  int `json:"views"`
			People int `json:"people"`
		} `json:"all_time"`
		Today struct {
			Date   string `json:"date"`
			Views  int    `json:"views"`
			People int    `json:"people"`
		} `json:"today"`
		Yesterday struct {
			Date   string `json:"date"`
			Views  int    `json:"views"`
			People int    `json:"people"`
		} `json:"yesterday"`
		RecentHours []struct {
			Hour   string `json:"hour"`
			Views  int    `json:"views"`
			People int    `json:"people"`
		} `json:"recent_hours"`
		RecentDays []struct {
			Date   string `json:"date"`
			Views  int    `json:"views"`
			People int    `json:"people"`
		} `json:"recent_days"`
		RecentMonths []struct {
			Date   string `json:"date"`
			Views  int    `json:"views"`
			People int    `json:"people"`
		} `json:"recent_months"`
		RecentYears []struct {
			Date   string `json:"date"`
			Views  int    `json:"views"`
			People int    `json:"people"`
		} `json:"recent_years"`
	} `json:"gauges"`
}

// profileData is a page of metrics for a gauge in a time range
type profileData struct {
	Date    string `json:"date"`
	Traffic []struct {
		Date   string `json:"date"`
		Views  int    `json:"views"`
		People int    `json:"people"`
	} `json:"traffic"`
	Views  int `json:"views"`
	People int `json:"people"`
	Data   struct {
		Month struct {
			Traffic []struct {
				Date   string `json:"date"`
				Views  int    `json:"views"`
				People int    `json:"people"`
			} `json:"traffic"`
			Views  int `json:"views"`
			People int `json:"people"`
		} `json:"month"`
		Year struct {
			Traffic []struct {
				Date   string `json:"date"`
				Views  int    `json:"views"`
				People int    `json:"people"`
			} `json:"traffic"`
			Views  int `json:"views"`
			People int `json:"people"`
		} `json:"year"`
	} `json:"data"`
	Urls struct {
		Older string      `json:"older"`
		Newer interface{} `json:"newer"`
		Month struct {
			Older string      `json:"older"`
			Newer interface{} `json:"newer"`
		} `json:"month"`
		Year struct {
			Older string      `json:"older"`
			Newer interface{} `json:"newer"`
		} `json:"year"`
	} `json:"urls"`
}
