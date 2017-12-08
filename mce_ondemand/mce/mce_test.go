package mce

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

var mceLogPrefix = "/Users/daminaka/Go/src/github.com/ami-GS/snap-plugin-collector-mce/testlog/"

func TestGetMetricTypes(t *testing.T) {
	vName := "ami-GS"
	pName := "mce"
	m := New(mceLogPrefix + "mcelog1")
	allNameList := []string{
		plugin.NewNamespace(vName, pName, MetricAll).String(),
	}
	for _, name := range AllMetricsNames {
		if name == "Includes" {
			for _, name2 := range AllMetricsNames {
				allNameList = append(allNameList, plugin.NewNamespace(vName, pName, name, name2).String())
			}
		} else {
			allNameList = append(allNameList, plugin.NewNamespace(vName, pName, name).String())
		}
	}
	Convey("if nothing special, then returns default values", t, func() {
		rVal, err := m.GetMetricTypes(nil)
		So(err, ShouldBeNil)
		for _, v := range rVal {
			So(v.Namespace.String(), ShouldBeIn, allNameList)
		}
	})
}

func TestGetMceLog(t *testing.T) {
	Convey("if this is first call (), then returns all parsed info", t, func() {
		logfile := mceLogPrefix + "mcelog1"
		m := New(logfile)
		mcelogs, err := m.GetMceLog()
		So(err, ShouldBeNil)

		file, err := os.Open(logfile)
		So(err, ShouldBeNil)
		defer file.Close()
		sc := bufio.NewScanner(file)
		str := ""
		for sc.Scan() {
			data := strings.TrimSpace(sc.Text())
			if strings.HasPrefix(data, "mcelog: failed") {
				continue
			}
			str += data + "\n"
		}
		So(mcelogs[0].AsItIs, ShouldEqual, str)
	})
	Convey("if no log in mcelog, then return empty slice", t, func() {
		m := New(mceLogPrefix + "mcelog0")
		mcelogs, err := m.GetMceLog()
		So(err, ShouldBeNil)
		So(len(mcelogs), ShouldBeZeroValue)
	})
	Convey("if no update from last call, then return empty slice", t, func() {
		logfile := mceLogPrefix + "mcelog1"
		m := New(logfile)
		mcelogs, err := m.GetMceLog()
		So(err, ShouldBeNil)
		mcelogs, err = m.GetMceLog()
		So(err, ShouldBeNil)
		So(len(mcelogs), ShouldBeZeroValue)
	})
	Convey("if single new log comes, return the info", t, func() {
		m := New(mceLogPrefix + "mcelog1")
		mcelogs, err := m.GetMceLog()
		So(err, ShouldBeNil)
		m.LogPath = mceLogPrefix + "mcelog2"
		mcelogs, err = m.GetMceLog()
		So(err, ShouldBeNil)
		So(len(mcelogs), ShouldEqual, 1)
		file, err := os.Open(m.LogPath)
		So(err, ShouldBeNil)
		defer file.Close()
		sc := bufio.NewScanner(file)
		str := ""
		ignorecount := 0
		for sc.Scan() {
			data := strings.TrimSpace(sc.Text())
			if strings.HasPrefix(data, "mcelog: failed") || strings.HasPrefix(data, "Hardware event.") {
				ignorecount++
			}
			if ignorecount < 3 {
				continue
			}
			str += data + "\n"
		}
		So(mcelogs[0].AsItIs, ShouldEqual, str)
	})
	/*
			Convey("if several (>1) new logs come, return these info", t, func() {
				m := New()
		p		mcelogs, err := m.GetMceLog(mceLogPrefix+"mcelog2", m.prevLogTimeStamp)
				So(err, ShouldBeNil)
			})
				Convey("AsItIs should have all log in mcelog file", t, func() {
					m := New()
					mcelogs, _ := m.GetMceLog(mceLogPrefix+"mcelog1", m.prevLogTimeStamp)
					file, err := op.Open(mceLogPrefix + "mcelog1")
					if err != nil {
					}
					defer file.Close()
					mcelogs[0].AsItIs
				})
	*/
}

func TestWasFileUpdated(t *testing.T) {
	Convey("if first call, return true, nil", t, func() {
		m := New(mceLogPrefix + "mcelog1")
		ok, err := m.WasFileUpdated()
		So(ok, ShouldBeTrue)
		So(err, ShouldBeNil)
	})
	Convey("if no update, return false, nil", t, func() {
		m := New(mceLogPrefix + "mcelog1")
		ok, err := m.WasFileUpdated()
		ok, err = m.WasFileUpdated()
		So(ok, ShouldBeFalse)
		So(err, ShouldBeNil)
	})
	/* // This test doesn't work on Travis
	Convey("if update, return true, nil", t, func() {
		m := New(mceLogPrefix + "mcelog1")
		ok, err := m.WasFileUpdated()
		m.LogPath = mceLogPrefix + "mcelog2"

		ok, err = m.WasFileUpdated()
		So(ok, ShouldBeTrue)
		So(err, ShouldBeNil)
	})
	*/
}
