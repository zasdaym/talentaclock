package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

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

var errInvalidArgument = fmt.Errorf("expected one argument: clock-in, clock-out, or check")

func run(ctx context.Context) error {
	cfg, err := parseConfig()
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	if len(os.Args) != 2 {
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

	var tasks chromedp.Tasks
	tasks = append(tasks, setGeolocation(cfg.Latitude, cfg.Longitude))
	tasks = append(tasks, signIn(cfg.TalentaEmail, cfg.TalentaPassword))
	switch os.Args[1] {
	case "clock-in":
		tasks = append(tasks, clockIn())
	case "clock-out":
		tasks = append(tasks, clockOut())
	case "check":
	default:
		return errInvalidArgument
	}

	return chromedp.Run(taskCtx, tasks)
}

func setGeolocation(latitude, longitude float64) chromedp.Tasks {
	d := browser.PermissionDescriptor{
		Name: browser.PermissionTypeGeolocation.String(),
	}
	return chromedp.Tasks{
		browser.SetPermission(&d, browser.PermissionSettingGranted),
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
