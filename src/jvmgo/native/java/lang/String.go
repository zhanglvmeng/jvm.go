package lang

import (
    . "jvmgo/any"
    "jvmgo/rtda"
    rtc "jvmgo/rtda/class"
)

func init() {
    _string(intern, "intern", "()Ljava/lang/String;")
}

func _string(method Any, name, desc string) {
    rtc.RegisterNativeMethod("java/lang/String", name, desc, method)
}

// public native String intern();
// ()Ljava/lang/String;
func intern(frame *rtda.Frame) {
    stack := frame.OperandStack()
    str := stack.PopRef() // this
    charArr := str.Class().GetField("value", "[C").GetValue(str).(*rtc.Obj)
    chars := charArr.Fields().([]uint16)
    internedStr := rtc.InternString(chars, str)
    stack.PushRef(internedStr)
}
