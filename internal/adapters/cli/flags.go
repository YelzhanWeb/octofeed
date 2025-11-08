package cli

import (
	"fmt"
	"strconv"
	"time"
)

func parseAddFlags(args []string) (name, url string, err error) {
	if len(args) < 4 {
		return "", "", fmt.Errorf("usage: rsshub add --name <name> --url <url>")
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--name":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--name requires a value")
			}
			name = args[i+1]
			i++
		case "--url":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--url requires a value")
			}
			url = args[i+1]
			i++
		}
	}

	if name == "" {
		return "", "", fmt.Errorf("--name is required")
	}
	if url == "" {
		return "", "", fmt.Errorf("--url is required")
	}

	return name, url, nil
}

func parseSetIntervalFlags(args []string) (time.Duration, error) {
	if len(args) < 2 {
		return 0, fmt.Errorf("usage: rsshub set-interval --duration <duration>")
	}

	var durationStr string
	for i := 0; i < len(args); i++ {
		if args[i] == "--duration" {
			if i+1 >= len(args) {
				return 0, fmt.Errorf("--duration requires a value")
			}
			durationStr = args[i+1]
			break
		}
	}

	if durationStr == "" {
		return 0, fmt.Errorf("--duration is required")
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: %w", err)
	}

	return duration, nil
}

func parseSetWorkersFlags(args []string) (int, error) {
	if len(args) < 2 {
		return 0, fmt.Errorf("usage: rsshub set-workers --count <count>")
	}

	var countStr string
	for i := 0; i < len(args); i++ {
		if args[i] == "--count" {
			if i+1 >= len(args) {
				return 0, fmt.Errorf("--count requires a value")
			}
			countStr = args[i+1]
			break
		}
	}

	if countStr == "" {
		return 0, fmt.Errorf("--count is required")
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("invalid count format: %w", err)
	}

	if count <= 0 {
		return 0, fmt.Errorf("count must be greater than 0")
	}

	return count, nil
}

func parseListFlags(args []string) (int, error) {
	if len(args) == 0 {
		return 0, nil
	}

	for i := 0; i < len(args); i++ {
		if args[i] == "--num" {
			if i+1 >= len(args) {
				return 0, fmt.Errorf("--num requires a value")
			}
			num, err := strconv.Atoi(args[i+1])
			if err != nil {
				return 0, fmt.Errorf("invalid num format: %w", err)
			}
			return num, nil
		}
	}

	return 0, nil
}

func parseDeleteFlags(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("usage: rsshub delete --name <name>")
	}

	for i := 0; i < len(args); i++ {
		if args[i] == "--name" {
			if i+1 >= len(args) {
				return "", fmt.Errorf("--name requires a value")
			}
			return args[i+1], nil
		}
	}

	return "", fmt.Errorf("--name is required")
}

func parseArticlesFlags(args []string) (feedName string, num int, err error) {
	if len(args) < 2 {
		return "", 0, fmt.Errorf("usage: rsshub articles --feed-name <name> [--num <count>]")
	}

	num = 3 // default

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--feed-name":
			if i+1 >= len(args) {
				return "", 0, fmt.Errorf("--feed-name requires a value")
			}
			feedName = args[i+1]
			i++
		case "--num":
			if i+1 >= len(args) {
				return "", 0, fmt.Errorf("--num requires a value")
			}
			parsedNum, err := strconv.Atoi(args[i+1])
			if err != nil {
				return "", 0, fmt.Errorf("invalid num format: %w", err)
			}
			num = parsedNum
			i++
		}
	}

	if feedName == "" {
		return "", 0, fmt.Errorf("--feed-name is required")
	}

	return feedName, num, nil
}
