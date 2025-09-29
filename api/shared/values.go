package shared

import (
	"github.com/intraware/rodan/internal/sandbox"
)

var SandBoxMap map[uint]*sandbox.SandBox = make(map[uint]*sandbox.SandBox)

var UserBlackList []uint
var TeamBlackList []uint
