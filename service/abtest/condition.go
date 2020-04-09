package abtest

import (
	"strings"
	"bytes"
	"strconv"
	"regexp"
	"encoding/json"
)

const VALUE_TRIM = " '\""
const KEY_REGEX = "(?:^|\\s*(and|or))\\s*(\\w+)(\\s*[=<>≤≥≠]|!=|\\s+(?:in|not|is))"

type Formula struct {
	Key				string 		`json:"key"`
	Formula 		string 		`json:"formula"`
	Value			string 		`json:"value"`
}

func (self *Formula) Calculate(vals map[string]interface{}) bool {
	if mVal, ok := vals[self.Key]; ok {
		return self.Calculate4Value(mVal)
	}
	return false
}

func (self *Formula) Calculate4Value(mVal interface{}) bool {
	result := false
	if self.Formula == "in" {
		if strings.HasPrefix(self.Value, "(") && strings.HasSuffix(self.Value, ")") {
			inFormula := &Formula{Key: self.Key, Formula: "=", Value: ""}
			for _, inVal := range strings.Split(self.Value[1: len(self.Value)-1], ",") {
				inFormula.Value = strings.Trim(inVal, VALUE_TRIM)
				if inFormula.Calculate4Value(mVal) {
					result = true
					break
				}
			}
		}
	} else {
		switch mValTypeValue := mVal.(type) {
			case int64, int32, int:
				if fValue, fErr := strconv.ParseInt(self.Value, 10, 64); fErr == nil {
					var mValInt int64
					switch mValTypeValue2 := mValTypeValue.(type) {
						case int64: mValInt = mValTypeValue2
						case int32: mValInt = int64(mValTypeValue2)
						case int: mValInt = int64(mValTypeValue2)
					}
					switch self.Formula {
						case "=", "is": result = mValInt == fValue
						case ">": result = mValInt > fValue
						case "<": result = mValInt < fValue
						case "≥", ">=": result = mValInt >= fValue
						case "≤", "<=": result = mValInt <= fValue
						case "≠", "!=", "not": result = mValInt != fValue
					}
				}
			case float64, float32:
				if fValue, fErr := strconv.ParseFloat(self.Value, 64); fErr == nil {
					var mValFloat float64
					switch mValTypeValue2 := mValTypeValue.(type) {
						case float64: mValFloat = mValTypeValue2
						case float32: mValFloat = float64(mValTypeValue2)
					}
					switch self.Formula {
						case "=", "is": result = mValFloat == fValue
						case ">": result = mValFloat > fValue
						case "<": result = mValFloat < fValue
						case "≥", ">=": result = mValFloat >= fValue
						case "≤", "<=": result = mValFloat <= fValue
						case "≠", "!=", "not": result = mValFloat != fValue
					}
				}
			case bool:
				if fValue, fErr := strconv.ParseBool(self.Value); fErr == nil {
					switch self.Formula {
						case "=", "is": result = mValTypeValue == fValue
						case "≠", "!=", "not": result = mValTypeValue != fValue
					}
				}
			case string:
				fValue := self.Value
				switch self.Formula {
					case "=", "is": result = mValTypeValue == fValue
					case ">": result = mValTypeValue > fValue
					case "<": result = mValTypeValue < fValue
					case "≥", ">=": result = mValTypeValue >= fValue
					case "≤", "<=": result = mValTypeValue <= fValue
					case "≠", "!=", "not": result = mValTypeValue != fValue
				}
		}
	}
	return result
}


type Condition struct {
	Formula		string					`json:"formula"`
	Value 			string 				`json:"value"`
	formulas		[][]Formula 	
}

func (self *Condition) GetFormulas() [][]Formula {
	if self.formulas == nil {
		self.parseFormula()
	}
	return self.formulas
}

func (self *Condition) parseFormula() [][]Formula {
	f1 := strings.ToLower(self.Formula)					// 转换为小写
	ors_ands := [][]Formula{}

	keyRe, _ := regexp.Compile(KEY_REGEX)
	formulas := []Formula{}

	reIndexsList := keyRe.FindAllStringSubmatchIndex(f1, -1)
	for i, reIndexs := range reIndexsList {
		afterFor, afterBegin := "", len(f1) 
		if i < len(reIndexsList) -1 {
			afterIndexs := reIndexsList[i + 1]
			afterFor, afterBegin = f1[afterIndexs[2]:afterIndexs[3]], afterIndexs[0]
		}

		fKey := f1[reIndexs[4]:reIndexs[5]]
		fFor := strings.Trim(f1[reIndexs[6]:reIndexs[7]], VALUE_TRIM)
		fVal := strings.Trim(f1[reIndexs[1]:afterBegin], VALUE_TRIM)
		formula := Formula{Key: fKey, Formula: fFor, Value: fVal}
		formulas = append(formulas, formula)
		// fmt.Printf("%d key:%s for:%s val:%s after:%s\n", i, fKey, fFor, fVal, afterFor)
		if afterFor != "and" && len(formulas) > 0 {
			ors_ands = append(ors_ands, formulas)
			formulas = []Formula{}
		}
	}
	self.formulas = ors_ands
	return self.formulas
}

func (self *Condition) Calculate(vals map[string]interface{}) bool {
	if self.formulas == nil {
		self.parseFormula()
	}

	result := false
	for _, ands := range self.formulas {
		andsRes := true
		for _, and := range ands {
			if andsRes = and.Calculate(vals); !andsRes {
				break
			}
		}
		if andsRes {
			result = true
			break
		}
	}
	return result
}


// 因子，由默认值和多个表达式组成
type Factor struct {
	Value			string		`json:"value"`
	Conditions		[]Condition	`json:"conditions"`
}

// 查找第一个匹配的表达式的值，否则返回默认值
func (self *Factor) GetValue(vals map[string]interface{}) string {
	if self.Conditions != nil {
		for _, con := range self.Conditions {
			if con.Calculate(vals) {
				return con.Value
			}
		}
	}
	return self.Value
}

// 中转对象
type factor struct {
	Value			string		`json:"value"`
	Conditions		[]Condition	`json:"conditions"`
}

func (self *Factor) GetFormulaKeys() []string {
	keys := []string{}
	for _, con := range self.Conditions {
		for _, formulas := range con.GetFormulas() {
			for _, formula := range formulas {
				keys = append(keys, formula.Key)
			}
		}
	}
	return keys
}

func (self *Factor) UnmarshalJSON(data []byte) error {
	newBs := bytes.Trim(data, " \n\r\t\v\f")
	if bytes.HasPrefix(newBs, []byte("{")) {
		if bytes.Contains(newBs, []byte("\"conditions\"")) {
			fact := &factor{}
			if err := json.Unmarshal(newBs, fact); err == nil {
				if fact.Value != "" || len(fact.Conditions) > 0 {
					self.Value = fact.Value
					self.Conditions = fact.Conditions
					return nil
				}
			}
		}
	}
	self.Value = string(data)
	return nil
}

func NewFactor(data []byte) Factor {
	fact := Factor{}
	if err := json.Unmarshal(data, &fact); err == nil {
		return fact
	} 
	fact.Value = string(data)
	return fact
}
