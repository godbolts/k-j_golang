package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func inputRead(route string) []string {
	rawFile, _ := os.ReadFile(route)
	stringFile := string(rawFile)

	expression := regexp.MustCompile(` {2,}`)
	processedFile := expression.ReplaceAllString(stringFile, " ")

	expression = regexp.MustCompile(`[\v\f\r]+`)
	processedFile = expression.ReplaceAllString(processedFile, "\n")

	expression = regexp.MustCompile(`\n{3,}`)
	processedFile = expression.ReplaceAllString(processedFile, "\n\n")

	lines := strings.Split(processedFile, "\n")

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

func airportRead(route string, condition string) string {
	var unchanged string
	city := false

	if condition[0] == '*' {
		city = true
		condition = condition[1:]
		unchanged = unchanged + "*"
	}

	if condition[0:2] == "##" {
		if len(condition) < 6 {
			return condition
		}
		condition = condition[2:6]
		unchanged = unchanged + "##"
	} else if condition[0:1] == "#" {
		condition = condition[1:4]
		unchanged = unchanged + "#"
	}

	unchanged = unchanged + condition
	file, _ := os.Open(route)

	defer file.Close()
	reader := csv.NewReader(bufio.NewReader(file))
	headers, _ := reader.Read()
	data := make(map[string]string)

	for {
		record, err := reader.Read()

		if err != nil {
			if err == io.EOF {
				break
			}
		}

		for i := 0; i < len(headers) && i < len(record); i++ {
			data[headers[i]] = record[i]

			if strings.Contains(record[i], condition) && len(record[i]) == len(condition) {
				airName := data["name"]
				if city {
					airName = data["municipality"]
				}
				return airName
			}
		}
	}
	return unchanged
}

func formatTime(raw string) (string, error) {
	var formated string

	dPattern := regexp.MustCompile(`^D\((\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(Z|[-+]\d{2}:\d{2}))\)[.,!?]?`)
	tPattern := regexp.MustCompile(`^(T12|T24)\((\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(Z|[-+]\d{2}:\d{2}))\)[.,!?]?`)

	if match := tPattern.FindStringSubmatch(raw); match == nil {
		if match := dPattern.FindStringSubmatch(raw); match == nil {
			return raw, errors.New(" ")
		}
	}

	if raw[0:2] == "D(" {
		firstSplit := strings.Split(raw, "T")
		secondSplit := strings.Split(firstSplit[0], "-")
		splitDay := secondSplit[2]
		splitYear := secondSplit[0][2:]
		numbMonth, _ := strconv.Atoi(secondSplit[1])

		monthTime := time.Date(0, time.Month(numbMonth), 1, 0, 0, 0, 0, time.UTC)
		month := monthTime.Format("Jan")
		formated = splitDay + " " + month + " " + splitYear

	} else {
		var times []string
		var currentTime string
		var movement string

		add := false
		neg := false
		direction := "+"
		dateTime := strings.Split(raw, "T")

		if raw[0:4] == "T12(" {
			add = true
		}

		if strings.Contains(dateTime[2], "+") {
			times = strings.Split(dateTime[2], "+")
			currentTime = times[0][:5]
			movement = times[1][:5]
		} else if strings.Contains(dateTime[2], "-") {
			times = strings.Split(dateTime[2], "-")
			currentTime = times[0][:5]
			movement = times[1][:5]
			neg = true
		} else {
			currentTime = dateTime[2][0:5]
			movement = "00:00"
		}

		if neg {
			direction = "-"
		}
		hour, _ := strconv.Atoi(currentTime[0:2])

		if add {
			affix := "AM ("
			if hour > 12 {
				newHour := strconv.Itoa(hour - 12)
				currentTime = newHour + currentTime[2:]
				affix = "PM ("
			}
			formated = currentTime + affix + direction + movement + ")"
		} else {
			formated = currentTime + " (" + direction + movement + ")"
		}
	}
	return formated, nil
}

func outputWrite(route, processed string) {

	file, _ := os.Create(route)
	defer file.Close()
	_, _ = file.WriteString(processed)

}

func processText(raw []string, airportPath string) string {
	var processed []string
	for i := 0; i < len(raw); i++ {
		parts := strings.Split(raw[i], " ")

		for i := 0; i < len(parts); i++ {
			var punct string
			re := regexp.MustCompile(`^(D\(.*?|T12\(.*?|T24\(.*?)`)

			if match := re.FindStringSubmatch(parts[i]); match != nil {
				if strings.ContainsAny(parts[i][len(parts[i])-1:], ",.!?") {
					punct = parts[i][len(parts[i])-1:]
				}
				newTime, err := formatTime(parts[i])
				if err != nil {
					continue
				}
				parts[i] = newTime + punct
			}
			if strings.Contains(parts[i], "#") {
				if strings.ContainsAny(parts[i][len(parts[i])-1:], ",.!?") {
					punct = parts[i][len(parts[i])-1:]
				}
				airportName := airportRead(airportPath, parts[i])
				parts[i] = airportName + punct
			}
		}
		raw[i] = strings.Join(parts, " ")
		processed = append(processed, raw[i])
	}

	finalizing := strings.Join(processed, "\n")
	return finalizing
}

func Validation(inRoute, airRoute string) error {
	red := "\033[31m"
	reset := "\033[0m"
	_, err := os.ReadFile(inRoute)
	if err != nil {
		fmt.Println(red + "Input not found." + reset)
		return err
	}

	airports, err := os.Open(airRoute)
	if err != nil {
		fmt.Println(red + "Airport lookup not found." + reset)
		return err
	}
	defer airports.Close()

	reader := csv.NewReader(bufio.NewReader(airports))
	header, _ := reader.Read()

	if len(header) != 6 {
		fmt.Println(red + "Airport lookup malformed." + reset)
		return errors.New(" ")
	}

	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		for _, cell := range row {
			if cell == "" {
				fmt.Println(red + "Airport lookup malformed." + reset)
				return errors.New(" ")
			}
		}
	}
	return nil
}

func outputDisplay(text string) {
	offsetPattern := regexp.MustCompile(`[\+\-]\d{2}:\d{2}`)
	matches := offsetPattern.FindAllString(text, -1)

	for _, match := range matches {
		escapedMatch := "\033[1m" + match + "\033[0m"
		quotedMatch := regexp.QuoteMeta(match)
		text = regexp.MustCompile(quotedMatch).ReplaceAllString(text, escapedMatch)
	}

	fmt.Println(text)
}

func main() {
	var helpFlag bool
	var displayFlag bool

	flag.BoolVar(&helpFlag, "h", false, "Display usage.")
	flag.BoolVar(&displayFlag, "d", false, "Display the output.")
	flag.Parse()

	if helpFlag || flag.NArg() < 3 {
		fmt.Println("Usage:\n go run . ./input.txt ./output.txt ./airport-lookup.csv")
		return
	}

	inputPath := flag.Arg(0)[2:]
	outputPath := flag.Arg(1)[2:]
	airportPath := flag.Arg(2)[2:]

	err := Validation(inputPath, airportPath)
	if err != nil {
		return
	}

	rawText := inputRead(inputPath)

	convertedText := processText(rawText, airportPath)

	outputWrite(outputPath, convertedText)

	if displayFlag {
		outputDisplay(convertedText)
	}
}
