// Package models provide utility for model

package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mashingan/smapping"
	"github.com/rs/zerolog/log"
)

func CloserErr(fn func() error) {
	if err := fn(); err != nil {
		log.Error().Err(err).Msg("CloserErr")
	}
}

func LogErr(err error) {
	if err != nil {
		log.Error().Err(err).Msg("LogErr")
	}
}

func PanicErr(err error) {
	if err != nil {
		log.Error().Err(err).Msg("fatal")
		panic(err)
	}
}

// Int64ToString ..
func Int64ToString(n int64) string {
	return fmt.Sprint(n)
}

// IntToString ..
func IntToString(n int) string {
	return fmt.Sprint(n)
}

// StringToInt64 ..
func StringToInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}

	return i
}

// StringToInt ..
func StringToInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}

	return val
}

// StringToBool ..
func StringToBool(s string) bool {
	return strings.ToLower(s) == "true"
}

// MapRequestToModel ..
func MapRequestToModel(req interface{}, model interface{}) error {
	return smapping.FillStruct(model, smapping.MapFields(req))
}

// BoolP boolean pointer
func BoolP(b bool) *bool {
	return &b
}

// BoolPString boolean pointer to string
func BoolPString(b *bool) string {
	return BoolString(*b)
}

// BoolString boolean to string
func BoolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// IntP ..
func IntP(i int) *int {
	return &i
}

// IntPString int pointer to string
func IntPString(i *int) string {
	return IntToString(*i)
}

func BoolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func IntBool(i int) bool {
	return i == 1
}

func JSON(i interface{}) []byte {
	bt, _ := json.Marshal(i)
	return bt
}

func JSONS(i interface{}) string {
	return string(JSON(i))
}
