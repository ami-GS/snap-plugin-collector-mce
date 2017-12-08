package mceStream

import (
	"testing"

	"github.com/ami-GS/snap-plugin-collector-mce/mce_ondemand/mce"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewStream(t *testing.T) {
	logPath := "/testlogpath/mcelog"
	Convey("if New, then return default", t, func() {
		stream := NewStream(logPath)
		So(stream, ShouldResemble,
			&MCEStreamCollector{
				mce.New(logPath),
				nil,
			})
	})
}
