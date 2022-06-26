package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gocolly/colly/v2"
)

type Company struct {
	ID           primitive.ObjectID `bson:"_id"`
	Name         string
	Logo         string
	CoverImage   string
	Size         string
	Location     string
	Stars        float64
	KeySkills    []string
	Path         string
	National     string
	Service      string
	BusinessDays string
	jobCounts    int32
	IsOverTime   bool
	JobsDetails  []JobDetails
	CreatedAt    int64
	UpdatedAt    int64
}

type JobDetails struct {
	ID                 primitive.ObjectID `bson:"_id"`
	Title              string
	Tags               []string
	Location           string
	TopReason          string
	JobDescription     string
	SkillAndExperience string
	LoveWorking        string
	CreatedAt          int64
	UpdatedAt          int64
	Path               string
	CompanyId          primitive.ObjectID
}

var (
	regex                = regexp.MustCompile("\\s+")
	jobURLToCompanyIdMap = make(map[string]primitive.ObjectID)
	//mapLock              sync.RWMutex
)

func main() {

	// get Client, Context, CancelFunc and err from connect method.
	client, ctx, cancel, err := connect("mongodb://admin:admin@127.0.0.1:27017/")
	if err != nil {
		panic(err)
	}

	// Release resource when main function is returned.
	defer close(client, ctx, cancel)

	// Ping mongoDB with Ping method
	ping(client, ctx)

	if err != nil {
		panic(err)
	}

	// Create  a object of type interface to  store
	// the bson values, that  we are inserting into database.
	// var document interface{}

	// document = bson.D{
	// 	{"rollNo", 175},
	// 	{"maths", 80},
	// 	{"science", 90},
	// 	{"computer", 95},
	// }

	// // insertOne accepts client , context, database
	// // name collection name and an interface that
	// // will be inserted into the  collection.
	// // insertOne returns an error and a result of
	// // insert in a single document into the collection.
	// insertOneResult, err := insertOne(client, ctx, "gfg",
	// 	"marks", document)

	// // handle the error
	// if err != nil {
	// 	panic(err)
	// }

	// print the insertion id of the document,
	// // if it is inserted.
	// fmt.Println("Result of InsertOne")
	// fmt.Println(insertOneResult.InsertedID)

	// Instantiate default collector
	companyCollector := colly.NewCollector(
		// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
		colly.AllowedDomains("itviec.com"),
		// Allow visiting the same page multiple times
		colly.AllowURLRevisit(),

		colly.Async(true),
	)

	companyCollector.Limit(&colly.LimitRule{
		// Filter domains affected by this rule
		// Set a delay between requests to these domains
		Delay: 1 * time.Second,
		// Add an additional random delay
		RandomDelay: 1 * time.Second,
	})

	// authenticate
	// err := c.Post("http://itviec.com/sign_in", map[string]string{"user_email": "andrew12@yopmail.com", "password": "12345678"})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// cookie := ""

	// // attach callbacks after login
	// c.OnResponse(func(r *colly.Response) {
	// 	log.Println("response received", r.StatusCode)
	// 	cookie = r.Headers.Get("set-cookie")
	// 	log.Printf("%s\n", cookie)
	// 	r.Save("temp.txt")

	// })

	//On every a element which has href attribute call callback
	// c.OnHTML("a.job__skill", func(e *colly.HTMLElement) {
	// 	fmt.Println(e.ChildText("span"))
	// 	// link := e.Attr("href")
	// 	// // Print link
	// 	// fmt.Printf("Link found: %q -> %s\n", e.Text, link)
	// 	// Visit link found on page
	// 	// Only those links are visited which are in AllowedDomains
	// 	// c.Visit(e.Request.AbsoluteURL(link))
	// })
	//crawlJob(c, client)

	jobCollector := companyCollector.Clone()
	crawlJob(jobCollector, client, ctx)

	crawlCompany(companyCollector, jobCollector, client, ctx)

	companyCollector.Wait()
	jobCollector.Wait()
}

