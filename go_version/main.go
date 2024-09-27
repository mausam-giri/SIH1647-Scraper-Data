package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "sync"
    "strings"
)

var baseURL = "https://dca.ceda.ashoka.edu.in/index.php/home/"
var endpoint = "getcsv"

// Sample data; replace with actual data.
var sellers = map[string]string{
    "Retail": "1",
}
var commodities = map[string]string{
    "Atta (Wheat)": "2",
}
var centres = map[string]string{
    "Delhi": "3",
}
var years = map[string]string{
    "2021": "4",
}

func updateParams(commodityName, centreName, sellerType, year string) string {
    return fmt.Sprintf("c=[%%22%s%%22]&s=%s&t=%s&y=%s",
        commodities[commodityName], centres[centreName], sellers[sellerType], years[year])
}

func getURL(params string) string {
    return fmt.Sprintf("%s%s?%s", baseURL, endpoint, params)
}

func downloader(filename, url string, wg *sync.WaitGroup) {
    defer wg.Done()

    client := &http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        fmt.Printf("Failed to create request: %v\n", err)
        return
    }
    req.Header.Set("User-Agent", "Mozilla/5.0")
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("%s, An error occurred: %v\n", url, err)
        appendToFailed(url)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode == 200 {
        os.MkdirAll(filepath.Dir(filename), os.ModePerm)

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Printf("Failed to read response body: %v\n", err)
            return
        }

        err = ioutil.WriteFile(fmt.Sprintf("%s.csv", filename), body, 0644)
        if err != nil {
            fmt.Printf("Failed to write file: %v\n", err)
            return
        }
        fmt.Printf("Downloaded %s\n", filename)
    }
}

func appendToFailed(url string) {
    file, err := os.OpenFile("failed.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Printf("Could not open failed.txt: %v\n", err)
        return
    }
    defer file.Close()
    file.WriteString(url + "\n")
}

func writeCSV(filename, url string) {
    endpoint := strings.Split(url, "?")[1]
    file, err := os.OpenFile("DownloadLinks.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Printf("Could not open DownloadLinks.csv: %v\n", err)
        return
    }
    defer file.Close()
    file.WriteString(fmt.Sprintf("%s,%s\n", filename, endpoint))
}

func downloadPulsesWise() {
    var wg sync.WaitGroup
    maxGoroutines := 10
    sem := make(chan struct{}, maxGoroutines) // Semaphore to limit concurrent downloads

    for sellerType := range sellers {
        for commodityType := range commodities {
            for centreType := range centres {
                for yearType := range years {
                    if !isExcludedCentre(centreType) {
                        params := updateParams(commodityType, centreType, sellerType, yearType)
                        url := getURL(params)
                        filename := fmt.Sprintf("./%s/%s/%s/%s", strings.Title(sellerType), commodityType, centreType, yearType)
                        writeCSV(filename, url)

                        wg.Add(1)
                        sem <- struct{}{} // Acquire a token
                        go func(filename, url string) {
                            defer wg.Done()
                            downloader(filename, url)
                            <-sem // Release a token
                        }(filename, url)

                        fmt.Printf("[+] [%s] %s - %s %s\n", strings.Title(sellerType), commodityType, centreType, yearType)
                    }
                }
            }
        }
    }
    wg.Wait()
}

func isExcludedCentre(centre string) bool {
    excluded := []string{"Adilabad", "Agar Malwa", "Agartala"}
    for _, ex := range excluded {
        if centre == ex {
            return true
        }
    }
    return false
}

func main() {
    fmt.Println("Starting: Scrapper")
    downloadPulsesWise()
}
