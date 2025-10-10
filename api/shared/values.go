package shared

import (
	"sync/atomic"

	"github.com/intraware/rodan/internal/sandbox"
	"github.com/intraware/rodan/internal/utils/maps"
)

var SandBoxMap = maps.NewVMap[uint, []*sandbox.SandBox]()

var UserBlackList []uint
var TeamBlackList []uint

var allowSubmissions atomic.Bool

func SetSubmissions(allowSub bool) {
	allowSubmissions.Store(allowSub)
}

func GetSubmissions() bool {
	return allowSubmissions.Load()
}
