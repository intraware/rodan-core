package shared

import (
	"github.com/intraware/rodan/internal/sandbox"
	"github.com/intraware/rodan/internal/utils/maps"
)

var SandBoxMap = maps.NewVMap[uint, *sandbox.SandBox]()

var UserBlackList []uint
var TeamBlackList []uint
