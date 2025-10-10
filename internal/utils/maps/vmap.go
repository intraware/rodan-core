// Copyright (c) 2023 WarpDL
// Licensed under the MIT License
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://opensource.org/licenses/MIT
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package maps

import (
	"sync"
)

type VMap[kT comparable, vT any] struct {
	kv map[kT]vT
	mu sync.RWMutex
}

func NewVMap[kT comparable, vT any]() VMap[kT, vT] {
	return VMap[kT, vT]{
		kv: make(map[kT]vT),
	}
}

func (vm *VMap[kT, vT]) Make() {
	vm.kv = make(map[kT]vT)
}

func (vm *VMap[kT, vT]) Set(key kT, val vT) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.kv[key] = val
}

func (vm *VMap[kT, vT]) GetUnsafe(key kT) (val vT, ok bool) {
	val, ok = vm.kv[key]
	return
}

func (vm *VMap[kT, vT]) Get(key kT) (val vT, ok bool) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	val, ok = vm.GetUnsafe(key)
	return
}

func (vm *VMap[kT, vT]) DumpValues() (vals []vT) {
	n := len(vm.kv)
	vals = make([]vT, n)

	vm.mu.Lock()
	defer vm.mu.Unlock()

	var i int
	for _, val := range vm.kv {
		vals[i] = val
		i++
	}
	return
}

func (vm *VMap[kT, vT]) Dump() (keys []kT, vals []vT) {
	n := len(vm.kv)

	keys = make([]kT, n)
	vals = make([]vT, n)

	vm.mu.Lock()
	defer vm.mu.Unlock()

	var i int
	for key, val := range vm.kv {
		keys[i] = key
		vals[i] = val
		i++
	}
	return
}
