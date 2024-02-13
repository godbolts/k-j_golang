package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func slicer(wholeText string) []string {
	var cashe string
	var slicedText []string
	var result []string

	for i := 0; i < len(wholeText); i++ {
		if string(wholeText[i]) == "[" {
			slicedText = append(slicedText, cashe)
			cashe = ""
			cashe += string(wholeText[i])
		}
		if string(wholeText[i]) == "]" {
			cashe += string(wholeText[i])
			slicedText = append(slicedText, cashe)
			cashe = ""
		}
		if string(wholeText[i]) != "[" && string(wholeText[i]) != "]" {
			cashe += string(wholeText[i])
			if i+1 == len(wholeText) {
				slicedText = append(slicedText, cashe)
			}
		}
	}

	for i := 0; i < len(slicedText); i++ {
		if string(slicedText[i]) != "" {
			result = append(result, slicedText[i])
		}
	}
	return result
}

func getNumbers(encoding string) (int, int, error) {
	var number string

	for i := 1; i < len(encoding)-1; i++ {
		number = number + string(encoding[i])
		if !unicode.IsDigit(rune(encoding[i+1])) {
			break
		}
	}

	lenght := len(number)
	multiplyer, err := strconv.Atoi(number)
	return multiplyer, lenght, err
}

func capture(encodedText []string) (string, error) {
	pattern := regexp.MustCompile(`^\[.*?\]$`)

	for i := 0; i < len(encodedText); i++ {
		if match := pattern.FindStringSubmatch(encodedText[i]); match != nil {
			numberTimes, numberLenght, err := getNumbers(encodedText[i])
			if err != nil {
				return "", err
			}
			if encodedText[i][numberLenght+1] != ' ' {
				return "", errors.New("no space")
			}
			symbol := encodedText[i][numberLenght+2 : len(encodedText[i])-1]
			if strings.ContainsAny(symbol, "[]") {
				return "", errors.New("no symbol")
			}
			newMatch := strings.Repeat(symbol, numberTimes)
			encodedText[i] = newMatch
		}
	}
	dencodedText := strings.Join(encodedText, "")
	return dencodedText, nil
}

func input() string {
	var slicedText []string
	var result []string
	reader := bufio.NewReader(os.Stdin)

	for {
		line, _ := reader.ReadString('\n')
		slicedText = append(slicedText, line)
		if strings.Count(strings.Join(slicedText, ""), "\n\n") == 1 {
			break
		}
	}
	for i := 0; i < len(slicedText); i++ {
		if string(slicedText[i]) != "" {
			result = append(result, slicedText[i])
		}
	}
	encodedText := strings.Join(result, "")
	for {
		if encodedText[len(encodedText)-1] == '\n' {
			encodedText = encodedText[:len(encodedText)-1]
		} else {
			break
		}
	}
	return encodedText
}

func validity(inputStream string) error {
	var open int
	var closed int

	for i := 0; i < len(inputStream); i++ {
		if inputStream[i] == '[' {
			open++
		} else if inputStream[i] == ']' {
			closed++
		}
	}
	if open == closed {
		return nil
	} else {
		return errors.New("unmatched brackets")
	}
}

func numberSymbol(input string) string {
	number := len(input)
	symbol := input[0]
	return "[" + strconv.Itoa(number) + " " + string(symbol) + "]"
}
func doubleSymbol(input string) string {
	number := len(input) / 2
	symbol := input[0:2]
	return "[" + strconv.Itoa(number) + " " + string(symbol) + "]"
}

func reEncoderSingle(decodedText string) string {
	var cache string
	var sliced []string

	for i := 0; i < len(decodedText); i++ {
		if len(cache) == 0 || cache[0] == decodedText[i] {
			cache = cache + string(decodedText[i])
		} else if cache[0] != decodedText[i] && len(cache) > 0 {
			if len(cache) > 1 {
				sliced = append(sliced, numberSymbol(cache))
				cache = ""
				cache = cache + string(decodedText[i])
			} else if len(cache) == 1 {
				sliced = append(sliced, cache)
				cache = ""
				cache = cache + string(decodedText[i])
			}
		}
		if i == len(decodedText)-1 && cache[0] == decodedText[i] && len(cache) > 1 {
			sliced = append(sliced, numberSymbol(cache))
		} else if i == len(decodedText)-1 {
			sliced = append(sliced, cache)
		}
	}
	result := strings.Join(sliced, "")
	return result
}

func reEncoderDouble(decodedText string) string {
	var cache string
	var sliced []string

	for j := 0; j <= len(decodedText); j += 2 {
		if len(decodedText)%2 == 0 {
			if j == len(decodedText) {
				if len(cache) > 2 {
					cache = doubleSymbol(cache)
				}
				sliced = append(sliced, cache)
				break
			}
		} else {
			if j+1 == len(decodedText) {
				if len(cache) > 2 {
					cache = doubleSymbol(cache)
				}
				cache += string(decodedText[j])
				sliced = append(sliced, cache)
				break
			}
		}
		if len(cache) < 2 {
			cache += string(decodedText[j]) + string(decodedText[j+1])
		} else if (string(cache[0]) + string(cache[1])) == (string(decodedText[j]) + string(decodedText[j+1])) {
			cache += string(decodedText[j]) + string(decodedText[j+1])
		} else if (string(cache[0])+string(cache[1])) != (string(decodedText[j])+string(decodedText[j+1])) && len(cache) > 0 {
			if len(cache) > 2 {
				sliced = append(sliced, doubleSymbol(cache))
				cache = string(decodedText[j]) + string(decodedText[j+1])
			} else if len(cache) == 2 {
				sliced = append(sliced, cache)
				cache = string(decodedText[j]) + string(decodedText[j+1])
			}
		}
	}
	result := strings.Join(sliced, "")
	return result
}

func main() {
	var encodedText string
	var multiLine bool
	var encoder bool

	flag.BoolVar(&multiLine, "multi", false, "Multiline art")
	flag.BoolVar(&encoder, "encode", false, "Encoder")
	flag.Parse()

	if multiLine {
		encodedText = input()
	} else {
		encodedText = flag.Arg(0)
	}
	if encoder {

		reencodedText := reEncoderDouble(reEncoderSingle(encodedText))
		fmt.Println(reencodedText)
	} else {
		err := validity(encodedText)
		if err != nil {
			fmt.Println("Error")
			return
		}
		slicedText := slicer(encodedText)
		decodedText, err := capture(slicedText)
		if err != nil {
			fmt.Println("Error")
			return
		}
		fmt.Println(decodedText)
	}
}
