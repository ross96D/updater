package share

import (
	_ "embed"
	"strconv"
	"strings"
)

//go:embed version
var _version string
var ver VersionData

type VersionData struct {
	Major int
	Minor int
	Patch int
}

func (v VersionData) IsLater(other VersionData) bool {
	if v == other {
		return false
	}

	if v.Major > other.Major {
		return true
	}
	if v.Major < other.Major {
		return false
	}
	if v.Minor > other.Minor {
		return true
	}
	if v.Minor < other.Minor {
		return false
	}
	if v.Patch > other.Patch {
		return true
	}
	if v.Patch < other.Patch {
		return false
	}
	return false
}

func VersionDataFromString(version string) (vd VersionData, err error) {
	// it seems that the json unmarshaler send the string with quotes.
	version, _ = strings.CutPrefix(version, "\"")
	version, _ = strings.CutSuffix(version, "\"")

	version, _ = strings.CutPrefix(version, "v")
	spl := strings.Split(version, ".")
	if vd.Major, err = strconv.Atoi(spl[0]); err != nil {
		return vd, err
	}
	if vd.Minor, err = strconv.Atoi(spl[1]); err != nil {
		return vd, err
	}
	if vd.Patch, err = strconv.Atoi(spl[2]); err != nil {
		return vd, err
	}
	return
}

func Version() VersionData {
	return ver
}

func init() {
	var err error
	if ver, err = VersionDataFromString(_version); err != nil {
		panic(err)
	}
}
