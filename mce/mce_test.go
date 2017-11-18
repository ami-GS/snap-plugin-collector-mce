package mce

import (
	"bufio"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

/*
func TestGetMetricTypes(t *testing.T) {
	Convey("if mcelog --daemon without any trigger setting enabled", t, func() {
	})
}
*/
func TestGetMceLog(t *testing.T) {
	mceLogPrefix := "../testlog/"
	Convey("if this is first call (), then returns all parsed info", t, func() {
		m := New()
		mcelogs, err := getMceLog(mceLogPrefix+"mcelog1", m.prevLogTimeStamp)
		So(err, ShouldBeNil)

		file, err := os.Open(mceLogPrefix + "mcelog1")
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
		m := New()
		mcelogs, err := getMceLog(mceLogPrefix+"mcelog0", m.prevLogTimeStamp)
		So(err, ShouldBeNil)
		So(len(mcelogs), ShouldBeZeroValue)
	})
	Convey("if no update from last call, then return empty slice", t, func() {
		m := New()
		mcelogs, err := getMceLog(mceLogPrefix+"mcelog1", m.prevLogTimeStamp)
		So(err, ShouldBeNil)
		// TODO : not here
		m.prevLogTimeStamp = mcelogs[len(mcelogs)-1].TIME
		mcelogs, err = getMceLog(mceLogPrefix+"mcelog1", m.prevLogTimeStamp)
		So(err, ShouldBeNil)
		So(len(mcelogs), ShouldBeZeroValue)
	})
	Convey("if single new log comes, return the info", t, func() {
		m := New()
		mcelogs, err := getMceLog(mceLogPrefix+"mcelog1", m.prevLogTimeStamp)
		So(err, ShouldBeNil)
		m.prevLogTimeStamp = mcelogs[0].TIME
		mcelogs, err = getMceLog(mceLogPrefix+"mcelog2", m.prevLogTimeStamp)
		So(err, ShouldBeNil)
		So(len(mcelogs), ShouldEqual, 1)
		file, err := os.Open(mceLogPrefix + "mcelog2")
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
			mcelogs, err := getMceLog(mceLogPrefix+"mcelog2", m.prevLogTimeStamp)
			So(err, ShouldBeNil)
		})
			Convey("AsItIs should have all log in mcelog file", t, func() {
				m := New()
				mcelogs, _ := getMceLog(mceLogPrefix+"mcelog1", m.prevLogTimeStamp)
				file, err := op.Open(mceLogPrefix + "mcelog1")
				if err != nil {
				}
				defer file.Close()
				mcelogs[0].AsItIs
			})
	*/
}
