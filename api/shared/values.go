package shared

import (
	"github.com/intraware/rodan/internal/sandbox"
)

var SandBoxMap map[int]*sandbox.SandBox = make(map[int]*sandbox.SandBox)

var UserBlackList []int
var TeamBlackList []int
