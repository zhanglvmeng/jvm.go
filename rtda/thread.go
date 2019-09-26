package rtda

import (
	"fmt"
	"strings"
	"sync"

	"github.com/zxh0/jvm.go/options"
	"github.com/zxh0/jvm.go/rtda/heap"
)

/*
JVM
  Thread
    pc
    Stack
      Frame
        LocalVars
        OperandStack
*/
type Thread struct {
	pc              int // the address of the instruction currently being executed
	stack           *Stack
	frameCache      *FrameCache
	jThread         *heap.Object // java.lang.Thread
	lock            *sync.Mutex  // state lock
	ch              chan int
	sleepingFlag    bool
	interruptedFlag bool
	parkingFlag     bool // used by Unsafe
	unparkedFlag    bool // used by Unsafe
	// todo
}

func NewThread(jThread *heap.Object) *Thread {
	stack := newStack(options.ThreadStackSize)
	thread := &Thread{
		stack:   stack,
		jThread: jThread,
		lock:    &sync.Mutex{},
		ch:      make(chan int),
	}
	thread.frameCache = newFrameCache(thread, 16) // todo
	return thread
}

// getters & setters
func (thread *Thread) PC() int {
	return thread.pc
}
func (thread *Thread) SetPC(pc int) {
	thread.pc = pc
}
func (thread *Thread) JThread() *heap.Object {
	return thread.jThread
}

func (thread *Thread) IsStackEmpty() bool {
	return thread.stack.isEmpty()
}
func (thread *Thread) StackDepth() uint {
	return thread.stack.size
}

func (thread *Thread) CurrentFrame() *Frame {
	return thread.stack.top()
}
func (thread *Thread) TopFrame() *Frame {
	return thread.stack.top()
}
func (thread *Thread) TopFrameN(n uint) *Frame {
	return thread.stack.topN(n)
}

func (thread *Thread) PushFrame(frame *Frame) {
	thread.stack.push(frame)
}
func (thread *Thread) PopFrame() *Frame {
	top := thread.stack.pop()
	if top.onPopAction != nil {
		// todo
		top.onPopAction()
	}

	thread.frameCache.returnFrame(top)
	return top
}

func (thread *Thread) NewFrame(method *heap.Method) *Frame {
	if method.IsNative() {
		return newNativeFrame(thread, method)
	} else {
		return thread.frameCache.borrowFrame(method)
		//return newFrame(thread, method)
	}
}

func (thread *Thread) InvokeMethod(method *heap.Method) {
	//thread._logInvoke(thread.stack.size, method)
	currentFrame := thread.CurrentFrame()
	newFrame := thread.NewFrame(method)
	thread.PushFrame(newFrame)
	argSlotsCount := method.ArgSlotCount()
	if argSlotsCount > 0 {
		_passArgs(currentFrame.operandStack, newFrame.localVars, argSlotsCount)
	}

	if method.IsSynchronized() {
		var monitor *heap.Monitor
		if method.IsStatic() {
			classObj := method.Class().JClass()
			monitor = classObj.Monitor()
		} else {
			thisObj := newFrame.LocalVars().GetThis()
			monitor = thisObj.Monitor()
		}

		monitor.Enter(thread)
		newFrame.SetOnPopAction(func() {
			monitor.Exit(thread)
		})
	}
}
func _passArgs(stack *OperandStack, vars *LocalVars, argSlotsCount uint) {
	args := stack.PopTops(argSlotsCount)
	for i := uint(0); i < argSlotsCount; i++ {
		arg := args[i]
		args[i] = EmptySlot
		vars.Set(i, arg)
	}
}
func (thread *Thread) _logInvoke(stackSize uint, method *heap.Method) {
	space := strings.Repeat(" ", int(stackSize))
	className := method.Class().Name()
	methodName := method.Name()

	if method.IsStatic() {
		fmt.Printf("[method]%v thread:%p %v.%v()\n", space, thread, className, methodName)
	} else {
		fmt.Printf("[method]%v thread:%p %v#%v()\n", space, thread, className, methodName)
	}
}

func (thread *Thread) InvokeMethodWithShim(method *heap.Method, args []Slot) {
	shimFrame := newShimFrame(thread, args)
	thread.PushFrame(shimFrame)
	thread.InvokeMethod(method)
}

func (thread *Thread) InitClass(class *heap.Class) {
	initClass(thread, class)
}

func (thread *Thread) HandleUncaughtException(ex *heap.Object) {
	thread.stack.clear()
	sysClass := heap.BootLoader().LoadClass("java/lang/System")
	sysErr := sysClass.GetStaticValue("out", "Ljava/io/PrintStream;").Ref
	printStackTrace := ex.Class().GetInstanceMethod("printStackTrace", "(Ljava/io/PrintStream;)V")

	// call ex.printStackTrace(System.err)
	newFrame := thread.NewFrame(printStackTrace)
	vars := newFrame.localVars
	vars.SetRef(0, ex)
	vars.SetRef(1, sysErr)
	thread.PushFrame(newFrame)

	//
	// printString := sysErr.Class().GetInstanceMethod("print", "(Ljava/lang/String;)V")
	// newFrame = thread.NewFrame(printString)
	// vars = newFrame.localVars
	// vars.SetRef(0, sysErr)
	// vars.SetRef(1, JString("Exception in thread \"main\" ", newFrame))
	// thread.PushFrame(newFrame)
}

// hack
func (thread *Thread) HackSetJThread(jThread *heap.Object) {
	thread.jThread = jThread
}
