package misc

import (
	"sync/atomic"
	"unsafe"

	"github.com/zxh0/jvm.go/rtda"
	"github.com/zxh0/jvm.go/rtda/heap"
)

func init() {
	_unsafe(compareAndSwapInt, "compareAndSwapInt", "(Ljava/lang/Object;JII)Z")
	_unsafe(compareAndSwapLong, "compareAndSwapLong", "(Ljava/lang/Object;JJJ)Z")
	_unsafe(compareAndSwapObject, "compareAndSwapObject", "(Ljava/lang/Object;JLjava/lang/Object;Ljava/lang/Object;)Z")
}

// public final native boolean compareAndSwapInt(Object o, long offset, int expected, int x);
// (Ljava/lang/Object;JII)Z
func compareAndSwapInt(frame *rtda.Frame) {
	vars := frame.LocalVars()
	fields := vars.GetRef(1).Fields()
	offset := vars.GetLong(2)
	expected := vars.GetInt(4)
	newVal := vars.GetInt(5)

	if slots, ok := fields.([]heap.Slot); ok {
		// object
		swapped := atomic.CompareAndSwapInt64(&(slots[offset].Val), int64(expected), int64(newVal))
		frame.OperandStack().PushBoolean(swapped)
	} else if ints, ok := fields.([]int32); ok {
		// int[]
		swapped := atomic.CompareAndSwapInt32(&ints[offset], expected, newVal)
		frame.OperandStack().PushBoolean(swapped)
	} else {
		// todo
		panic("todo: compareAndSwapInt!")
	}
}

// public final native boolean compareAndSwapLong(Object o, long offset, long expected, long x);
// (Ljava/lang/Object;JJJ)Z
func compareAndSwapLong(frame *rtda.Frame) {
	vars := frame.LocalVars()
	fields := vars.GetRef(1).Fields()
	offset := vars.GetLong(2)
	expected := vars.GetLong(4)
	newVal := vars.GetLong(6)

	if slots, ok := fields.([]heap.Slot); ok {
		// object
		swapped := atomic.CompareAndSwapInt64(&(slots[offset].Val), expected, newVal)
		frame.OperandStack().PushBoolean(swapped)
	} else if ints, ok := fields.([]int64); ok {
		// long[]
		swapped := atomic.CompareAndSwapInt64(&ints[offset], expected, newVal)
		frame.OperandStack().PushBoolean(swapped)
	} else {
		// todo
		panic("todo: compareAndSwapLong!")
	}
}

// public final native boolean compareAndSwapObject(Object o, long offset, Object expected, Object x)
// (Ljava/lang/Object;JLjava/lang/Object;Ljava/lang/Object;)Z
func compareAndSwapObject(frame *rtda.Frame) {
	vars := frame.LocalVars()
	obj := vars.GetRef(1)
	fields := obj.Fields()
	offset := vars.GetLong(2)
	expected := vars.GetRef(4)
	newVal := vars.GetRef(5)

	// todo
	if slots, ok := fields.([]heap.Slot); ok {
		// object
		swapped := _casObj(obj, slots, offset, expected, newVal)
		frame.OperandStack().PushBoolean(swapped)
	} else if objs, ok := fields.([]*heap.Object); ok {
		// ref[]
		swapped := _casArr(objs, offset, expected, newVal)
		frame.OperandStack().PushBoolean(swapped)
	} else {
		// todo
		panic("todo: compareAndSwapObject!")
	}
}
func _casObj(obj *heap.Object, fields []heap.Slot, offset int64, expected, newVal *heap.Object) bool {
	// todo
	obj.LockState()
	defer obj.UnlockState()

	current := fields[offset].Ref
	if current == expected {
		fields[offset].Ref = newVal
		return true
	} else {
		return false
	}
}
func _casArr(objs []*heap.Object, offset int64, expected, newVal *heap.Object) bool {
	// cast to []unsafe.Pointer
	ps := *((*[]unsafe.Pointer)(unsafe.Pointer(&objs)))

	addr := &ps[offset]
	old := unsafe.Pointer(expected)
	_new := unsafe.Pointer(newVal)
	return atomic.CompareAndSwapPointer(addr, old, _new)
}
