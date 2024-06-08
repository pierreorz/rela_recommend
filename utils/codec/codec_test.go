package codec

import (
	"fmt"
	"testing"
)

func TestChangeToSimple(t *testing.T) {
	var text = "干s尽拚s脏拼A乾儘臟"
	out := ChangeToSimple(text)
	//fmt.Println(out)
	if out != "干s尽拼s脏拼A干尽脏" {
		t.Error("wrong")
	}
}

func TestChangeToTradition(t *testing.T) {
	var text = "干s尽拚s脏拼A乾儘臟"
	out := ChangeToTradition(text)
	//fmt.Println(out)
	if out != "幹s盡拚s髒拚A乾儘臟" {
		t.Error("wrong")
	}
}

func TestDistinguishLang(t *testing.T) {
	var text = "干s尽拚s脏拼A乾儘臟"
	out := DistinguishLang(text)
	//fmt.Println(out)
	if out != LANGUAGE_CHT {
		t.Error("wrong")
	}

	text = "干尽"
	out = DistinguishLang(text)
	//fmt.Println(out)
	if out != LANGUAGE_CHS {
		t.Error("wrong")
	}

	text = "wdfwefwefwef"
	out = DistinguishLang(text)
	//fmt.Println(out)
	if out != LANGUAGE_EN {
		t.Error("wrong")
	}
}

func TestIsHan(t *testing.T) {
	if !isHan("干你好add乾儘臟fafaf") {
		t.Error("wrong")
	}
}

func TestEncode(t *testing.T) {
	/*
		var text = "1"
		fmt.Println("text:", text)
		fmt.Println(text[0])
		encode := DefaultEncoding.EncodeToString([]byte(text))
		fmt.Println("encode:", encode)
		fmt.Println(len(codecTable), codecTable)
		decode, err := DefaultEncoding.DecodeString(encode)
		if err != nil{
			t.Error(err.Error())
		}
		fmt.Println(string(decode))
		if string(decode) != text {
			t.Error("wrong")
		}
	*/

	var text = "1"
	fmt.Println("text:", text)
	fmt.Println(text[0])
	encode := Encode(text)
	fmt.Println("encode:", encode)
	fmt.Println(len(codecTable), codecTable)
	decode, err := Decode(encode)
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Println(decode)
	if decode != text {
		t.Error("wrong")
	}

}
