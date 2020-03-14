/*
 * Copyright 2020 Tero Vierimaa
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package task

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
	"testing"
	"time"
	"tryffel.net/go/jellycli/util"
)

func TestTask_recoverPanic(t *testing.T) {
	// test recovering from panics and printing traces
	loop := func() {
		time.Sleep(time.Millisecond)
		panic("generic error")
	}

	wantExitMsg := "Task 'test' panic: generic error\n"

	exitMsg := ""
	exitData := map[string]interface{}{}

	exit := func(instance *logrus.Entry, msg string) {
		exitData = instance.Data
		exitMsg = msg
	}

	// do not call logrus.fatal but read message we are getting
	util.Exit = exit

	task := &Task{
		Name:        "test",
		initialized: true,
	}
	task.init()
	task.SetLoop(loop)

	output := &bytes.Buffer{}
	logrus.SetOutput(output)
	logrus.SetLevel(logrus.WarnLevel)

	err := task.Start()
	if err != nil {
		t.Errorf("start task: %v", err)
	}

	// wait for panic
	time.Sleep(time.Millisecond * 20)

	/*
		StackTrace should look like:
		goroutine 7 [running]:
		tryffel.net/go/jellycli/task.TestTask_recoverPanic.func1()
		        x/jellycli/task/task_test.go:30 +0x46
		tryffel.net/go/jellycli/task.(*Task).run(0xc000010880)
		        x/jellycli/task/task.go:110 +0x7b
		created by tryffel.net/go/jellycli/task.(*Task).Start
		        x/jellycli/task/task.go:86 +0x247
	*/

	if exitMsg != wantExitMsg {
		t.Errorf("Exit msg: got %s, expected %s", exitMsg, wantExitMsg)
	}

	s := exitData["Stacktrace"]
	stackTrace, ok := s.(string)
	if !ok {
		t.Errorf("Stacktrace not string")
		return
	}

	lines := strings.Split(stackTrace, "\n")

	if len(lines) != 8 {
		t.Errorf("expect 7 ")
	}

	// goroutine num
	if match, _ := regexp.Match(`goroutine\s\d+\s\[running\]:`, []byte(lines[0])); !match {
		t.Errorf("expect stacktrace 1st line goroutine x")
	}

	// package name and funcion
	packages := [][]byte{[]byte(lines[1]), []byte(lines[3]), []byte(lines[5])}
	for _, v := range packages {
		if match, _ := regexp.Match(`(tryffel.net/go/jellycli/task)`, v); !match {
			t.Errorf("package name and func not showing: %v", v)
		}
	}

	// filenames
	files := [][]byte{[]byte(lines[2]), []byte(lines[4]), []byte(lines[6])}
	for _, v := range files {
		if match, _ := regexp.Match(`task.+\.go:\d+`, v); !match {
			t.Errorf("file name not showing: %v", v)
		}
	}

}
