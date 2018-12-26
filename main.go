package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/chzyer/readline"
	"github.com/dop251/goja"
	"github.com/logrusorgru/aurora"
)

var vm *goja.Runtime

type Something struct{}

func (s *Something) Add(a, b int) int {
	return a + b
}

func (s *Something) Rev(x string) string {
	a := []byte(x)
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
	return string(a)
}

func MyObject(call goja.ConstructorCall) *goja.Object {
	call.This.Set("method", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(1 + 2)
	})

	s := new(Something)
	v := reflect.ValueOf(s)
	for methi := 0; methi < v.NumMethod(); methi++ {
		methodValue := v.Method(methi)
		method := v.Type().Method(methi)
		call.This.Set(method.Name, func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < method.Type.NumIn()-1 {
				panic(vm.ToValue(fmt.Sprintf(
					"too few arguments to call %s: got %d, expected %d",
					method.Name, len(call.Arguments), method.Type.NumIn()-1)))
			}
			var args []reflect.Value
			for argi := 1; argi < method.Type.NumIn(); argi++ {
				jsarg := call.Arguments[argi-1].Export()
				arg := reflect.ValueOf(jsarg).Convert(method.Type.In(argi))
				args = append(args, arg)
			}
			res := methodValue.Call(args)
			if method.Type.NumOut() == 0 {
				return vm.ToValue(nil)
			} else if method.Type.NumOut() == 1 {
				return vm.ToValue(res[0].Interface())
			}
			return vm.ToValue(res[0].Interface())
		})
	}

	return nil
}

func main() {
	vm = goja.New()
	vm.Set("MyObject", MyObject)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:                 "> ",
		HistoryFile:            "/tmp/readline-multiline",
		DisableAutoSaveHistory: true,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	var lines []string
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if strings.HasSuffix(line, "\\") {
			lines = append(lines, strings.TrimRight(line, "\\"))
			continue
		}
		lines = append(lines, line)
		cmd := strings.Join(lines, "\n")
		lines = nil
		handleCmd(cmd)
		rl.SaveHistory(cmd)
	}
}

func handleCmd(cmd string) {
	fmt.Println(aurora.Cyan(cmd))
	v, err := vm.RunString(cmd)
	if err != nil {
		fmt.Println(aurora.Red(fmt.Sprintf("ERROR: %s", err)))
		return
	}
	fmt.Println(v)
}
