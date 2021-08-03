package benchdaily

import (
	"encoding/json"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func LoadDataDir(dir string) ([]BenchOutput, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	res := make([]BenchOutput, 0, 100)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		f, err := os.Open(path.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		defer f.Close()

		var b BenchOutput
		dec := json.NewDecoder(f)
		err = dec.Decode(&b)
		if err != nil {
			return nil, err
		}
		res = append(res, b)
	}
	return res, nil
}

func WriteJSONFile(outputFile string, data interface{}) error {
	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	return enc.Encode(data)
}

func FileName(date time.Time, githash string) string {
	return date.Format("2006-01-02") + "_" + githash + ".json"
}

func UnixDateToTime(date string) (t time.Time, err error) {
	var v int64
	v, err = strconv.ParseInt(date, 10, 64)
	if err != nil {
		return
	}
	t = time.Unix(v, 0)
	return
}
