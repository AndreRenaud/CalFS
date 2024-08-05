package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type CalendarEntry struct {
	Start time.Time
	End   time.Time

	Summary     string
	Description string
	Notes       string

	Other map[string]any
}

type CalendarInterface interface {
	Years() []int
	Months(int) []time.Month
	Days(int, time.Month) []int
	Entries(time.Time) []CalendarEntry
}

type CalendarRootNode struct {
	fs.Inode
	cal CalendarInterface
}

type icalYearNode struct {
	fs.Inode

	cal  CalendarInterface
	year int
}

type icalMonthNode struct {
	fs.Inode

	cal   CalendarInterface
	year  int
	month time.Month
}

type icalDayNode struct {
	fs.Inode

	cal   CalendarInterface
	year  int
	month time.Month
	day   int
}

func (icn *CalendarRootNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	years := icn.cal.Years()
	r := make([]fuse.DirEntry, 0, len(years))
	for _, y := range years {
		d := fuse.DirEntry{
			Name: fmt.Sprintf("%4.4d", y),
			Ino:  uint64(y),
			Mode: fuse.S_IFDIR,
		}
		r = append(r, d)
	}
	return fs.NewListDirStream(r), fs.OK
}

func (icn *CalendarRootNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	year, err := strconv.Atoi(name)
	if err != nil {
		return nil, syscall.ENOENT
	}

	n := sort.SearchInts(icn.cal.Years(), year)
	if n < 0 {
		return nil, syscall.ENOENT
	}

	stable := fs.StableAttr{
		Mode: fuse.S_IFDIR,
		// The child inode is identified by its Inode number.
		// If multiple concurrent lookups try to find the same
		// inode, they are deduplicated on this key.
		Ino: uint64(year),
	}
	operations := &icalYearNode{cal: icn.cal, year: year}
	child := icn.Inode.NewInode(ctx, operations, stable)

	if out != nil {
		t := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
		out.Mtime = uint64(t.Unix())
		out.Atime = uint64(t.Unix())
		out.Ctime = uint64(t.Unix())
		out.AttrValid |= fuse.FATTR_ATIME | fuse.FATTR_CTIME | fuse.FATTR_MTIME
	}

	return child, fs.OK
}

func (icn *icalYearNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	months := icn.cal.Months(icn.year)
	r := make([]fuse.DirEntry, 0, len(months))
	for _, m := range months {
		d := fuse.DirEntry{
			Name: m.String(),
			Ino:  uint64(m),
			Mode: fuse.S_IFDIR,
		}
		r = append(r, d)
	}
	return fs.NewListDirStream(r), fs.OK
}

func (icy *icalYearNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if len(name) > 3 {
		name = name[:3]
	}
	month, err := time.Parse("Jan", name)
	if err != nil {
		log.Printf("Cannot parse %s: %s", name, err)
		return nil, fs.ENOATTR
	}
	log.Printf("month: %s", month)
	// TODO: Do something correct here
	stable := fs.StableAttr{
		Mode: fuse.S_IFDIR,
		// The child inode is identified by its Inode number.
		// If multiple concurrent lookups try to find the same
		// inode, they are deduplicated on this key.
		Ino: 0, // TODO
	}
	operations := &icalMonthNode{cal: icy.cal, year: icy.year, month: month.Month()}
	child := icy.Inode.NewInode(ctx, operations, stable)
	if out != nil {
		t := time.Date(icy.year, month.Month(), 1, 0, 0, 0, 0, time.Local)
		out.Mtime = uint64(t.Unix())
		out.Atime = uint64(t.Unix())
		out.Ctime = uint64(t.Unix())
		out.AttrValid |= fuse.FATTR_ATIME | fuse.FATTR_CTIME | fuse.FATTR_MTIME
	}

	return child, fs.OK
}

func (icm *icalMonthNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	days := icm.cal.Days(icm.year, icm.month)
	r := make([]fuse.DirEntry, 0, len(days))
	for _, day := range days {
		d := fuse.DirEntry{
			Name: fmt.Sprintf("%02d", day),
			Ino:  uint64(day),
			Mode: fuse.S_IFREG,
		}
		r = append(r, d)
	}
	return fs.NewListDirStream(r), fs.OK
}

func (icm *icalMonthNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	day, err := strconv.Atoi(name)
	if err != nil {
		log.Printf("Cannot parse %s: %s", name, err)
		return nil, fs.ENOATTR
	}
	// TODO: Do something correct here
	stable := fs.StableAttr{
		Mode: fuse.S_IFREG,
		// The child inode is identified by its Inode number.
		// If multiple concurrent lookups try to find the same
		// inode, they are deduplicated on this key.
		Ino: 0, // TODO
	}
	operations := &icalDayNode{cal: icm.cal, year: icm.year, month: icm.month, day: day}
	child := icm.Inode.NewInode(ctx, operations, stable)

	if out != nil {
		t := time.Date(icm.year, icm.month, day, 0, 0, 0, 0, time.Local)
		out.Mtime = uint64(t.Unix())
		out.Atime = uint64(t.Unix())
		out.Ctime = uint64(t.Unix())
		out.AttrValid |= fuse.FATTR_ATIME | fuse.FATTR_CTIME | fuse.FATTR_MTIME
	}

	return child, fs.OK
}

func (icd *icalDayNode) Open(ctx context.Context, openFlags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	// disallow writes
	if openFlags&(syscall.O_RDWR|syscall.O_WRONLY) != 0 {
		return nil, 0, syscall.EROFS
	}

	return icd, fuse.FOPEN_DIRECT_IO, 0
}

func (icd *icalDayNode) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	var content bytes.Buffer

	fmt.Fprintf(&content, "====================\n%04d-%s-%02d\n", icd.year, icd.month, icd.day)

	now := time.Date(icd.year, icd.month, icd.day, 0, 0, 0, 0, time.Local)

	for _, e := range icd.cal.Entries(now) {
		year, month, day := e.Start.Date()
		if icd.year == year && icd.month == month && icd.day == day {
			duration := e.End.Sub(e.Start)
			fmt.Fprintf(&content, "---\n%s - %s (%s)\n%s\n%s", e.Start, e.End, duration, e.Summary, e.Description)
		}
	}

	end := int(off) + content.Len()
	if end > content.Len() {
		end = content.Len()
	}

	return fuse.ReadResultData(content.Bytes()[off:end]), fs.OK
}
