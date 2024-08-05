package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	//"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func main() {
	icsFile := flag.String("ics", "", "ICS calendar file to parse")
	credentials := flag.String("gcal", "", "Google Calendar credentials file")
	mountpoint := flag.String("mountpoint", "/Volumes/cal", "Directory to mount the calendar on")

	flag.Parse()

	var cal CalendarInterface
	var err error
	if *icsFile != "" {
		cal, err = OpenICal(*icsFile)
	} else if *credentials != "" {
		cal, err = OpenGCal(*credentials)
	} else {
		err = fmt.Errorf("no supported calendar interface supplied")
	}

	cal = NewCache(cal, time.Minute)

	root := &CalendarRootNode{cal: cal}

	if err != nil {
		log.Fatalf("Cannot start ical fs: %s", err)
	}
	server, err := fs.Mount(*mountpoint, root, &fs.Options{
		MountOptions: fuse.MountOptions{
			// Set to true to see how the file system works.
			Debug: true,
		},
	})
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Mounted on %s", *mountpoint)
	log.Printf("Unmount by calling 'fusermount -u %s'", *mountpoint)

	// Wait until unmount before exiting
	server.Wait()
}
