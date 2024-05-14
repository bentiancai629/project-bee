package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	s := NewStack(128)
	s.Push(1)
	s.Push(2)
	fmt.Println(s)
	value := s.Pop()
	assert.Equal(t, 2, value)

	assert.Equal(t, value, 2)
}

// func TestVM(t *testing.T) {
// 	data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d, 0x05, 0x0a, 0x0f}
// 	constractState := NewState()
// 	vm := NewVM(data, constractState)
// 	assert.Nil(t, vm.Run())

// 	valueBytes, err := constractState.Get([]byte("FOO"))
// 	value := deserializeInt64(valueBytes)
// 	assert.Nil(t, err)
// 	assert.Equal(t, value, int64(5))
// }
func TestVM2(t *testing.T) {

	// 1 + 2 = 3
	// 1
	// push stack
	// 2
	// push stack
	// add
	// 3
	// push stack

	// push foo to the stack (key)
	// push 3 to the stack
	// push 2 to the stack
	// 3 - 1
	// 1 is on the stack
	// [foo, 1]
	// store
	// data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d} // 0x03, 0x0a, 0x02, 0x0a, 0x0e}
	// F O O => pack[F O O]
	//  data := []byte{ 3, PushInt, "f", PushByte,"o",PushByte,"o",PushByte, Pack }
	// data := []byte{0x02, 0x0a, 0x04, 0x0a, 0x0b, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0d}

	// constractState := NewState()
	// vm := NewVM(data, constractState)
	// assert.Nil(t, vm.Run())

	// fmt.Printf("%+v\n", vm.stack.data)
	// fmt.Printf("%+v", constractState)

	// result := vm.stack.Pop().([]byte)
	// fmt.Printf("%+v", string(result))

	// result := vm.stack.Pop().(int)
	// assert.Equal(t, 1, result)

	// assert.Equal(t, "FOO",string(result))
}

func TestVM3(t *testing.T) {

	// 1 + 2 = 3
	// 1
	// push stack
	// 2
	// push stack
	// add
	// 3
	// push stack
	// data := []byte{0x02, 0x0a, 0x02, 0x0a, 0x0b}
	// constractState := NewState()
	// vm := NewVM(data, constractState)
	// assert.Equal(t, byte(4), vm.Run())
}

func TestVMMul(t *testing.T) {
	data := []byte{0x02, 0x0a, 0x02, 0x0a, 0xea}
	constractState := NewState()
	vm := NewVM(data, constractState)
	assert.Nil(t, vm.Run())

	result := vm.stack.Pop().(int)
	assert.Equal(t, result, 4)
}
