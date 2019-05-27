package segment

import (
	"os"
	"rela_recommend/log"
	"path"
	"io"
	"bufio"
	"strings"
	"github.com/go-ego/gse"
)


type ISegmenter interface {
	LoadDict(...string) error
	FilterStopWords([]string) []string
	Cut(string) []string
}

type BaseSegmenter struct {
	DataPath	string
	StopWords	map[string]bool
}

func (self *BaseSegmenter) ReadLines(file string) ([]string, error) {
	res := make([]string, 0)
	dictFile, err := os.Open(file)
	if err != nil {
		log.Warnf("Could not read file lines: \"%s\", %v \n", file, err)
		return nil, err
	}
	defer dictFile.Close()
	buf := bufio.NewReader(dictFile)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		res = append(res, line)
		if err != nil {
			if err == io.EOF {
				return res, nil
			}
			return res, err
		}
	}
}

func (self *BaseSegmenter) LoadStopWords(files ...string) error {
	if self.StopWords == nil {
		self.StopWords = make(map[string]bool, 0)
	}
	for _, file := range files {
		realFile := path.Join(self.DataPath, file)
		
		lines, err := self.ReadLines(realFile)
		if err == nil {
			for _, line := range lines {
				self.StopWords[line] = true
			}
		} else {
			log.Warnf("Could not load stop words: \"%s\", %v \n", file, err)
		}
		log.Infof("loading stop words: %s , lines %d\n", realFile, len(lines))
	}
	return nil
}

func (self *BaseSegmenter) FilterStopWords(words []string) []string {
	res := make([]string, 0)
	for _, word := range words {
		if _, ok := self.StopWords[word]; !ok {
			res = append(res, word)
		}
	}
	return res
}


type Segmenter struct {
	BaseSegmenter
	seg			*gse.Segmenter
}

func (self *Segmenter) LoadDict(files ...string) error {
	self.seg.LoadDict()
	for _, file := range files {
		realFile := path.Join(self.DataPath, file)
		err := self.seg.Read(realFile)
		if err != nil {
			return err
		}
	}
	self.seg.CalcToken()
	return nil
}

func (self *Segmenter) Cut(str string) []string {
	words := self.seg.Cut(str, true)
	return self.FilterStopWords(words)
}

func NewSegmenter() ISegmenter {
	work_dir, _ := os.Getwd()
	dictFiles := []string {"dict/dictionary.txt", "dict/zh/dict.txt", "stop_words.txt", "user_dict.txt"}
	stopWordsFile := []string {"stop_words.txt"}
	seg := &Segmenter{BaseSegmenter:BaseSegmenter{DataPath: work_dir + "/algo_files/segment/"}, seg: &gse.Segmenter{}}
	seg.LoadDict(dictFiles...)
	seg.LoadStopWords(stopWordsFile...)
	return seg
}
