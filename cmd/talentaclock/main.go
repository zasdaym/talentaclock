package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err.Error())
	}
}

const talentaBaseURL = "https://hr.talenta.co"

var errInvalidArgument = fmt.Errorf("expected one argument: clock-in or clock-out")

func run(ctx context.Context) error {
	cfg, err := parseConfig()
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	if len(os.Args) < 2 {
		return errInvalidArgument
	}

	allocatorOpts := chromedp.DefaultExecAllocatorOptions[:]
	if cfg.Debug {
		allocatorOpts = append(allocatorOpts, chromedp.Flag("headless", false))
	}

	allocatorCtx, stop := chromedp.NewExecAllocator(ctx, allocatorOpts...)
	defer stop()

	taskCtx, stop := chromedp.NewContext(allocatorCtx, chromedp.WithLogf(log.Printf))
	defer stop()

	var finalAction chromedp.Tasks
	switch os.Args[1] {
	case "clock-in":
		finalAction = clockIn()
	case "clock-out":
		finalAction = clockOut()
	case "check":
	default:
		return errInvalidArgument
	}

	var todayNodeStyle string
	var lastTimeOffText string
	if err := chromedp.Run(
		taskCtx,
		setGeolocation(cfg.Latitude, cfg.Longitude),
		signIn(cfg.TalentaEmail, cfg.TalentaPassword),
		getTodayNodeStyle(&todayNodeStyle),
		getLastTimeOffText(&lastTimeOffText),
	); err != nil {
		return fmt.Errorf("sign in & initial check: %w", err)
	}

	if strings.Contains(todayNodeStyle, "red") {
		log.Printf("today is a holiday, skipping clock in/out")
		return nil
	}

	lastTimeOff, err := time.Parse("2006-01-02", lastTimeOffText)
	if err != nil {
		return fmt.Errorf("parse last time off date: %w", err)
	}
	if lastTimeOff.Format("2006-01-02") == time.Now().Format("2006-01-02") {
		log.Printf("last time off is today, skipping clock in/out")
		return nil
	}

	log.Printf("clocking in/out")
	if err := chromedp.Run(taskCtx, finalAction); err != nil {
		return fmt.Errorf("clock in/out: %w", err)
	}

	return nil
}

func setGeolocation(latitude, longitude float64) chromedp.Tasks {
	notification := browser.PermissionDescriptor{
		Name: browser.PermissionTypeNotifications.String(),
	}
	geolocation := browser.PermissionDescriptor{
		Name: browser.PermissionTypeGeolocation.String(),
	}
	return chromedp.Tasks{
		browser.SetPermission(&notification, browser.PermissionSettingGranted),
		browser.SetPermission(&geolocation, browser.PermissionSettingGranted),
		emulation.SetGeolocationOverride().
			WithAccuracy(100).
			WithLatitude(latitude).
			WithLongitude(longitude),
	}
}

func signIn(email, password string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(talentaBaseURL),
		chromedp.SendKeys("input#user_email", email),
		chromedp.SendKeys("input#user_password", password),
		chromedp.Click("#new-signin-button"),
		chromedp.WaitNotPresent(`#new-signin-button`),
	}
}

func openLiveAttendancePage() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(talentaBaseURL + "/live-attendance"),
		chromedp.WaitVisible("#tl-live-attendance-index"),
	}
}

func clockIn() chromedp.Tasks {
	return chromedp.Tasks{
		openLiveAttendancePage(),
		chromedp.Click(`//span[text()="Clock In"]`),
	}
}

func clockOut() chromedp.Tasks {
	return chromedp.Tasks{
		openLiveAttendancePage(),
		chromedp.Click(`//span[text()="Clock Out"]`),
	}
}

// getTodayNodeStyle gets the style attribute of the node that represents today.
// The style attribute will be used to determine if today is a holiday.
func getTodayNodeStyle(today *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(talentaBaseURL + "/employee/company-calendar"),
		chromedp.AttributeValue(`//td[contains(@class, "fc-today")]/span`, "style", today, nil),
	}
}

func getLastTimeOffText(timeOff *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(talentaBaseURL + "/my-info/time-off"),
		chromedp.Text(`//tr/td[@class="sorting_1"]`, timeOff),
	}
}
