package utils

// onehot
type OneHotEncoder struct {
	Categories []map[int]int `json:"categories"`
}

func (self *OneHotEncoder) Transform(features []int) *Features {
	new_features := &Features{}
	for i := 0; i < len(features); i++ {
		index, ok := self.Categories[i][features[i]]
		if ok {
			new_features.Add(index, 1.0)
		}
	}
	return new_features
}

// feature map
type FeaturesMapEncoder struct {
	Maps map[int]int `json:"maps"`
}

func (self *FeaturesMapEncoder) Transform(features *Features) *Features {
	if self.Maps == nil || len(self.Maps) == 0 {
		return features
	}

	new_features := &Features{}
	for key, val := range features.ToMap() {
		if new_key, ok := self.Maps[key]; ok {
			new_features.Add(new_key, val)
		}
	}
	return new_features
}

// feature transformer
type featuresFieldSummary struct {
	Min        float32   `json:"min"`
	Max        float32   `json:"max"`
	Avg        float32   `json:"avg"`
	Std        float32   `json:"std"`
	Percentile []float32 `json:"percentile"`
}

type featuresFieldTransformer struct {
	Start   int                   `json:"start"`
	Length  int                   `json:"length"`
	Type    int                   `json:"type"`
	Name    string                `json:"name"`
	Summary *featuresFieldSummary `json:"summary"`
}

// 特征转化处理
type FeaturesTransformer struct {
	StartIndex  int                        `json:"start_index"`
	FeatureSize int                        `json:"feature_size"`
	FieldSize   int                        `json:"field_size"`
	Fields      []featuresFieldTransformer `json:"fields"`
}

func (self *FeaturesTransformer) Transform(features *Features) *Features {
	if self.Fields == nil || len(self.Fields) == 0 {
		return features
	}

	newFeatures := &Features{}
	currentIndex := self.StartIndex
	for _, field := range self.Fields {
		switch field.Type {
		case 0: // 数值类型直接拿取
			for i := field.Start; i < field.Start+field.Length; i++ {
				val, _ := features.Get(i)
				newFeatures.Add(currentIndex, val)
				currentIndex++
			}
		case 1:
			oneHotIndex, oneHotValue := currentIndex, float32(0.0)
			for i := field.Start; i < field.Start+field.Length; i++ {
				if val, ok := features.Get(i); ok {
					oneHotIndex = currentIndex
					oneHotValue = val
				}
				currentIndex++
			}
			newFeatures.Add(oneHotIndex, oneHotValue)
		case 2: // mutil-onehot 类型直接
			for i := field.Start; i < field.Start+field.Length; i++ {
				val, _ := features.Get(i)
				newFeatures.Add(currentIndex, val)
				currentIndex++
			}
		case 3: // bulk 分桶类型
			bulkIndex := currentIndex
			val, _ := features.Get(field.Start)
			currentIndex++
			for _, spliter := range field.Summary.Percentile {
				if val >= spliter {
					bulkIndex = currentIndex
				}
				currentIndex++
			}
			newFeatures.Add(bulkIndex, 1.0)
		case 4: // z-score 标准化类型
			for i := field.Start; i < field.Start+field.Length; i++ {
				val, _ := features.Get(i)
				newVal := (val - field.Summary.Avg) / field.Summary.Std
				newFeatures.Add(currentIndex, newVal)
				currentIndex++
			}
		case 5: // min-max 标准化类型
			for i := field.Start; i < field.Start+field.Length; i++ {
				val, _ := features.Get(i)
				newVal := (val - field.Summary.Min) / (field.Summary.Max - field.Summary.Min)
				newFeatures.Add(currentIndex, newVal)
				currentIndex++
			}
		}
	}
	return newFeatures
}
