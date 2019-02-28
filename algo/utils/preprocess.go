package utils


// onehot
type OneHotEncoder struct {
	Categories []map[int]int 		`json:"categories"`
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
	Maps map[int]int				`json:"maps"`
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
