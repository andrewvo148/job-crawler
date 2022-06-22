package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
		colly.AllowedDomains("itviec.com"),
	)

	// On every a element which has href attribute call callback
	// c.OnHTML("a.job__skill", func(e *colly.HTMLElement) {
	// 	fmt.Println(e.ChildText("span"))
	// 	// link := e.Attr("href")
	// 	// // Print link
	// 	// fmt.Printf("Link found: %q -> %s\n", e.Text, link)
	// 	// Visit link found on page
	// 	// Only those links are visited which are in AllowedDomains
	// 	// c.Visit(e.Request.AbsoluteURL(link))
	// })

	c.OnHTML("div.details > h3.title.job-details-link-wrapper > a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Print link
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		c.Visit(e.Request.AbsoluteURL(link))
	})

	c.OnHTML("div.jd-page__job-details div.job-details", func(e *colly.HTMLElement) {
		fmt.Printf("\n\n")
		fmt.Println("-------JOB DETAILS--------")
		fmt.Println("---------------------------")
		fmt.Printf("Title: %s\n", e.ChildText("h1.job-details__title"))
		fmt.Printf("Tags: %s\n", strings.Trim(e.ChildText("div.job-details__tag-list span"), "\t\n\r"))

		fmt.Printf("Location: %s\n", e.ChildText("div.job-details__overview > div.svg-icon > div.svg-icon__text > span"))

		fmt.Printf("%s: \n", e.ChildText("h2.job-details__second-title:nth-of-type(1)"))
		topReasonHtml, _ := e.DOM.Find("div.job-details__top-reason-to-join-us").Html()
		fmt.Printf("%s\n", topReasonHtml)

		fmt.Printf("%s\n", e.ChildText("h2.job-details__second-title:nth-of-type(2)"))
		paragraphDom := e.DOM.Find(".job-details__paragraph")
		jobDescriptionHtml, _ := paragraphDom.Eq(0).Html()
		fmt.Printf("%s\n", jobDescriptionHtml)

		fmt.Printf("%s\n", e.ChildText("h2.job-details__second-title:nth-of-type(3)"))
		skillAndExperienceHtml, _ := paragraphDom.Eq(1).Html()
		fmt.Printf("%s\n", skillAndExperienceHtml)

		fmt.Printf("%s\n", e.ChildText("h2.job-details__second-title:nth-of-type(4)"))
		loveWorkingHtml, _ := paragraphDom.Eq(2).Html()
		fmt.Printf("%s\n", loveWorkingHtml)
		// fmt.Printf("Title: %s\n", e.Text)
		// fmt.Printf("Title: %s\n", e.Text)
		// fmt.Printf("Title: %s\n", e.Text)
		// fmt.Printf("Title: %s\n", e.Text)
		fmt.Println("---------------------------")
		fmt.Println("---------------------------")
		fmt.Printf("\n\n")
	})

	c.OnHTML("ul.pagination li > a[href]", func(e *colly.HTMLElement) {
		rel := e.Attr("rel")
		if rel == "next" {
			nextPage := e.Request.AbsoluteURL(e.Attr("href"))
			c.Visit(nextPage)
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping on https://hackerspaces.org
	c.Visit("https://itviec.com/it-jobs?query=&city=&commit=Search")

}
