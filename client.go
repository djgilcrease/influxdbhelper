package influxdbhelper

import (
	"regexp"
	"strings"

	client "github.com/influxdata/influxdb/client/v2"
)

var reRemoveExtraSpace = regexp.MustCompile(`\s\s+`)

func CleanQuery(query string) string {
	ret := strings.Replace(query, "\n", "", -1)
	ret = reRemoveExtraSpace.ReplaceAllString(ret, " ")
	return ret
}

type Client struct {
	url       string
	client    client.Client
	precision string
}

func NewClient(url, user, passwd, precision string) (*Client, error) {
	ret := Client{
		url:       url,
		precision: precision,
	}

	client, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     url,
		Username: user,
		Password: passwd,
	})

	ret.client = client

	return &ret, err
}

func (c Client) InfluxClient() client.Client {
	return c.client
}

func (c Client) Query(db, cmd string, result interface{}) (err error) {
	query := client.Query{
		Command:   cmd,
		Database:  db,
		Chunked:   false,
		ChunkSize: 100,
	}

	var response *client.Response
	response, err = c.client.Query(query)

	if response.Error() != nil {
		return response.Error()
	}

	if err != nil {
		return
	}

	results := response.Results
	if len(results) < 1 || len(results[0].Series) < 1 {
		return
	}

	series := results[0].Series[0]

	err = Decode(series.Columns, series.Values, result)

	return
}

func (c Client) WritePoint(db, measurement string, data interface{}) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  db,
		Precision: c.precision,
	})

	if err != nil {
		return err
	}

	t, tags, fields, err := Encode(data)

	if err != nil {
		return err
	}

	pt, err := client.NewPoint(measurement, tags, fields, t)

	if err != nil {
		return err
	}

	bp.AddPoint(pt)

	return c.client.Write(bp)
}
