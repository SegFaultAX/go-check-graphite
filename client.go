package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

type graphiteClient struct {
	client   *http.Client
	graphite string
	username string
	password string
}

type renderRequest struct {
	Query  string `url:"target"`
	From   string `url:"from"`
	Until  string `url:"until,omitempty"`
	Format string `url:"format"`
}

type metrics []metric

type metric struct {
	Target     string
	Datapoints []datapoint
}

type datapoint struct {
	Value     *float64
	Timestamp time.Time
}

func newClient(graphite, username, password string, timeout int) *graphiteClient {
	cli := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	if !(strings.HasPrefix(graphite, "https://") || strings.HasPrefix(graphite, "http://")) {
		graphite = "http://" + graphite
	}

	return &graphiteClient{
		client:   cli,
		graphite: graphite,
		username: username,
		password: password,
	}
}

func (c *graphiteClient) getMetrics(query, from, until string) (metrics, error) {
	if !strings.HasPrefix(from, "-") && isRelative(from) {
		from = "-" + from
	}
	if !strings.HasPrefix(until, "-") && isRelative(until) {
		until = "-" + until
	}

	resp, err := c.doGET("/render", renderRequest{
		Query:  query,
		From:   from,
		Until:  until,
		Format: "json",
	})
	if err != nil {
		return nil, err
	}

	var ms metrics
	err = json.Unmarshal(resp, &ms)
	if err != nil {
		return nil, err
	}

	return ms, nil
}

func (c *graphiteClient) doGET(path string, params interface{}) ([]byte, error) {
	var qs string
	if params == nil {
		qs = ""
	} else {
		v, err := query.Values(params)
		if err != nil {
			return nil, err
		}
		qs = v.Encode()
	}

	u, err := url.Parse(c.graphite)
	if err != nil {
		return nil, err
	}

	u.Path = path
	u.RawQuery = qs

	//fmt.Printf("Querying: %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (d *datapoint) UnmarshalJSON(in []byte) error {
	var vs []*float64
	err := json.Unmarshal(in, &vs)
	if err != nil {
		return err
	}

	d.Timestamp = time.Unix(int64(*vs[1]), 0.0)
	d.Value = vs[0]

	return nil
}

func isRelative(s string) bool {
	return strings.ContainsAny(s, "smhdwyoin")
}
