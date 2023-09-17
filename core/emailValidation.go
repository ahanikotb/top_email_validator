package core

import (
	"context"
	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

var (
	verifier = emailverifier.
		NewVerifier().
		EnableSMTPCheck()
	//.EnableCatchAllCheck()
)

var genericEmails = []string{
	"info", "contact", "support", "sales", "billing", "admin", "hello", "help", "team", "press", "media", "jobs", "career", "hr", "recruit", "recruitment", "marketing", "market", "business", "partners", "partner", "invest", "investment", "investor", "investors", "feedback", "suggestions", "suggestion", "feedbacks", "suggestions", "feedbacks", "suggestion", "suggestions"}

func BruteForceValidateEmail(fname string, lname string, domain string, useGenerics bool) (string, string) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Println("Recovered in f", r)
		}
	}()
	if domain == "" {
		return "null", "failed"
	}
	options := MakeNameOptions(CleanName(fname), CleanName(lname))
	for i := range options {
		username := options[i]
		res := ValidateEmail(username + "@" + domain)
		if res != "failed" && res != "not_valid" {
			return username + "@" + domain, res
		}
	}
	if useGenerics {
		for i := range genericEmails {
			username := genericEmails[i]
			res := ValidateEmail(username + "@" + domain)
			if res != "failed" && res != "not_valid" {
				return username + "@" + domain, res

			}
		}
	}
	return "null", "failed"
}

func ValidateEmail(email string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // set a timeout of 5 seconds
	defer cancel()

	ch := make(chan string, 1) // create a channel to receive the result

	go func() {
		ret, err := verifier.CheckSMTP(strings.Split(email, "@")[1], strings.Split(email, "@")[0])
		//fmt.Println(ret, err)
		if err != nil {
			ch <- "failed"
		}
		if ret.CatchAll == true {
			ch <- "catch_all"
		}
		if ret.FullInbox == true {
			ch <- "full_inbox"
		}
		if ret.Deliverable == true {
			ch <- "valid"
		}
		ch <- "not_valid"
	}()

	select {
	case result := <-ch: // wait for a result from the channel
		return result
	case <-ctx.Done(): // if the timeout has elapsed, return "timeout"
		return "null"
	}
}
func CleanName(s string) string {
	//trim spaces
	s = strings.TrimSpace(s)
	//replace all spaces
	s = strings.ReplaceAll(s, " ", "")
	//replace all dashes
	s = strings.ReplaceAll(s, "-", "")
	//lowercase the name
	s = strings.ToLower(s)
	return s
}
func CleanFName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
func CleanLName(s string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(s)), " ", "")
}
func ParseDomain(d string) string {
	d = strings.Split(strings.Replace(strings.Replace(strings.Replace(d, "https://", "", -1), "http://", "", -1), "www.", "", -1), "/")[0]
	return d
}
func MakeNameOptions(f_name string, l_name string) []string {
	if len(f_name) < 1 && len(l_name) < 1 {
		return []string{}
	}
	if len(f_name) < 1 {
		return []string{
			l_name,
		}
	}
	if len(l_name) < 1 {
		return []string{
			f_name,
		}
	}
	options := []string{
		string(f_name[0]) + l_name,
		//string(f_name[0]) + "_" + l_name,
		string(f_name[0]) + "." + l_name,
		f_name + "." + l_name,
		f_name + "." + string(l_name[0]),
		//f_name + "_" + string(l_name[0]),
		f_name + l_name,
		//f_name + "_" + l_name,
		f_name,
		l_name,
	}
	return options
}
