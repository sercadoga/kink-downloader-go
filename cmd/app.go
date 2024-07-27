package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/chromedp/chromedp"
)

// CaptureCookies captures cookies from a page using chromedp
func CaptureCookies(targetURL string) ([]*http.Cookie, error) {
	// Create a new context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new browser instance
	c, err := chromedp.New(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Shutdown(ctx)

	// Store cookies
	var cookies []*http.Cookie

	// Navigate to the page
	err = chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Retrieve cookies
			var result []struct {
				Name     string `json:"name"`
				Value    string `json:"value"`
				Domain   string `json:"domain"`
				Path     string `json:"path"`
				Secure   bool   `json:"secure"`
				HttpOnly bool   `json:"httponly"`
			}
			err := chromedp.Evaluate(`JSON.stringify(document.cookie.split('; ').map(c => {
                let [name, value] = c.split('=');
                return {
                    name: name,
                    value: value,
                    domain: document.domain,
                    path: '/',
                    secure: false,
                    httponly: false
                };
            }))`, &result).Do(ctx)
			if err != nil {
				return err
			}

			for _, r := range result {
				cookies = append(cookies, &http.Cookie{
					Name:     r.Name,
					Value:    r.Value,
					Domain:   r.Domain,
					Path:     r.Path,
					Secure:   r.Secure,
					HttpOnly: r.HttpOnly,
				})
			}

			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return cookies, nil
}

func main() {
	// Define the URL
	targetURL := "https://www.kink.com/shoot/105688" // Replace with the target URL

	// Capture cookies from the page
	cookies, err := CaptureCookies(targetURL)
	if err != nil {
		log.Fatalf("Failed to capture cookies: %v", err)
	}

	// Create a new cookie jar
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	// Create a URL object for the domain
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse URL: %v", err)
	}

	// Set cookies for the domain
	jar.SetCookies(parsedURL, cookies)

	// Make a request
	resp, err := client.Get(targetURL)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// Print response status
	fmt.Println("Response Status:", resp.Status)
}