func crawlJob(c *colly.Collector, client *mongo.Client, ctx context.Context) {
	// c.OnHTML("div.details > h3.title.job-details-link-wrapper > a[href]", func(e *colly.HTMLElement) {
	// 	link := e.Attr("href")

	// 	// Print link
	// 	fmt.Printf("Link found: %q -> %s\n", e.Text, link)

	// 	// absoluteURL := e.Request.AbsoluteURL(link)
	// 	// Visit link found on page
	// 	// Only those links are visited which are in AllowedDomains
	// 	c.Visit(e.Request.AbsoluteURL(link))
	// })

	c.OnHTML("div.jd-page__job-details div.job-details", func(e *colly.HTMLElement) {
		var jobDetails JobDetails
		fmt.Printf("\n\n")
		fmt.Println("-------JOB DETAILS--------")
		fmt.Println("---------------------------")

		jobDetails.Path = e.Request.URL.String()
		fmt.Printf("Path: %s\n", jobDetails.Path)

		jobDetails.Title = e.ChildText("h1.job-details__title")
		fmt.Printf("Title: %s\n", jobDetails.Title)

		tags := e.ChildText("div.job-details__tag-list span")
		tags = regex.ReplaceAllString(tags, " ")
		fmt.Printf("Tags: %s\n", tags)
		jobDetails.Tags = strings.Split(tags, " ")

		fmt.Printf("Salary: %s\n", e.ChildText("div.job-details__overview > div.svg-icon--green > div.svg-icon__text"))

		jobDetails.Location = e.ChildText("div.job-details__overview > div.svg-icon > div.svg-icon__text > span")
		fmt.Printf("Location: %s\n", jobDetails.Location)

		fmt.Printf("%s: \n", e.ChildText("h2.job-details__second-title:nth-of-type(1)"))
		topReasonHtml, _ := e.DOM.Find("div.job-details__top-reason-to-join-us").Html()
		fmt.Printf("%s\n", topReasonHtml)
		jobDetails.TopReason = topReasonHtml

		fmt.Printf("%s\n", e.ChildText("h2.job-details__second-title:nth-of-type(2)"))
		paragraphDom := e.DOM.Find(".job-details__paragraph")
		jobDescriptionHtml, _ := paragraphDom.Eq(0).Html()
		fmt.Printf("%s\n", jobDescriptionHtml)

		jobDetails.JobDescription = jobDescriptionHtml

		fmt.Printf("%s\n", e.ChildText("h2.job-details__second-title:nth-of-type(3)"))
		skillAndExperienceHtml, _ := paragraphDom.Eq(1).Html()
		fmt.Printf("%s\n", skillAndExperienceHtml)
		jobDetails.SkillAndExperience = skillAndExperienceHtml

		fmt.Printf("%s\n", e.ChildText("h2.job-details__second-title:nth-of-type(4)"))
		loveWorkingHtml, _ := paragraphDom.Eq(2).Html()
		fmt.Printf("%s\n", loveWorkingHtml)
		jobDetails.LoveWorking = loveWorkingHtml
		jobDetails.CreatedAt = time.Now().Unix()
		jobDetails.UpdatedAt = jobDetails.CreatedAt
		jobDetails.ID = primitive.NewObjectID()
		jobDetails.CompanyId = jobURLToCompanyIdMap[jobDetails.Path]
		// fmt.Printf("Title: %s\n", e.Text)
		// fmt.Printf("Title: %s\n", e.Text)
		// fmt.Printf("Title: %s\n", e.Text)
		// fmt.Printf("Title: %s\n", e.Text)
		ctx2, _ := context.WithTimeout(ctx, 1000*time.Millisecond)
		//defer cancel() // releases resources if slowOperation completes before timeout elapses
		insertOneResult, err := insertOne(client, ctx2, "jobs",
			"job-details", jobDetails)

		// handle the error
		if err != nil {
			log.Println(err)
		}

		// if it is inserted.
		fmt.Println("Result of InsertOne")
		fmt.Println(insertOneResult)

		fmt.Println("---------------------------")
		fmt.Println("---------------------------")
		fmt.Printf("\n\n")
	})

	// c.OnHTML("ul.pagination li > a[href]", func(e *colly.HTMLElement) {
	// 	rel := e.Attr("rel")
	// 	if rel == "next" {
	// 		nextPage := e.Request.AbsoluteURL(e.Attr("href"))
	// 		c.Visit(nextPage)
	// 	}
	// })

	// c.Limit(&colly.LimitRule{
	// 	// Filter domains affected by this rule
	// 	// Set a delay between requests to these domains
	// 	DomainGlob: "*",
	// 	Delay:      1 * time.Second,
	// 	// Add an additional random delay
	// 	RandomDelay: 1 * time.Second,
	// })
	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping on https://hackerspaces.org
	//c.Visit("https://itviec.com/it-jobs")

}

