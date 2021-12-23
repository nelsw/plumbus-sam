package util

import "strconv"
import log "github.com/sirupsen/logrus"

func StringToFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		return -400
	}
	return f
}

func StringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		return -400
	}
	return i
}

func StringToInt64(s string) int64 {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		return -400
	}
	i64 := int64(i)
	return i64
}
