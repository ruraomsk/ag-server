package svg

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

type Phases struct {
	Phases []Phase `json:"phases"`
}
type Cameras struct {
	Cameras []Camera `json:"cameras"`
}
type Extend struct {
	Phases  []Phase  `json:"phases"`
	Cameras []Camera `json:"cameras"`
}
type Phase struct {
	Number string `json:"num"`
	PNG    string `json:"phase"`
}
type Camera struct {
	Name string `json:"name"`
	Cam  int    `json:"cam"`
	Area int    `json:"area"`
	Ip   string `json:"ip"`
}

func Parse(path string) ([]byte, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return []byte("{}"), err
	}
	strs := strings.Split(string(file), "\x0a")
	pos := 0
	for pos = 0; pos < len(strs); pos++ {
		if strings.Contains(strs[pos], "function getPhasesMass(){") {
			break
		}
	}
	pos++
	if pos >= len(strs) {
		ext := Extend{Phases: make([]Phase, 0), Cameras: make([]Camera, 0)}
		buff, err := json.Marshal(ext)
		return buff, err
	}
	str := strs[pos]
	if strings.HasSuffix(str, "mass = [") {
		str = ""
	}
	for ; pos < len(strs); pos++ {
		if strings.HasSuffix(strs[pos], "];") {
			break
		}
		str += strs[pos]
	}
	for strings.Contains(str, " ") {
		str = strings.ReplaceAll(str, " ", "")
	}
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, "varmass=", "")
	str = strings.ReplaceAll(str, "{", "{\"")
	str = strings.ReplaceAll(str, ":", "\":")
	str = strings.ReplaceAll(str, "'", "\"")
	str = strings.ReplaceAll(str, "\",", "\",\"")
	str += "]}"
	str = `{"phases":` + str
	var phs Phases
	err = json.Unmarshal([]byte(str), &phs)
	if err != nil {
		return []byte("{}"), err
	}
	pos++
	var cams Cameras
	cams.Cameras = make([]Camera, 0)
	for ; pos < len(strs); pos++ {
		if strings.Contains(strs[pos], "function getAnglesCamera(){") {
			pos++
			str = strs[pos]
			for strings.Contains(str, " ") {
				str = strings.ReplaceAll(str, " ", "")
			}
			if !strings.Contains(str, "[];") {
				str = strings.ReplaceAll(str, "...:", "\"")
				str = strings.ReplaceAll(str, "return", "")
				str = strings.ReplaceAll(str, "{", "{\"")
				str = strings.ReplaceAll(str, "e:", "e\":")
				str = strings.ReplaceAll(str, "m:", "m\":")
				str = strings.ReplaceAll(str, "a:", "a\":")
				str = strings.ReplaceAll(str, "p:", "p\":")
				str = strings.ReplaceAll(str, "'", "\"")
				str = strings.ReplaceAll(str, ",", ",\"")
				str = strings.ReplaceAll(str, "];};", "")
				str = strings.ReplaceAll(str, ",\"{", ",{")
				str = strings.ReplaceAll(str, "\"\"\"", "\"\"")
				str = strings.ReplaceAll(str, "}\"", "}")
				str = strings.ReplaceAll(str, ";]", "")
				str += "]}"
				str = `{"cameras":` + str
				err = json.Unmarshal([]byte(str), &cams)
				if err != nil {
					return []byte("{}"), err
				}
			}
			break
		}
	}
	var ext Extend
	ext.Cameras = cams.Cameras
	ext.Phases = phs.Phases
	buff, err := json.Marshal(ext)
	return buff, err
}