func crawlCompany(companyCollector *colly.Collector, jobCollector *colly.Collector, client *mongo.Client, ctx context.Context) {

	companyCollector.OnHTML("div.featured-companies a.featured-company", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		link = strings.TrimSuffix(link, "/review")
		// Print link
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		// absoluteURL := e.Request.AbsoluteURL(link)
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		companyCollector.Visit(e.Request.AbsoluteURL(link))
	})

	companyCollector.OnHTML("div.company-content > div.company-page", func(e *colly.HTMLElement) {
		var company Company
		fmt.Printf("\n\n")

		company.Path = e.Request.URL.String()
		fmt.Printf("Path: %s\n", company.Path)

		company.Name = e.ChildText(".headers h1.headers__info__name")
		fmt.Printf("Name: %s\n", company.Name)

		ratings := e.ChildText("div.company-page__container span.company-ratings__star-point")
		floatNum, err := strconv.ParseFloat(ratings, 64)
		if err != nil {
			log.Println(err)
		} else {
			company.Stars = floatNum
		}

		fmt.Printf("Stars: %f\n", company.Stars)

		//tags := e.ChildText("")
		//tags = regex.ReplaceAllString(tags, " ")
		//tags = strings.TrimSpace(tags)

		e.ForEach("div.company-page__container ul.employer-skills .employer-skills__item > a[href]", func(_ int, elem *colly.HTMLElement) {
			fmt.Println(elem.Text)
			company.KeySkills = append(company.KeySkills, elem.Text)

		})

		fmt.Printf("KeySkills: %s\n", company.KeySkills)

		info := e.DOM.Find(".headers .headers__info .svg-icon__text")
		company.Location, _ = info.Eq(0).Html()
		fmt.Printf("Location: %s\n", company.Location)

		company.Location, _ = info.Eq(0).Html()
		fmt.Printf("Location: %s\n", company.Location)

		company.Service, _ = info.Eq(1).Html()
		fmt.Printf("Service: %s\n", company.Service)

		company.Size, _ = info.Eq(2).Html()
		fmt.Printf("Size: %s\n", company.Size)

		company.National, _ = info.Eq(3).Html()
		fmt.Printf("National: %s\n", company.National)

		company.BusinessDays, _ = info.Eq(4).Html()
		fmt.Printf("BusinessDays: %s\n", company.BusinessDays)

		overTime, _ := info.Eq(5).Html()
		if overTime == "No OT" {
			company.IsOverTime = false
		} else {
			company.IsOverTime = true
		}

		fmt.Printf("OverTime: %s\n", overTime)

		company.ID = primitive.NewObjectID()

		e.ForEach("div.company-page__container .company-page__left .job .job-details-link-wrapper a[href]", func(_ int, elem *colly.HTMLElement) {
			link := elem.Attr("href")
			// Print link
			fmt.Printf("Link found: %q -> %s\n", elem.Text, link)
			// absoluteURL := e.Request.AbsoluteURL(link)
			// Visit link found on page
			// Only those links are visited which are in AllowedDomains
			//
			absoluteURL := elem.Request.AbsoluteURL(link)
			jobURLToCompanyIdMap[absoluteURL] = company.ID
			jobCollector.Visit(absoluteURL)

		})

		fmt.Printf("%v", company)

		ctx2, _ := context.WithTimeout(ctx, 1000*time.Millisecond)
		//defer cancel() // releases resources if slowOperation completes before timeout elapses
		insertOneResult, err := insertOne(client, ctx2, "jobs",
			"company", company)

		// handle the error
		if err != nil {
			log.Println(err)
		}

		// if it is inserted.
		fmt.Println("Result of InsertOne")
		fmt.Println(insertOneResult)

		fmt.Println("---------------------------")
		fmt.Println("---------------------------")
		fmt.Printf("\n\n")
	})

	// companyCollector.OnHTML("h3.job-details-link-wrapper a[href]", func(e *colly.HTMLElement) {
	// 	link := e.Attr("href")
	// 	fmt.Printf("Link found: %q -> %s\n", e.Text, link)
	// 	nextPage := e.Request.AbsoluteURL(link)
	// 	jobCollector.Visit(nextPage)

	// })

	companyCollector.OnHTML("div.featured-companies #show-more-wrapper > #show_more a[href]", func(e *colly.HTMLElement) {
		rel := e.Attr("rel")
		if rel == "next" {
			nextPage := e.Request.AbsoluteURL(e.Attr("href"))
			companyCollector.Visit(nextPage)
		}
	})

	companyCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting URL: ", r.URL.String())
	})

	companyCollector.Visit("https://itviec.com/companies?page=1")

}
