package codec

import (
	"encoding/base32"
	"fmt"
	"strings"
)

/**
 * 字符串的(编码/加密)与(解码/解密)。
 *
 * 规则：
 *
 * 在Base64中，码表是由[A-Z,a-z,0-9,+,/,=(pad)]组成的。 而在这里，码表由[a-z,2-7]组成的：
 * ----------------------------------------------- a b c d e f g h i j k l m n o
 * p q r 0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17
 * ----------------------------------------------- s t u v w x y z 2 3 4 5 6 7
 * 18 19 20 21 22 23 24 25 26 27 28 29 30 31
 * ------------------------------------------------
 *
 * 在Base64中，是将二进制连成一串，然后再按6位来分割，分割完后在前面补0，这个地球人都知道，不多说了。
 * 而在这里，在分割的那一步稍微有变动，是按5位来分割，如果刚好够分，那就好了，如果不够，那咋办呢？
 *
 * 在Base64中，是用"="来解决的吧。 而在这里，就是在前面补0，然后在后面再补零。
 *
 * 例如：字符串 "aaa"，(编码/加密)后是 "mfqwc"
 *
 * 二进制：01100001 01100001 01100001 转换后：(000)01100 (000)00101 (000)10000
 * (000)10110 (000)0001(0) 十进制： 12 5 16 22 2 码表对应： m f q w c
 *
 * (解码/解密)就更简单了：
 *
 * 码表对应： m f q w c 十进制： 12 5 16 22 2 二进制： 00001100 00000101 00010000 00010110
 * 00000010 去前0后：01100 00101 10000 10110 00010 合并后： 0110000101100001011000010
 *
 * 然后把合并后的串的长度除一下8，发现多了个0：
 *
 * 二进制：01100001 01100001 01100001 0
 *
 * 多了就算了，不要了（其实是在{编码/加密}的分割时候，在分剩的余数的后面补的0）。 然后再将 byte[] 转回字符串，OK！又见"aaa"了。
 *
 * 有一点值得注意的，UTF-8、GBK、GB18030 一般都没什么问题， 但是 GB2312 可能字符集不够丰富，繁体字在decode的时候成问号了。
 *
 * @version 2008-12-3 下午03:01:50
 */

const (
	/**
	 * 码表
	 */
	codecTable = "abcdefghijklmnopqrstuvwxyz234567"
)

// StdEncoding is the standard base32 encoding, as defined in
// RFC 4648.
var DefaultEncoding = base32.NewEncoding(codecTable)

/**
 * (编码/加密)字符串，采用 UTF-8 的 character set。
 *
 * @param keys
 *            需要(编码/加密)的字符串
 *	      Base32编码后把 = 去掉
 * @return (编码/加密)后的字符串
 */
func Encode(text string) string {
	return strings.TrimRight(DefaultEncoding.EncodeToString([]byte(text)), "=")
}

/**
 * (解码/解密)字符串，采用 UTF-8 的 character set。
 *
 * @param code
 *            需要(解码/解密)的字符串
 *
 * @return (解码/解密)后的字符串
 */
func Decode(text string) (string, error) {
	//因为Base32是把5个字节变为8个字节，所以，Base32编码的长度永远是8的倍数，因此，需要加上=把Base32字符串的长度变为8的倍数，就可以正常解码了。
	count := len(text) % 8
	for i := 0; i < 8-count; i++ {
		text += "="
	}
	fmt.Println(text)
	data, err := DefaultEncoding.DecodeString(text)
	return string(data), err
}
