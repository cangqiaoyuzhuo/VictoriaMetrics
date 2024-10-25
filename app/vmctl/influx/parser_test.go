package influx

import (
	"encoding/json"
	"reflect"
	"testing"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/models"
)

func TestSeriesUnmarshal(t *testing.T) {
	f := func(s string, resultExpected *Series) {
		t.Helper()

		result := &Series{}
		if err := result.unmarshal(s); err != nil {
			t.Fatalf("cannot unmarshal series from %q: %s", s, err)
		}
		if !reflect.DeepEqual(result, resultExpected) {
			t.Fatalf("unexpected result\ngot\n%v\nwant\n%v", result, resultExpected)
		}
	}

	tag := func(name, value string) LabelPair {
		return LabelPair{
			Name:  name,
			Value: value,
		}
	}
	series := func(measurement string, lp ...LabelPair) *Series {
		return &Series{
			Measurement: measurement,
			LabelPairs:  lp,
		}
	}

	f("cpu", series("cpu"))

	f("cpu,host=localhost", series("cpu", tag("host", "localhost")))

	f("cpu,host=localhost,instance=instance", series("cpu", tag("host", "localhost"), tag("instance", "instance")))

	f(`fo\,bar\=baz,x\=\b=\\a\,\=\q\ `, series("fo,bar=baz", tag(`x=\b`, `\a,=\q `)))

	f("cpu,host=192.168.0.1,instance=fe80::fdc8:5e36:c2c6:baac%utun1", series("cpu", tag("host", "192.168.0.1"), tag("instance", "fe80::fdc8:5e36:c2c6:baac%utun1")))

	f(`cpu,db=db1,host=localhost,server=host\=localhost\ user\=user\ `, series("cpu", tag("db", "db1"), tag("host", "localhost"), tag("server", "host=localhost user=user ")))
}

func TestToFloat64_Failure(t *testing.T) {
	f := func(in any) {
		t.Helper()

		_, err := toFloat64(in)
		if err == nil {
			t.Fatalf("expecting non-nil error")
		}
	}

	f("text")
}

func TestToFloat64_Success(t *testing.T) {
	f := func(in any, resultExpected float64) {
		t.Helper()

		result, err := toFloat64(in)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if result != resultExpected {
			t.Fatalf("unexpected result: got %v; want %v", result, resultExpected)
		}
	}

	f("123.4", 123.4)
	f(float64(123.4), 123.4)
	f(float32(12), 12)
	f(123, 123)
	f(true, 1)
	f(false, 0)
	f(json.Number("123456.789"), 123456.789)
}

func TestParseResultCheckTags(t *testing.T) {
	f := func(wantedColumn, queryColumn []string, queryRow []interface{}, drop, wantErr bool) {
		t.Helper()
		lp := make([]LabelPair, 0, len(wantedColumn))
		for _, columnName := range wantedColumn {
			lp = append(lp, LabelPair{Name: columnName, Value: ""})
		}

		series := &Series{
			Measurement: "measurement",
			Field:       "value",
			LabelPairs:  lp,
		}
		ir := influx.Result{
			Series: []models.Row{
				{
					Name:    "test_table",
					Columns: append(queryColumn, "value", "time"),
					Values:  [][]interface{}{append(queryRow, "10", "1729829153000000000")},
				},
			},
		}
		result, err := parseResultCheckTags(series, ir)
		if wantErr != (err != nil) {
			t.Fatalf("want err: %v, got error: %s", wantErr, err)
		}

		// should drop but not dropped
		if err == nil && drop == (len(result[0].values) > 0) {
			t.Fatalf("retained incorrect rows, wantedColumn %v, queryColumn %v, queryRow %v, dropped %v", wantedColumn, queryColumn, queryRow, !(len(result) > 0))
		}
	}

	// only matched series columns
	f([]string{"a", "b"}, []string{"a", "b"}, []interface{}{"1", "2"}, false, false)
	// other series columns exist but row valid
	f([]string{"a", "b"}, []string{"a", "b", "c", "d"}, []interface{}{"1", "2", nil, nil}, false, false)
	// other series columns exist and no valid rows
	f([]string{"a", "b"}, []string{"a", "b", "c", "d"}, []interface{}{"1", "2", "3", "4"}, true, false)
	// invalid query result with fewer columns than expect
	f([]string{"a", "b", "c", "d"}, []string{"a", "b"}, []interface{}{"1", "2"}, true, true)
	// invalid query result with values of a row don't match columns
	f([]string{"a", "b"}, []string{"a", "b"}, []interface{}{"1", "2", "3"}, true, true)
}
