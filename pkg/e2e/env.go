package e2e

import (
	"os"
	"strconv"
	"strings"
)

func envVar(name, defaultVal string) string {
	val := os.Getenv(name)
	if val == "" {
		return defaultVal
	}
	return val
}

func envVarStrings(name string) []string {
	val := os.Getenv(name)
	vals := strings.Split(val, ",")
	retVals := make([]string, 0, len(vals))
	for _, v := range vals {
		if v == "" {
			continue
		}
		retVals = append(retVals, v)
	}
	return retVals
}

func envVarBool(name string) bool {
	valStr := os.Getenv(name)
	return valStr != ""
}

func envVarInt(name string, defaultVal int) int {
	valStr := os.Getenv(name)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}
