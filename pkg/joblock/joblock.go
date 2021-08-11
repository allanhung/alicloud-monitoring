/*
Copyright © 2019 Allan Hung <hung.allan@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package joblock

import (
	"sync"
)

type JobLock struct {
	mtx       sync.Mutex
	IsRunning bool
	Kind      string
}

func (m *JobLock) SetRun(kind string) {
	m.mtx.Lock()
	m.IsRunning = true
	m.Kind = kind
	defer m.mtx.Unlock()
}

func (m *JobLock) DoneRun() {
	m.mtx.Lock()
	m.IsRunning = false
	m.Kind = ""
	defer m.mtx.Unlock()
}
