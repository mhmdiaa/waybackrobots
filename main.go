package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/schollz/progressbar/v3"
)

func main() {
	versionsLimit := flag.Int("limit", 50, "limit the number crawled snapshots. Use -1 for unlimited")
	recent := flag.Bool("recent", false, "use the most recent snapshots without evenly distributing them")
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		url, err := cleanURL(scanner.Text())
		if err != nil {
			continue
		}

		versions, err := GetRobotsTxtVersions(url, *versionsLimit, *recent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting versions: %v\n", err)
			os.Exit(1)
		}

		numThreads := 10
		jobCh := make(chan string, numThreads)
		pathCh := make(chan []string)

		progressbarMessage := fmt.Sprintf("Enumerating %s/robots.txt versions...", url)
		bar := progressbar.Default(int64(len(versions)), progressbarMessage)

		var wg sync.WaitGroup
		wg.Add(numThreads)

		for i := 0; i < numThreads; i++ {
			go func() {
				defer wg.Done()
				for version := range jobCh {
					GetRobotsTxtPaths(version, url, pathCh, bar)
				}
			}()
		}

		go func() {
			for _, version := range versions {
				jobCh <- version
			}
			close(jobCh)
		}()

		go func() {
			wg.Wait()
			close(pathCh)
		}()

		allPaths := make(map[string]bool)
		for pathsBatch := range pathCh {
			for _, path := range pathsBatch {
				allPaths[path] = true
			}
		}

		for path := range allPaths {
			fmt.Println(path)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading URLs from stdin: %v\n", err)
		os.Exit(1)
	}
}

func GetRobotsTxtVersions(url string, limit int, recent bool) ([]string, error) {
	requestURL := fmt.Sprintf("https://web.archive.org/cdx/search/cdx?url=%s/robots.txt&output=json&fl=timestamp&filter=statuscode:200&collapse=digest", url)
	if limit != -1 && recent {
		requestURL += "&limit=-" + strconv.Itoa(limit)
	}

	res, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}

	raw, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	var versions [][]string
	err = json.Unmarshal(raw, &versions)
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return []string{}, nil
	}

	versions = versions[1:]

	selectedVersions := make([]string, 0)
	length := len(versions)

	if recent || limit == -1 || length <= limit {
		for _, version := range versions {
			selectedVersions = append(selectedVersions, version...)
		}
	} else {
		interval := (length + limit - 1) / limit

		for i := 0; i < limit; i++ {
			index := length - 1 - (i * interval)
			if index >= length {
				index = length - (limit - i)
			}
			selectedVersions = append(selectedVersions, versions[index]...)
		}
	}
	return selectedVersions, nil
}

func GetRobotsTxtPaths(version string, url string, pathCh chan []string, bar *progressbar.ProgressBar) {
	requestURL := fmt.Sprintf("https://web.archive.org/web/%sif_/%s/robots.txt", version, url)
	res, err := http.Get(requestURL)
	bar.Add(1)
	if err != nil || res.StatusCode != 200 {
		return
	}

	outputURLs := make([]string, 0)

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Disallow:") || strings.HasPrefix(line, "Allow:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			path := strings.TrimSpace(fields[1])
			if path != "" {
				fullURL, err := mergeURLPath(url, path)
				if err != nil {
					continue
				}
				outputURLs = append(outputURLs, fullURL)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return
	}
	pathCh <- outputURLs
}

func mergeURLPath(baseURL, path string) (string, error) {
	host, err := cleanURL(baseURL)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	url := fmt.Sprintf(host + path)
	return url, nil
}

func cleanURL(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	if u.Scheme == "" {
		u.Scheme = "https"
		u.Host = baseURL
	}

	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}
