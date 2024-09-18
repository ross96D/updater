package share

import (
	"bytes"
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

func (v VersionData) String() string {
	b := bytes.Buffer{}
	b.WriteString(strconv.Itoa(v.Major))
	b.WriteByte('.')
	b.WriteString(strconv.Itoa(v.Minor))
	b.WriteByte('.')
	b.WriteString(strconv.Itoa(v.Patch))
	return b.String()
}

func (v VersionData) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	b.WriteByte('"')
	b.WriteString(strconv.Itoa(v.Major))
	b.WriteByte('.')
	b.WriteString(strconv.Itoa(v.Minor))
	b.WriteByte('.')
	b.WriteString(strconv.Itoa(v.Patch))
	b.WriteByte('"')
	return b.Bytes(), nil
}

func (v *VersionData) UnmarshalJSON(p []byte) (err error) {
	*v, err = VersionDataFromString(string(p))
	return
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
