package keepgo

import (
	"errors"
	"strconv"
	"strings"
)

func VersionAtMost(version, max []int) (bool, error) {
	if comp, err := VersionCompare(version, max); err != nil {
		return false, err
	} else if comp == 1 {
		return false, nil
	}
	return true, nil
}
func VersionCompare(v1, v2 []int) (int, error) {
	if len(v1) != len(v2) {
		return 0, errors.New("version length mismatch")
	}

	for idx, v2S := range v2 {
		v1S := v1[idx]
		if v1S > v2S {
			return 1, nil
		}

		if v1S < v2S {
			return -1, nil
		}
	}
	return 0, nil
}
func ParseVersion(v string) []int {
	version := make([]int, 3)

	for idx, vStr := range strings.Split(v, ".") {
		vS, err := strconv.Atoi(vStr)
		if err != nil {
			return nil
		}
		version[idx] = vS
	}

	return version
}
