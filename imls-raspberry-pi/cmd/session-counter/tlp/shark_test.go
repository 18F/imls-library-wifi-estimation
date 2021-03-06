package tlp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"gsa.gov/18f/internal/state"
	"gsa.gov/18f/internal/structs"
	"gsa.gov/18f/internal/wifi-hardware-search/models"
)

// There's a lot of copypasta in these tests.

func cleanupTempFiles() {
	cfg := state.GetConfig()

	f1, err := filepath.Glob(filepath.Join(cfg.GetWWWRoot(), "*.sqlite*"))
	if err != nil {
		panic(err)
	}
	f2, err := filepath.Glob(filepath.Join(cfg.GetWWWImages(), "*.png"))
	if err != nil {
		panic(err)
	}

	for _, f := range append(f1, f2...) {
		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}
}
func setup() {
	tempDB, err := os.CreateTemp("", "shark-test.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	state.SetConfigAtPath(tempDB.Name())
	cfg := state.GetConfig()
	cfg.SetRunMode("test")
	cfg.SetStorageMode("sqlite")

	_, filename, _, _ := runtime.Caller(0)
	fmt.Println(filename)
	path := filepath.Dir(filename)
	cfg.SetQueuesPath(filepath.Join(path, "..", "test", "www", "queues.sqlite"))
	cfg.SetDurationsPath(filepath.Join(path, "..", "test", "www", "durations.sqlite"))
	cfg.SetRootPath(filepath.Join(path, "..", "test", "www"))
	cfg.SetImagesPath(filepath.Join(path, "..", "test", "www", "images"))
	cfg.SetFCFSSeqID("ME0000-001")
	cfg.SetDeviceTag("testing")
	cfg.Log().SetLogLevel("DEBUG")

	state.FlushCache()
	log.Println("Calling init config in setup")
	log.Println("Trying to get session id")
	log.Println("session id is ", cfg.GetCurrentSessionID())
	cfg.Log().SetLogLevel("DEBUG")

	os.Mkdir(cfg.GetWWWRoot(), 0755)
	os.Mkdir(cfg.GetWWWImages(), 0755)
	mock := clock.NewMock()
	mt, _ := time.Parse("2006-01-02T15:04", "1975-10-11T02:00")
	mock.Set(mt)
	state.SetClock(mock)

	if state.GetClock() == nil {
		cfg.Log().Fatal("clock should not be nil")
	}
	cfg.Log().Debug("mock is now ", state.GetClock().Now().In(time.Local))
}

// type SharkFn func(string) []string
// type MonitorFn func(*models.Device)
// type SearchFn func() *models.Device

func fakeMonitorFn(d *models.Device) {

}

func fakeSearchFn() (d *models.Device) {
	d = &models.Device{Exists: true, Logicalname: "fakewan0"}
	return d
}

func fakeShark2(dev string) []string {
	return []string{"DE:AD:BE:EF:00:00", "BE:EF:00:00:00:00"}
}

func fakeShark1(dev string) []string {
	return []string{"DE:AD:BE:EF:00:00"}
}
func TestOneHour(t *testing.T) {

	setup()
	cleanupTempFiles()
	cfg := state.GetConfig()

	// Create a DB for simpleshark to write to.
	db := state.NewSqliteDB(":memory:")
	db.CreateTableFromStruct(structs.EphemeralDuration{})

	startTime, _ := time.Parse(time.RFC3339, "1975-10-11T08:00:00-04:00")
	endTime, _ := time.Parse(time.RFC3339, "1975-10-11T09:00:00-04:00")

	mock := clock.NewMock()
	// Bump the clock forward
	mock.Set(startTime)
	state.SetClock(mock)
	// Run once at the initial time.
	SimpleShark(db, fakeMonitorFn, fakeSearchFn, fakeShark2)
	mock.Set(endTime)
	state.SetClock(mock)

	SimpleShark(db, fakeMonitorFn, fakeSearchFn, fakeShark2)

	// We should now be able to check the DB.
	for _, testmac := range []string{"DE:AD:BE:EF:00:00", "BE:EF:00:00:00:00"} {
		var ed structs.EphemeralDuration
		row := db.GetPtr().QueryRowx("SELECT * FROM ephemeraldurations WHERE mac=?", testmac)
		err := row.StructScan(&ed)
		if err != nil {
			cfg.Log().Error("We did not get a struct for ", testmac)
			cfg.Log().Fatal(err)
		} else {
			if !((ed.MAC == testmac) && (ed.Start == startTime.Unix()) && (ed.End == endTime.Unix())) {
				cfg.Log().Error("things do not add up for ", testmac)
				cfg.Log().Error(ed.MAC, testmac, ed.MAC == testmac)
				cfg.Log().Error(startTime.Unix(), ed.Start, (ed.Start == startTime.Unix()))
				cfg.Log().Error(endTime.Unix(), ed.End, (ed.End == endTime.Unix()))
				t.Fail()
			}
		}
	}

}

func TestOneYear(t *testing.T) {

	setup()
	cleanupTempFiles()
	cfg := state.GetConfig()

	// Create a DB for simpleshark to write to.
	db := state.NewSqliteDB(":memory:")
	db.CreateTableFromStruct(structs.EphemeralDuration{})

	startTime, _ := time.Parse(time.RFC3339, "1975-10-11T08:00:00-04:00")
	endTime, _ := time.Parse(time.RFC3339, "1976-10-11T09:00:00-04:00")

	mock := clock.NewMock()
	// Bump the clock forward
	mock.Set(startTime)
	state.SetClock(mock)
	// Run once at the initial time.
	SimpleShark(db, fakeMonitorFn, fakeSearchFn, fakeShark2)
	mock.Set(endTime)
	state.SetClock(mock)

	SimpleShark(db, fakeMonitorFn, fakeSearchFn, fakeShark2)

	// We should now be able to check the DB.
	for _, testmac := range []string{"DE:AD:BE:EF:00:00", "BE:EF:00:00:00:00"} {
		var ed structs.EphemeralDuration
		row := db.GetPtr().QueryRowx("SELECT * FROM ephemeraldurations WHERE mac=?", testmac)
		err := row.StructScan(&ed)
		if err != nil {
			cfg.Log().Error("We did not get a struct for ", testmac)
			cfg.Log().Fatal(err)
		} else {
			if !((ed.MAC == testmac) && (ed.Start == startTime.Unix()) && (ed.End == endTime.Unix())) {
				cfg.Log().Error("things do not add up for ", testmac)
				cfg.Log().Error(ed.MAC, testmac, ed.MAC == testmac)
				cfg.Log().Error(startTime.Unix(), ed.Start, (ed.Start == startTime.Unix()))
				cfg.Log().Error(endTime.Unix(), ed.End, (ed.End == endTime.Unix()))
				t.Fail()
			}
		}
	}

}

func TestBumpOne(t *testing.T) {

	setup()
	cleanupTempFiles()
	cfg := state.GetConfig()

	// Create a DB for simpleshark to write to.
	db := state.NewSqliteDB(":memory:")
	db.CreateTableFromStruct(structs.EphemeralDuration{})

	startTime, _ := time.Parse(time.RFC3339, "1975-10-11T08:00:00-04:00")
	endTime, _ := time.Parse(time.RFC3339, "1975-10-11T09:00:00-04:00")

	mock := clock.NewMock()
	// Bump the clock forward
	mock.Set(startTime)
	state.SetClock(mock)
	// Run once at the initial time.
	SimpleShark(db, fakeMonitorFn, fakeSearchFn, fakeShark2)
	mock.Set(endTime)
	state.SetClock(mock)

	SimpleShark(db, fakeMonitorFn, fakeSearchFn, fakeShark1)

	// We should now be able to check the DB.
	for _, testmac := range []string{"DE:AD:BE:EF:00:00"} {
		var ed structs.EphemeralDuration
		row := db.GetPtr().QueryRowx("SELECT * FROM ephemeraldurations WHERE mac=?", testmac)
		err := row.StructScan(&ed)
		if err != nil {
			cfg.Log().Error("We did not get a struct for ", testmac)
			cfg.Log().Fatal(err)
		} else {
			if !((ed.MAC == testmac) && (ed.Start == startTime.Unix()) && (ed.End == endTime.Unix())) {
				cfg.Log().Error("things do not add up for ", testmac)
				cfg.Log().Error(ed.MAC, testmac, ed.MAC == testmac)
				cfg.Log().Error(startTime.Unix(), ed.Start, (ed.Start == startTime.Unix()))
				cfg.Log().Error(endTime.Unix(), ed.End, (ed.End == endTime.Unix()))
				t.Fail()
			}
		}
	}

	for _, testmac := range []string{"BE:EF:00:00:00:00"} {
		var ed structs.EphemeralDuration
		row := db.GetPtr().QueryRowx("SELECT * FROM ephemeraldurations WHERE mac=?", testmac)
		err := row.StructScan(&ed)
		if err != nil {
			cfg.Log().Error("We did not get a struct for ", testmac)
			cfg.Log().Fatal(err)
		} else {
			if (ed.MAC == testmac) && (ed.Start == startTime.Unix()) && (ed.End == endTime.Unix()) {
				cfg.Log().Error("things DO add up for ", testmac)
				cfg.Log().Error(ed.MAC, testmac, ed.MAC == testmac)
				cfg.Log().Error(startTime.Unix(), ed.Start, (ed.Start == startTime.Unix()))
				cfg.Log().Error(endTime.Unix(), ed.End, (ed.End == endTime.Unix()))
				t.Fail()
			}
		}
	}

}
